package txprocessor

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
)

// TxProcessor represents the TxProcessor object
type TxProcessor struct {
	s   *statedb.StateDB
	zki *common.ZKInputs
	// i is the current transaction index in the ZKInputs generation (zki)
	i int
	// AccumulatedFees contains the accumulated fees for each token (Coord
	// Idx) in the processed batch
	AccumulatedFees map[common.Idx]*big.Int
	config          Config
}

// Config contains the TxProcessor configuration parameters
type Config struct {
	NLevels  uint32
	MaxFeeTx uint32
	MaxTx    uint32
	MaxL1Tx  uint32
	ChainID  uint16
}

type processedExit struct {
	exit    bool
	newExit bool
	idx     common.Idx
	acc     common.Account
}

// ProcessTxOutput contains the output of the ProcessTxs method
type ProcessTxOutput struct {
	ZKInputs           *common.ZKInputs
	ExitInfos          []common.ExitInfo
	CreatedAccounts    []common.Account
	CoordinatorIdxsMap map[common.TokenID]common.Idx
	CollectedFees      map[common.TokenID]*big.Int
}

func newErrorNotEnoughBalance(tx common.Tx) error {
	var msg error
	if tx.IsL1 {
		msg = fmt.Errorf("Invalid transaction, not enough balance on sender account. TxID: %s, TxType: %s, FromIdx: %d, ToIdx: %d, Amount: %d",
			tx.TxID, tx.Type, tx.FromIdx, tx.ToIdx, tx.Amount)
	} else {
		msg = fmt.Errorf("Invalid transaction, not enough balance on sender account. TxID: %s, TxType: %s, FromIdx: %d, ToIdx: %d, Amount: %d, Fee: %d",
			tx.TxID, tx.Type, tx.FromIdx, tx.ToIdx, tx.Amount, tx.Fee)
	}
	return tracerr.Wrap(msg)
}

// NewTxProcessor returns a new TxProcessor with the given *StateDB & Config
func NewTxProcessor(sdb *statedb.StateDB, config Config) *TxProcessor {
	return &TxProcessor{
		s:      sdb,
		zki:    nil,
		i:      0,
		config: config,
	}
}

// StateDB returns a pointer to the StateDB of the TxProcessor
func (tp *TxProcessor) StateDB() *statedb.StateDB {
	return tp.s
}

func (tp *TxProcessor) resetZKInputs() {
	tp.zki = nil
	tp.i = 0 // initialize current transaction index in the ZKInputs generation
}

// ProcessTxs process the given L1Txs & L2Txs applying the needed updates to
// the StateDB depending on the transaction Type.  If StateDB
// type==TypeBatchBuilder, returns the common.ZKInputs to generate the
// SnarkProof later used by the BatchBuilder.  If StateDB
// type==TypeSynchronizer, assumes that the call is done from the Synchronizer,
// returns common.ExitTreeLeaf that is later used by the Synchronizer to update
// the HistoryDB, and adds Nonce & TokenID to the L2Txs.
// And if TypeSynchronizer returns an array of common.Account with all the
// created accounts.
func (tp *TxProcessor) ProcessTxs(coordIdxs []common.Idx, l1usertxs, l1coordinatortxs []common.L1Tx,
	l2txs []common.PoolL2Tx) (ptOut *ProcessTxOutput, err error) {
	defer func() {
		if err == nil {
			err = tp.s.MakeCheckpoint()
		}
	}()

	var exitTree *merkletree.MerkleTree
	var createdAccounts []common.Account

	if tp.zki != nil {
		return nil, tracerr.Wrap(errors.New("Expected StateDB.zki==nil, something went wrong and it's not empty"))
	}
	defer tp.resetZKInputs()

	if len(coordIdxs) >= int(tp.config.MaxFeeTx) {
		return nil, tracerr.Wrap(fmt.Errorf("CoordIdxs (%d) length must be smaller than MaxFeeTx (%d)", len(coordIdxs), tp.config.MaxFeeTx))
	}

	nTx := len(l1usertxs) + len(l1coordinatortxs) + len(l2txs)

	if nTx > int(tp.config.MaxTx) {
		return nil, tracerr.Wrap(fmt.Errorf("L1UserTx + L1CoordinatorTx + L2Tx (%d) can not be bigger than MaxTx (%d)", nTx, tp.config.MaxTx))
	}
	if len(l1usertxs)+len(l1coordinatortxs) > int(tp.config.MaxL1Tx) {
		return nil, tracerr.Wrap(fmt.Errorf("L1UserTx + L1CoordinatorTx (%d) can not be bigger than MaxL1Tx (%d)", len(l1usertxs)+len(l1coordinatortxs), tp.config.MaxTx))
	}

	exits := make([]processedExit, nTx)

	if tp.s.Type() == statedb.TypeBatchBuilder {
		tp.zki = common.NewZKInputs(tp.config.ChainID, tp.config.MaxTx, tp.config.MaxL1Tx,
			tp.config.MaxFeeTx, tp.config.NLevels, (tp.s.CurrentBatch() + 1).BigInt())
		tp.zki.OldLastIdx = tp.s.CurrentIdx().BigInt()
		tp.zki.OldStateRoot = tp.s.MT.Root().BigInt()
		tp.zki.Metadata.NewLastIdxRaw = tp.s.CurrentIdx()
	}

	// TBD if ExitTree is only in memory or stored in disk, for the moment
	// is only needed in memory
	if tp.s.Type() == statedb.TypeSynchronizer || tp.s.Type() == statedb.TypeBatchBuilder {
		tmpDir, err := ioutil.TempDir("", "hermez-statedb-exittree")
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				log.Errorw("Deleting statedb temp exit tree", "err", err)
			}
		}()
		sto, err := pebble.NewPebbleStorage(tmpDir, false)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		defer sto.Close()
		exitTree, err = merkletree.NewMerkleTree(sto, tp.s.MT.MaxLevels())
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}

	// Process L1UserTxs
	for i := 0; i < len(l1usertxs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		exitIdx, exitAccount, newExit, createdAccount, err := tp.ProcessL1Tx(exitTree,
			&l1usertxs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if tp.s.Type() == statedb.TypeSynchronizer {
			if createdAccount != nil {
				createdAccounts = append(createdAccounts, *createdAccount)
				l1usertxs[i].EffectiveFromIdx = createdAccount.Idx
			} else {
				l1usertxs[i].EffectiveFromIdx = l1usertxs[i].FromIdx
			}
		}
		if tp.zki != nil {
			l1TxData, err := l1usertxs[i].BytesGeneric()
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tp.zki.Metadata.L1TxsData = append(tp.zki.Metadata.L1TxsData, l1TxData)

			l1TxDataAvailability, err :=
				l1usertxs[i].BytesDataAvailability(tp.zki.Metadata.NLevels)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tp.zki.Metadata.L1TxsDataAvailability =
				append(tp.zki.Metadata.L1TxsDataAvailability, l1TxDataAvailability)

			tp.zki.ISOutIdx[tp.i] = tp.s.CurrentIdx().BigInt()
			tp.zki.ISStateRoot[tp.i] = tp.s.MT.Root().BigInt()
			if exitIdx == nil {
				tp.zki.ISExitRoot[tp.i] = exitTree.Root().BigInt()
			}
		}
		if tp.s.Type() == statedb.TypeSynchronizer || tp.s.Type() == statedb.TypeBatchBuilder {
			if exitIdx != nil && exitTree != nil {
				exits[tp.i] = processedExit{
					exit:    true,
					newExit: newExit,
					idx:     *exitIdx,
					acc:     *exitAccount,
				}
			}
			tp.i++
		}
	}

	// Process L1CoordinatorTxs
	for i := 0; i < len(l1coordinatortxs); i++ {
		exitIdx, _, _, createdAccount, err := tp.ProcessL1Tx(exitTree, &l1coordinatortxs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if exitIdx != nil {
			log.Error("Unexpected Exit in L1CoordinatorTx")
		}
		if tp.s.Type() == statedb.TypeSynchronizer {
			if createdAccount != nil {
				createdAccounts = append(createdAccounts, *createdAccount)
				l1coordinatortxs[i].EffectiveFromIdx = createdAccount.Idx
			} else {
				l1coordinatortxs[i].EffectiveFromIdx = l1coordinatortxs[i].FromIdx
			}
		}
		if tp.zki != nil {
			l1TxData, err := l1coordinatortxs[i].BytesGeneric()
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tp.zki.Metadata.L1TxsData = append(tp.zki.Metadata.L1TxsData, l1TxData)
			l1TxDataAvailability, err :=
				l1coordinatortxs[i].BytesDataAvailability(tp.zki.Metadata.NLevels)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tp.zki.Metadata.L1TxsDataAvailability =
				append(tp.zki.Metadata.L1TxsDataAvailability, l1TxDataAvailability)

			tp.zki.ISOutIdx[tp.i] = tp.s.CurrentIdx().BigInt()
			tp.zki.ISStateRoot[tp.i] = tp.s.MT.Root().BigInt()
			tp.i++
		}
	}

	// remove repeated CoordIdxs that are for the same TokenID (use the
	// first occurrence)
	usedCoordTokenIDs := make(map[common.TokenID]bool)
	var filteredCoordIdxs []common.Idx
	for i := 0; i < len(coordIdxs); i++ {
		accCoord, err := tp.s.GetAccount(coordIdxs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if !usedCoordTokenIDs[accCoord.TokenID] {
			usedCoordTokenIDs[accCoord.TokenID] = true
			filteredCoordIdxs = append(filteredCoordIdxs, coordIdxs[i])
		}
	}
	coordIdxs = filteredCoordIdxs

	tp.AccumulatedFees = make(map[common.Idx]*big.Int)
	for _, idx := range coordIdxs {
		tp.AccumulatedFees[idx] = big.NewInt(0)
	}

	// once L1UserTxs & L1CoordinatorTxs are processed, get TokenIDs of
	// coordIdxs. In this way, if a coordIdx uses an Idx that is being
	// created in the current batch, at this point the Idx will be created
	coordIdxsMap, err := tp.s.GetTokenIDsFromIdxs(coordIdxs)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	// collectedFees will contain the amount of fee collected for each
	// TokenID
	var collectedFees map[common.TokenID]*big.Int
	if tp.s.Type() == statedb.TypeSynchronizer || tp.s.Type() == statedb.TypeBatchBuilder {
		collectedFees = make(map[common.TokenID]*big.Int)
		for tokenID := range coordIdxsMap {
			collectedFees[tokenID] = big.NewInt(0)
		}
	}

	if tp.zki != nil {
		// get the feePlanTokens
		feePlanTokens, err := tp.getFeePlanTokens(coordIdxs)
		if err != nil {
			log.Error(err)
			return nil, tracerr.Wrap(err)
		}
		copy(tp.zki.FeePlanTokens, feePlanTokens)
	}

	// Process L2Txs
	for i := 0; i < len(l2txs); i++ {
		exitIdx, exitAccount, newExit, err := tp.ProcessL2Tx(coordIdxsMap, collectedFees,
			exitTree, &l2txs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if tp.zki != nil {
			l2TxData, err := l2txs[i].L2Tx().BytesDataAvailability(tp.zki.Metadata.NLevels)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tp.zki.Metadata.L2TxsData = append(tp.zki.Metadata.L2TxsData, l2TxData)

			// Intermediate States
			if tp.i < nTx-1 {
				tp.zki.ISOutIdx[tp.i] = tp.s.CurrentIdx().BigInt()
				tp.zki.ISStateRoot[tp.i] = tp.s.MT.Root().BigInt()
				tp.zki.ISAccFeeOut[tp.i] = formatAccumulatedFees(collectedFees, tp.zki.FeePlanTokens, coordIdxs)
				if exitIdx == nil {
					tp.zki.ISExitRoot[tp.i] = exitTree.Root().BigInt()
				}
			}
		}
		if tp.s.Type() == statedb.TypeSynchronizer || tp.s.Type() == statedb.TypeBatchBuilder {
			if exitIdx != nil && exitTree != nil {
				exits[tp.i] = processedExit{
					exit:    true,
					newExit: newExit,
					idx:     *exitIdx,
					acc:     *exitAccount,
				}
			}
			tp.i++
		}
	}

	if tp.zki != nil {
		txCompressedDataEmpty := common.TxCompressedDataEmpty(tp.config.ChainID)
		last := tp.i - 1
		if tp.i == 0 {
			last = 0
		}
		for i := last; i < int(tp.config.MaxTx); i++ {
			if i < int(tp.config.MaxTx)-1 {
				tp.zki.ISOutIdx[i] = tp.s.CurrentIdx().BigInt()
				tp.zki.ISStateRoot[i] = tp.s.MT.Root().BigInt()
				tp.zki.ISAccFeeOut[i] = formatAccumulatedFees(collectedFees,
					tp.zki.FeePlanTokens, coordIdxs)
				tp.zki.ISExitRoot[i] = exitTree.Root().BigInt()
			}
			if i >= tp.i {
				tp.zki.TxCompressedData[i] = txCompressedDataEmpty
			}
		}
		isFinalAccFee := formatAccumulatedFees(collectedFees, tp.zki.FeePlanTokens, coordIdxs)
		copy(tp.zki.ISFinalAccFee, isFinalAccFee)
		// before computing the Fees txs, set the ISInitStateRootFee
		tp.zki.ISInitStateRootFee = tp.s.MT.Root().BigInt()
	}

	// distribute the AccumulatedFees from the processed L2Txs into the
	// Coordinator Idxs
	iFee := 0
	for _, idx := range coordIdxs {
		accumulatedFee := tp.AccumulatedFees[idx]

		// send the fee to the Idx of the Coordinator for the TokenID
		// (even if the AccumulatedFee==0, as is how the zk circuit
		// works)
		accCoord, err := tp.s.GetAccount(idx)
		if err != nil {
			log.Errorw("Can not distribute accumulated fees to coordinator account: No coord Idx to receive fee", "idx", idx)
			return nil, tracerr.Wrap(err)
		}
		if tp.zki != nil {
			tp.zki.TokenID3[iFee] = accCoord.TokenID.BigInt()
			tp.zki.Nonce3[iFee] = accCoord.Nonce.BigInt()
			coordBJJSign, coordBJJY := babyjub.UnpackSignY(accCoord.BJJ)
			if coordBJJSign {
				tp.zki.Sign3[iFee] = big.NewInt(1)
			}
			tp.zki.Ay3[iFee] = coordBJJY
			tp.zki.Balance3[iFee] = accCoord.Balance
			tp.zki.EthAddr3[iFee] = common.EthAddrToBigInt(accCoord.EthAddr)
		}
		accCoord.Balance = new(big.Int).Add(accCoord.Balance, accumulatedFee)
		pFee, err := tp.s.UpdateAccount(idx, accCoord)
		if err != nil {
			log.Error(err)
			return nil, tracerr.Wrap(err)
		}
		if tp.zki != nil {
			tp.zki.Siblings3[iFee] = siblingsToZKInputFormat(pFee.Siblings)
			tp.zki.ISStateRootFee[iFee] = tp.s.MT.Root().BigInt()
		}
		iFee++
	}
	if tp.zki != nil {
		for i := len(tp.AccumulatedFees); i < int(tp.config.MaxFeeTx)-1; i++ {
			tp.zki.ISStateRootFee[i] = tp.s.MT.Root().BigInt()
		}
		// add Coord Idx to ZKInputs.FeeTxsData
		for i := 0; i < len(coordIdxs); i++ {
			tp.zki.FeeIdxs[i] = coordIdxs[i].BigInt()
		}
	}

	if tp.s.Type() == statedb.TypeTxSelector {
		return nil, nil
	}

	// once all txs processed (exitTree root frozen), for each Exit,
	// generate common.ExitInfo data
	var exitInfos []common.ExitInfo
	exitInfosByIdx := make(map[common.Idx]*common.ExitInfo)
	for i := 0; i < nTx; i++ {
		if !exits[i].exit {
			continue
		}
		exitIdx := exits[i].idx
		exitAccount := exits[i].acc

		// 0. generate MerkleProof
		p, err := exitTree.GenerateSCVerifierProof(exitIdx.BigInt(), nil)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		// 1. generate common.ExitInfo
		ei := common.ExitInfo{
			AccountIdx:  exitIdx,
			MerkleProof: p,
			Balance:     exitAccount.Balance,
		}
		if prevExit, ok := exitInfosByIdx[exitIdx]; !ok {
			exitInfos = append(exitInfos, ei)
			exitInfosByIdx[exitIdx] = &exitInfos[len(exitInfos)-1]
		} else {
			*prevExit = ei
		}
	}

	if tp.s.Type() == statedb.TypeSynchronizer {
		// retuTypeexitInfos, createdAccounts and collectedFees, so Synchronizer will
		// be able to store it into HistoryDB for the concrete BatchNum
		return &ProcessTxOutput{
			ZKInputs:           nil,
			ExitInfos:          exitInfos,
			CreatedAccounts:    createdAccounts,
			CoordinatorIdxsMap: coordIdxsMap,
			CollectedFees:      collectedFees,
		}, nil
	}

	// compute last ZKInputs parameters
	tp.zki.GlobalChainID = big.NewInt(int64(tp.config.ChainID))
	tp.zki.Metadata.NewStateRootRaw = tp.s.MT.Root()
	tp.zki.Metadata.NewExitRootRaw = exitTree.Root()

	// return ZKInputs as the BatchBuilder will return it to forge the Batch
	return &ProcessTxOutput{
		ZKInputs:           tp.zki,
		ExitInfos:          nil,
		CreatedAccounts:    nil,
		CoordinatorIdxsMap: coordIdxsMap,
		CollectedFees:      nil,
	}, nil
}

// getFeePlanTokens returns an array of *big.Int containing a list of tokenIDs
// corresponding to the given CoordIdxs and the processed L2Txs
func (tp *TxProcessor) getFeePlanTokens(coordIdxs []common.Idx) ([]*big.Int, error) {
	var tBI []*big.Int
	for i := 0; i < len(coordIdxs); i++ {
		acc, err := tp.s.GetAccount(coordIdxs[i])
		if err != nil {
			log.Errorf("could not get account to determine TokenID of CoordIdx %d not found: %s", coordIdxs[i], err.Error())
			return nil, tracerr.Wrap(err)
		}
		tBI = append(tBI, acc.TokenID.BigInt())
	}
	return tBI, nil
}

// ProcessL1Tx process the given L1Tx applying the needed updates to the
// StateDB depending on the transaction Type. It returns the 3 parameters
// related to the Exit (in case of): Idx, ExitAccount, boolean determining if
// the Exit created a new Leaf in the ExitTree.
// And another *common.Account parameter which contains the created account in
// case that has been a new created account and that the StateDB is of type
// TypeSynchronizer.
func (tp *TxProcessor) ProcessL1Tx(exitTree *merkletree.MerkleTree, tx *common.L1Tx) (*common.Idx,
	*common.Account, bool, *common.Account, error) {
	// ZKInputs
	if tp.zki != nil {
		// Txs
		var err error
		tp.zki.TxCompressedData[tp.i], err = tx.TxCompressedData(tp.config.ChainID)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		tp.zki.FromIdx[tp.i] = tx.FromIdx.BigInt()
		tp.zki.ToIdx[tp.i] = tx.ToIdx.BigInt()
		tp.zki.OnChain[tp.i] = big.NewInt(1)

		// L1Txs
		depositAmountF40, err := common.NewFloat40(tx.DepositAmount)
		if err != nil {
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		tp.zki.DepositAmountF[tp.i] = big.NewInt(int64(depositAmountF40))
		tp.zki.FromEthAddr[tp.i] = common.EthAddrToBigInt(tx.FromEthAddr)
		if tx.FromBJJ != common.EmptyBJJComp {
			tp.zki.FromBJJCompressed[tp.i] = BJJCompressedTo256BigInts(tx.FromBJJ)
		}

		// Intermediate States, for all the transactions except for the last one
		if tp.i < len(tp.zki.ISOnChain) { // len(tp.zki.ISOnChain) == nTx
			tp.zki.ISOnChain[tp.i] = big.NewInt(1)
		}
	}

	switch tx.Type {
	case common.TxTypeForceTransfer:
		tp.computeEffectiveAmounts(tx)

		// go to the MT account of sender and receiver, and update balance
		// & nonce

		// coordIdxsMap is 'nil', as at L1Txs there is no L2 fees.
		// 0 for the parameter toIdx, as at L1Tx ToIdx can only be 0 in
		// the Deposit type case.
		err := tp.applyTransfer(nil, nil, tx.Tx(), 0)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeCreateAccountDeposit:
		tp.computeEffectiveAmounts(tx)

		// add new account to the MT, update balance of the MT account
		err := tp.applyCreateAccount(tx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		// TODO applyCreateAccount will return the created account,
		// which in the case type==TypeSynchronizer will be added to an
		// array of created accounts that will be returned
	case common.TxTypeDeposit:
		tp.computeEffectiveAmounts(tx)

		// update balance of the MT account
		err := tp.applyDeposit(tx, false)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeDepositTransfer:
		tp.computeEffectiveAmounts(tx)

		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := tp.applyDeposit(tx, true)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeCreateAccountDepositTransfer:
		tp.computeEffectiveAmounts(tx)

		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := tp.applyCreateAccountDepositTransfer(tx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeForceExit:
		tp.computeEffectiveAmounts(tx)

		// execute exit flow
		// coordIdxsMap is 'nil', as at L1Txs there is no L2 fees
		exitAccount, newExit, err := tp.applyExit(nil, nil, exitTree, tx.Tx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		return &tx.FromIdx, exitAccount, newExit, nil, nil
	default:
	}

	var createdAccount *common.Account
	if tp.s.Type() == statedb.TypeSynchronizer &&
		(tx.Type == common.TxTypeCreateAccountDeposit ||
			tx.Type == common.TxTypeCreateAccountDepositTransfer) {
		var err error
		createdAccount, err = tp.s.GetAccount(tp.s.CurrentIdx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	}

	return nil, nil, false, createdAccount, nil
}

// ProcessL2Tx process the given L2Tx applying the needed updates to the
// StateDB depending on the transaction Type. It returns the 3 parameters
// related to the Exit (in case of): Idx, ExitAccount, boolean determining if
// the Exit created a new Leaf in the ExitTree.
func (tp *TxProcessor) ProcessL2Tx(coordIdxsMap map[common.TokenID]common.Idx,
	collectedFees map[common.TokenID]*big.Int, exitTree *merkletree.MerkleTree,
	tx *common.PoolL2Tx) (*common.Idx, *common.Account, bool, error) {
	var err error
	// if tx.ToIdx==0, get toIdx by ToEthAddr or ToBJJ
	if tx.ToIdx == common.Idx(0) && tx.AuxToIdx == common.Idx(0) {
		if tp.s.Type() == statedb.TypeSynchronizer {
			// thisTypeould never be reached
			log.Error("WARNING: In StateDB with Synchronizer mode L2.ToIdx can't be 0")
			return nil, nil, false, tracerr.Wrap(fmt.Errorf("In StateDB with Synchronizer mode L2.ToIdx can't be 0"))
		}
		// case when tx.Type== common.TxTypeTransferToEthAddr or common.TxTypeTransferToBJJ

		accSender, err := tp.s.GetAccount(tx.FromIdx)
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
		tx.AuxToIdx, err = tp.s.GetIdxByEthAddrBJJ(tx.ToEthAddr, tx.ToBJJ, accSender.TokenID)
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
	}

	// ZKInputs
	if tp.zki != nil {
		// Txs
		tp.zki.TxCompressedData[tp.i], err = tx.TxCompressedData(tp.config.ChainID)
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
		tp.zki.TxCompressedDataV2[tp.i], err = tx.TxCompressedDataV2()
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
		tp.zki.FromIdx[tp.i] = tx.FromIdx.BigInt()
		tp.zki.ToIdx[tp.i] = tx.ToIdx.BigInt()

		// fill AuxToIdx if needed
		if tx.ToIdx == 0 {
			// use toIdx that can have been filled by tx.ToIdx or
			// if tx.Idx==0 (this case), toIdx is filled by the Idx
			// from db by ToEthAddr&ToBJJ
			tp.zki.AuxToIdx[tp.i] = tx.AuxToIdx.BigInt()
		}

		if tx.ToBJJ != common.EmptyBJJComp {
			_, tp.zki.ToBJJAy[tp.i] = babyjub.UnpackSignY(tx.ToBJJ)
		}
		tp.zki.ToEthAddr[tp.i] = common.EthAddrToBigInt(tx.ToEthAddr)

		tp.zki.OnChain[tp.i] = big.NewInt(0)
		tp.zki.NewAccount[tp.i] = big.NewInt(0)

		// L2Txs
		// tp.zki.RqOffset[tp.i] =  // TODO Rq once TxSelector is ready
		// tp.zki.RqTxCompressedDataV2[tp.i] = // TODO
		// tp.zki.RqToEthAddr[tp.i] = common.EthAddrToBigInt(tx.RqToEthAddr) // TODO
		// tp.zki.RqToBJJAy[tp.i] = tx.ToBJJ.Y // TODO

		signature, err := tx.Signature.Decompress()
		if err != nil {
			log.Error(err)
			return nil, nil, false, tracerr.Wrap(err)
		}
		tp.zki.S[tp.i] = signature.S
		tp.zki.R8x[tp.i] = signature.R8.X
		tp.zki.R8y[tp.i] = signature.R8.Y
	}

	// if StateDB type==TypeSynchronizer, will need to add Nonce
	if tp.s.Type() == statedb.TypeSynchronizer {
		// as tType==TypeSynchronizer, always tx.ToIdx!=0
		acc, err := tp.s.GetAccount(tx.FromIdx)
		if err != nil {
			log.Errorw("GetAccount", "fromIdx", tx.FromIdx, "err", err)
			return nil, nil, false, tracerr.Wrap(err)
		}
		tx.Nonce = acc.Nonce
		tx.TokenID = acc.TokenID
	}

	switch tx.Type {
	case common.TxTypeTransfer, common.TxTypeTransferToEthAddr, common.TxTypeTransferToBJJ:
		// go to the MT account of sender and receiver, and update
		// balance & nonce
		err = tp.applyTransfer(coordIdxsMap, collectedFees, tx.Tx(), tx.AuxToIdx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, tracerr.Wrap(err)
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, newExit, err := tp.applyExit(coordIdxsMap, collectedFees, exitTree, tx.Tx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, tracerr.Wrap(err)
		}
		return &tx.FromIdx, exitAccount, newExit, nil
	default:
	}
	return nil, nil, false, nil
}

// applyCreateAccount creates a new account in the account of the depositer, it
// stores the deposit value
func (tp *TxProcessor) applyCreateAccount(tx *common.L1Tx) error {
	account := &common.Account{
		TokenID: tx.TokenID,
		Nonce:   0,
		Balance: tx.EffectiveDepositAmount,
		BJJ:     tx.FromBJJ,
		EthAddr: tx.FromEthAddr,
	}

	p, err := tp.s.CreateAccount(common.Idx(tp.s.CurrentIdx()+1), account)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.TokenID1[tp.i] = tx.TokenID.BigInt()
		tp.zki.Nonce1[tp.i] = big.NewInt(0)
		fromBJJSign, fromBJJY := babyjub.UnpackSignY(tx.FromBJJ)
		if fromBJJSign {
			tp.zki.Sign1[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay1[tp.i] = fromBJJY
		tp.zki.Balance1[tp.i] = tx.EffectiveDepositAmount
		tp.zki.EthAddr1[tp.i] = common.EthAddrToBigInt(tx.FromEthAddr)
		tp.zki.Siblings1[tp.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			tp.zki.IsOld0_1[tp.i] = big.NewInt(1)
		}
		tp.zki.OldKey1[tp.i] = p.OldKey.BigInt()
		tp.zki.OldValue1[tp.i] = p.OldValue.BigInt()

		tp.zki.Metadata.NewLastIdxRaw = tp.s.CurrentIdx() + 1

		tp.zki.AuxFromIdx[tp.i] = common.Idx(tp.s.CurrentIdx() + 1).BigInt()
		tp.zki.NewAccount[tp.i] = big.NewInt(1)

		if tp.i < len(tp.zki.ISOnChain) { // len(tp.zki.ISOnChain) == nTx
			// intermediate states
			tp.zki.ISOnChain[tp.i] = big.NewInt(1)
		}
	}

	return tp.s.SetCurrentIdx(tp.s.CurrentIdx() + 1)
}

// applyDeposit updates the balance in the account of the depositer, if
// andTransfer parameter is set to true, the method will also apply the
// Transfer of the L1Tx/DepositTransfer
func (tp *TxProcessor) applyDeposit(tx *common.L1Tx, transfer bool) error {
	accSender, err := tp.s.GetAccount(tx.FromIdx)
	if err != nil {
		return tracerr.Wrap(err)
	}

	if tp.zki != nil {
		tp.zki.TokenID1[tp.i] = accSender.TokenID.BigInt()
		tp.zki.Nonce1[tp.i] = accSender.Nonce.BigInt()
		senderBJJSign, senderBJJY := babyjub.UnpackSignY(accSender.BJJ)
		if senderBJJSign {
			tp.zki.Sign1[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay1[tp.i] = senderBJJY
		tp.zki.Balance1[tp.i] = accSender.Balance
		tp.zki.EthAddr1[tp.i] = common.EthAddrToBigInt(accSender.EthAddr)
	}

	// add the deposit to the sender
	accSender.Balance = new(big.Int).Add(accSender.Balance, tx.EffectiveDepositAmount)
	// subtract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.EffectiveAmount)
	if accSender.Balance.Cmp(big.NewInt(0)) == -1 { // balance<0
		return newErrorNotEnoughBalance(tx.Tx())
	}

	// update sender account in localStateDB
	p, err := tp.s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.Siblings1[tp.i] = siblingsToZKInputFormat(p.Siblings)
		// IsOld0_1, OldKey1, OldValue1 not needed as this is not an insert
	}

	// in case that the tx is a L1Tx>DepositTransfer
	var accReceiver *common.Account
	if transfer {
		if tx.ToIdx == tx.FromIdx {
			accReceiver = accSender
		} else {
			accReceiver, err = tp.s.GetAccount(tx.ToIdx)
			if err != nil {
				return tracerr.Wrap(err)
			}
		}

		if tp.zki != nil {
			tp.zki.TokenID2[tp.i] = accReceiver.TokenID.BigInt()
			tp.zki.Nonce2[tp.i] = accReceiver.Nonce.BigInt()
			receiverBJJSign, receiverBJJY := babyjub.UnpackSignY(accReceiver.BJJ)
			if receiverBJJSign {
				tp.zki.Sign2[tp.i] = big.NewInt(1)
			}
			tp.zki.Ay2[tp.i] = receiverBJJY
			tp.zki.Balance2[tp.i] = accReceiver.Balance
			tp.zki.EthAddr2[tp.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
		}

		// add amount to the receiver
		accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.EffectiveAmount)

		// update receiver account in localStateDB
		p, err := tp.s.UpdateAccount(tx.ToIdx, accReceiver)
		if err != nil {
			return tracerr.Wrap(err)
		}
		if tp.zki != nil {
			tp.zki.Siblings2[tp.i] = siblingsToZKInputFormat(p.Siblings)
			// IsOld0_2, OldKey2, OldValue2 not needed as this is not an insert
		}
	}

	return nil
}

// applyTransfer updates the balance & nonce in the account of the sender, and
// the balance in the account of the receiver.
// Parameter 'toIdx' should be at 0 if the tx already has tx.ToIdx!=0, if
// tx.ToIdx==0, then toIdx!=0, and will be used the toIdx parameter as Idx of
// the receiver. This parameter is used when the tx.ToIdx is not specified and
// the real ToIdx is found trhrough the ToEthAddr or ToBJJ.
func (tp *TxProcessor) applyTransfer(coordIdxsMap map[common.TokenID]common.Idx,
	collectedFees map[common.TokenID]*big.Int, tx common.Tx, auxToIdx common.Idx) error {
	if auxToIdx == common.Idx(0) {
		auxToIdx = tx.ToIdx
	}
	// get sender and receiver accounts from localStateDB
	accSender, err := tp.s.GetAccount(tx.FromIdx)
	if err != nil {
		log.Error(err)
		return tracerr.Wrap(err)
	}

	if tp.zki != nil {
		// Set the State1 before updating the Sender leaf
		tp.zki.TokenID1[tp.i] = accSender.TokenID.BigInt()
		tp.zki.Nonce1[tp.i] = accSender.Nonce.BigInt()
		senderBJJSign, senderBJJY := babyjub.UnpackSignY(accSender.BJJ)
		if senderBJJSign {
			tp.zki.Sign1[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay1[tp.i] = senderBJJY
		tp.zki.Balance1[tp.i] = accSender.Balance
		tp.zki.EthAddr1[tp.i] = common.EthAddrToBigInt(accSender.EthAddr)
	}
	if !tx.IsL1 { // L2
		// increment nonce
		accSender.Nonce++

		// compute fee and subtract it from the accSender
		fee, err := common.CalcFeeAmount(tx.Amount, *tx.Fee)
		if err != nil {
			return tracerr.Wrap(err)
		}
		feeAndAmount := new(big.Int).Add(tx.Amount, fee)
		accSender.Balance = new(big.Int).Sub(accSender.Balance, feeAndAmount)
		if accSender.Balance.Cmp(big.NewInt(0)) == -1 { // balance<0
			return newErrorNotEnoughBalance(tx)
		}

		if _, ok := coordIdxsMap[accSender.TokenID]; ok {
			accCoord, err := tp.s.GetAccount(coordIdxsMap[accSender.TokenID])
			if err != nil {
				return tracerr.Wrap(fmt.Errorf("Can not use CoordIdx that does not exist in the tree. TokenID: %d, CoordIdx: %d", accSender.TokenID, coordIdxsMap[accSender.TokenID]))
			}
			// accumulate the fee for the Coord account
			accumulated := tp.AccumulatedFees[accCoord.Idx]
			accumulated.Add(accumulated, fee)

			if tp.s.Type() == statedb.TypeSynchronizer ||
				tp.s.Type() == statedb.TypeBatchBuilder {
				collected := collectedFees[accCoord.TokenID]
				collected.Add(collected, fee)
			}
		} else {
			log.Debugw("No coord Idx to receive fee", "tx", tx)
		}
	} else {
		accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
		if accSender.Balance.Cmp(big.NewInt(0)) == -1 { // balance<0
			return newErrorNotEnoughBalance(tx)
		}
	}

	// update sender account in localStateDB
	pSender, err := tp.s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		log.Error(err)
		return tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.Siblings1[tp.i] = siblingsToZKInputFormat(pSender.Siblings)
	}

	var accReceiver *common.Account
	if auxToIdx == tx.FromIdx {
		// if Sender is the Receiver, reuse 'accSender' pointer,
		// because in the DB the account for 'auxToIdx' won't be
		// updated yet
		accReceiver = accSender
	} else {
		accReceiver, err = tp.s.GetAccount(auxToIdx)
		if err != nil {
			log.Error(err, auxToIdx)
			return tracerr.Wrap(err)
		}
	}
	if tp.zki != nil {
		// Set the State2 before updating the Receiver leaf
		tp.zki.TokenID2[tp.i] = accReceiver.TokenID.BigInt()
		tp.zki.Nonce2[tp.i] = accReceiver.Nonce.BigInt()
		receiverBJJSign, receiverBJJY := babyjub.UnpackSignY(accReceiver.BJJ)
		if receiverBJJSign {
			tp.zki.Sign2[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay2[tp.i] = receiverBJJY
		tp.zki.Balance2[tp.i] = accReceiver.Balance
		tp.zki.EthAddr2[tp.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
	}

	// add amount-feeAmount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)

	// update receiver account in localStateDB
	pReceiver, err := tp.s.UpdateAccount(auxToIdx, accReceiver)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.Siblings2[tp.i] = siblingsToZKInputFormat(pReceiver.Siblings)
	}

	return nil
}

// applyCreateAccountDepositTransfer, in a single tx, creates a new account,
// makes a deposit, and performs a transfer to another account
func (tp *TxProcessor) applyCreateAccountDepositTransfer(tx *common.L1Tx) error {
	auxFromIdx := common.Idx(tp.s.CurrentIdx() + 1)
	accSender := &common.Account{
		TokenID: tx.TokenID,
		Nonce:   0,
		Balance: tx.EffectiveDepositAmount,
		BJJ:     tx.FromBJJ,
		EthAddr: tx.FromEthAddr,
	}

	if tp.zki != nil {
		// Set the State1 before updating the Sender leaf
		tp.zki.TokenID1[tp.i] = tx.TokenID.BigInt()
		tp.zki.Nonce1[tp.i] = big.NewInt(0)
		fromBJJSign, fromBJJY := babyjub.UnpackSignY(tx.FromBJJ)
		if fromBJJSign {
			tp.zki.Sign1[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay1[tp.i] = fromBJJY
		tp.zki.Balance1[tp.i] = tx.EffectiveDepositAmount
		tp.zki.EthAddr1[tp.i] = common.EthAddrToBigInt(tx.FromEthAddr)
	}

	// subtract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.EffectiveAmount)
	if accSender.Balance.Cmp(big.NewInt(0)) == -1 { // balance<0
		return newErrorNotEnoughBalance(tx.Tx())
	}

	// create Account of the Sender
	p, err := tp.s.CreateAccount(common.Idx(tp.s.CurrentIdx()+1), accSender)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.Siblings1[tp.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			tp.zki.IsOld0_1[tp.i] = big.NewInt(1)
		}
		tp.zki.OldKey1[tp.i] = p.OldKey.BigInt()
		tp.zki.OldValue1[tp.i] = p.OldValue.BigInt()

		tp.zki.Metadata.NewLastIdxRaw = tp.s.CurrentIdx() + 1

		tp.zki.AuxFromIdx[tp.i] = auxFromIdx.BigInt()
		tp.zki.NewAccount[tp.i] = big.NewInt(1)

		// intermediate states
		tp.zki.ISOnChain[tp.i] = big.NewInt(1)
	}
	var accReceiver *common.Account
	if tx.ToIdx == auxFromIdx {
		accReceiver = accSender
	} else {
		accReceiver, err = tp.s.GetAccount(tx.ToIdx)
		if err != nil {
			log.Error(err)
			return tracerr.Wrap(err)
		}
	}

	if tp.zki != nil {
		// Set the State2 before updating the Receiver leaf
		tp.zki.TokenID2[tp.i] = accReceiver.TokenID.BigInt()
		tp.zki.Nonce2[tp.i] = accReceiver.Nonce.BigInt()
		receiverBJJSign, receiverBJJY := babyjub.UnpackSignY(accReceiver.BJJ)
		if receiverBJJSign {
			tp.zki.Sign2[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay2[tp.i] = receiverBJJY
		tp.zki.Balance2[tp.i] = accReceiver.Balance
		tp.zki.EthAddr2[tp.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
	}

	// add amount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.EffectiveAmount)

	// update receiver account in localStateDB
	p, err = tp.s.UpdateAccount(tx.ToIdx, accReceiver)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.Siblings2[tp.i] = siblingsToZKInputFormat(p.Siblings)
	}

	return tp.s.SetCurrentIdx(tp.s.CurrentIdx() + 1)
}

// It returns the ExitAccount and a boolean determining if the Exit created a
// new Leaf in the ExitTree.
func (tp *TxProcessor) applyExit(coordIdxsMap map[common.TokenID]common.Idx,
	collectedFees map[common.TokenID]*big.Int, exitTree *merkletree.MerkleTree,
	tx common.Tx) (*common.Account, bool, error) {
	// 0. subtract tx.Amount from current Account in StateMT
	// add the tx.Amount into the Account (tx.FromIdx) in the ExitMT
	acc, err := tp.s.GetAccount(tx.FromIdx)
	if err != nil {
		return nil, false, tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.TokenID1[tp.i] = acc.TokenID.BigInt()
		tp.zki.Nonce1[tp.i] = acc.Nonce.BigInt()
		accBJJSign, accBJJY := babyjub.UnpackSignY(acc.BJJ)
		if accBJJSign {
			tp.zki.Sign1[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay1[tp.i] = accBJJY
		tp.zki.Balance1[tp.i] = acc.Balance
		tp.zki.EthAddr1[tp.i] = common.EthAddrToBigInt(acc.EthAddr)
	}

	if !tx.IsL1 {
		// increment nonce
		acc.Nonce++

		// compute fee and subtract it from the accSender
		fee, err := common.CalcFeeAmount(tx.Amount, *tx.Fee)
		if err != nil {
			return nil, false, tracerr.Wrap(err)
		}
		feeAndAmount := new(big.Int).Add(tx.Amount, fee)
		acc.Balance = new(big.Int).Sub(acc.Balance, feeAndAmount)
		if acc.Balance.Cmp(big.NewInt(0)) == -1 { // balance<0
			return nil, false, newErrorNotEnoughBalance(tx)
		}

		if _, ok := coordIdxsMap[acc.TokenID]; ok {
			accCoord, err := tp.s.GetAccount(coordIdxsMap[acc.TokenID])
			if err != nil {
				return nil, false, tracerr.Wrap(fmt.Errorf("Can not use CoordIdx that does not exist in the tree. TokenID: %d, CoordIdx: %d", acc.TokenID, coordIdxsMap[acc.TokenID]))
			}

			// accumulate the fee for the Coord account
			accumulated := tp.AccumulatedFees[accCoord.Idx]
			accumulated.Add(accumulated, fee)

			if tp.s.Type() == statedb.TypeSynchronizer ||
				tp.s.Type() == statedb.TypeBatchBuilder {
				collected := collectedFees[accCoord.TokenID]
				collected.Add(collected, fee)
			}
		} else {
			log.Debugw("No coord Idx to receive fee", "tx", tx)
		}
	} else {
		acc.Balance = new(big.Int).Sub(acc.Balance, tx.Amount)
		if acc.Balance.Cmp(big.NewInt(0)) == -1 { // balance<0
			return nil, false, newErrorNotEnoughBalance(tx)
		}
	}

	p, err := tp.s.UpdateAccount(tx.FromIdx, acc)
	if err != nil {
		return nil, false, tracerr.Wrap(err)
	}
	if tp.zki != nil {
		tp.zki.Siblings1[tp.i] = siblingsToZKInputFormat(p.Siblings)
	}

	if exitTree == nil {
		return nil, false, nil
	}
	exitAccount, err := statedb.GetAccountInTreeDB(exitTree.DB(), tx.FromIdx)
	if tracerr.Unwrap(err) == db.ErrNotFound {
		// 1a. if idx does not exist in exitTree:
		// add new leaf 'ExitTreeLeaf', where ExitTreeLeaf.Balance =
		// exitAmount (exitAmount=tx.Amount)
		exitAccount := &common.Account{
			TokenID: acc.TokenID,
			Nonce:   common.Nonce(0),
			Balance: tx.Amount,
			BJJ:     acc.BJJ,
			EthAddr: acc.EthAddr,
		}
		if tp.zki != nil {
			// Set the State2 before creating the Exit leaf
			tp.zki.TokenID2[tp.i] = acc.TokenID.BigInt()
			tp.zki.Nonce2[tp.i] = big.NewInt(0)
			accBJJSign, accBJJY := babyjub.UnpackSignY(acc.BJJ)
			if accBJJSign {
				tp.zki.Sign2[tp.i] = big.NewInt(1)
			}
			tp.zki.Ay2[tp.i] = accBJJY
			tp.zki.Balance2[tp.i] = tx.Amount
			tp.zki.EthAddr2[tp.i] = common.EthAddrToBigInt(acc.EthAddr)
			// as Leaf didn't exist in the ExitTree, set NewExit[i]=1
			tp.zki.NewExit[tp.i] = big.NewInt(1)
		}
		p, err = statedb.CreateAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
		if err != nil {
			return nil, false, tracerr.Wrap(err)
		}
		if tp.zki != nil {
			tp.zki.Siblings2[tp.i] = siblingsToZKInputFormat(p.Siblings)
			if p.IsOld0 {
				tp.zki.IsOld0_2[tp.i] = big.NewInt(1)
			}
			tp.zki.OldKey2[tp.i] = p.OldKey.BigInt()
			tp.zki.OldValue2[tp.i] = p.OldValue.BigInt()
			tp.zki.ISExitRoot[tp.i] = exitTree.Root().BigInt()
		}
		return exitAccount, true, nil
	} else if err != nil {
		return exitAccount, false, tracerr.Wrap(err)
	}

	// 1b. if idx already exist in exitTree:
	if tp.zki != nil {
		// Set the State2 before updating the Exit leaf
		tp.zki.TokenID2[tp.i] = acc.TokenID.BigInt()
		// increment nonce from existing ExitLeaf
		tp.zki.Nonce2[tp.i] = exitAccount.Nonce.BigInt()
		accBJJSign, accBJJY := babyjub.UnpackSignY(acc.BJJ)
		if accBJJSign {
			tp.zki.Sign2[tp.i] = big.NewInt(1)
		}
		tp.zki.Ay2[tp.i] = accBJJY
		tp.zki.Balance2[tp.i] = tx.Amount
		tp.zki.EthAddr2[tp.i] = common.EthAddrToBigInt(acc.EthAddr)
	}

	// update account, where account.Balance += exitAmount
	exitAccount.Balance = new(big.Int).Add(exitAccount.Balance, tx.Amount)
	p, err = statedb.UpdateAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
	if err != nil {
		return nil, false, tracerr.Wrap(err)
	}

	if tp.zki != nil {
		tp.zki.Siblings2[tp.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			tp.zki.IsOld0_2[tp.i] = big.NewInt(1)
		}
		tp.zki.OldKey2[tp.i] = p.OldKey.BigInt()
		tp.zki.OldValue2[tp.i] = p.OldValue.BigInt()
	}

	return exitAccount, false, nil
}

// computeEffectiveAmounts checks that the L1Tx data is correct
func (tp *TxProcessor) computeEffectiveAmounts(tx *common.L1Tx) {
	tx.EffectiveAmount = tx.Amount
	tx.EffectiveDepositAmount = tx.DepositAmount

	if !tx.UserOrigin {
		// case where the L1Tx is generated by the Coordinator
		tx.EffectiveAmount = big.NewInt(0)
		tx.EffectiveDepositAmount = big.NewInt(0)
		return
	}

	if tx.Type == common.TxTypeCreateAccountDeposit {
		return
	}

	if tx.ToIdx >= common.UserThreshold && tx.FromIdx == common.Idx(0) {
		// CreateAccountDepositTransfer case
		cmp := tx.DepositAmount.Cmp(tx.Amount)
		if cmp == -1 { // DepositAmount<Amount
			tx.EffectiveAmount = big.NewInt(0)
			return
		}

		// check if tx.TokenID==receiver.TokenID
		accReceiver, err := tp.s.GetAccount(tx.ToIdx)
		if err != nil {
			log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: can not get account for tx.ToIdx: %d", tx.ToIdx)
			tx.EffectiveDepositAmount = big.NewInt(0)
			tx.EffectiveAmount = big.NewInt(0)
			return
		}
		if tx.TokenID != accReceiver.TokenID {
			log.Debugf("EffectiveAmount = 0: tx TokenID (%d) != receiver account TokenID (%d)", tx.TokenID, accReceiver.TokenID)
			tx.EffectiveAmount = big.NewInt(0)
			return
		}
		return
	}

	accSender, err := tp.s.GetAccount(tx.FromIdx)
	if err != nil {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: can not get account for tx.FromIdx: %d", tx.FromIdx)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}

	// check that tx.TokenID corresponds to the Sender account TokenID
	if tx.TokenID != accSender.TokenID {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: tx.TokenID (%d) !=sender account TokenID (%d)", tx.TokenID, accSender.TokenID)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}

	// check that Sender has enough balance
	bal := accSender.Balance
	if tx.DepositAmount != nil {
		bal = new(big.Int).Add(bal, tx.EffectiveDepositAmount)
	}
	cmp := bal.Cmp(tx.Amount)
	if cmp == -1 {
		log.Debugf("EffectiveAmount = 0: Not enough funds (%s<%s)", bal.String(), tx.Amount.String())
		tx.EffectiveAmount = big.NewInt(0)
		return
	}

	// check that the tx.FromEthAddr is the same than the EthAddress of the
	// Sender
	if !bytes.Equal(tx.FromEthAddr.Bytes(), accSender.EthAddr.Bytes()) {
		log.Debugf("EffectiveAmount = 0: tx.FromEthAddr (%s) must be the same EthAddr of the sender account by the Idx (%s)", tx.FromEthAddr.Hex(), accSender.EthAddr.Hex())
		tx.EffectiveAmount = big.NewInt(0)
	}

	if tx.ToIdx == common.Idx(1) || tx.ToIdx == common.Idx(0) {
		// if transfer is Exit type, there are no more checks
		return
	}

	// check that TokenID is the same for Sender & Receiver account
	accReceiver, err := tp.s.GetAccount(tx.ToIdx)
	if err != nil {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: can not get account for tx.ToIdx: %d", tx.ToIdx)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}
	if accSender.TokenID != accReceiver.TokenID {
		log.Debugf("EffectiveAmount = 0: sender account TokenID (%d) != receiver account TokenID (%d)", accSender.TokenID, accReceiver.TokenID)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}
	if tx.TokenID != accReceiver.TokenID {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: tx TokenID (%d) != receiver account TokenID (%d)", tx.TokenID, accReceiver.TokenID)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}
}

// CheckEnoughBalance returns true if the sender of the transaction has enough
// balance in the account to send the Amount+Fee, and also returns the account
// Balance and the Fee+Amount (which is used to give information about why the
// transaction is not selected in case that this method returns false.
func (tp *TxProcessor) CheckEnoughBalance(tx common.PoolL2Tx) (bool, *big.Int, *big.Int) {
	acc, err := tp.s.GetAccount(tx.FromIdx)
	if err != nil {
		return false, nil, nil
	}
	fee, err := common.CalcFeeAmount(tx.Amount, tx.Fee)
	if err != nil {
		return false, nil, nil
	}
	feeAndAmount := new(big.Int).Add(tx.Amount, fee)
	return acc.Balance.Cmp(feeAndAmount) != -1, // !=-1 balance<amount
		acc.Balance, feeAndAmount
}
