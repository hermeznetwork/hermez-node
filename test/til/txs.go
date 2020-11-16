package til

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

func newBatchData(batchNum int) common.BatchData {
	return common.BatchData{
		L1CoordinatorTxs: []common.L1Tx{},
		L2Txs:            []common.L2Tx{},
		Batch: common.Batch{
			BatchNum:  common.BatchNum(batchNum),
			StateRoot: big.NewInt(0), ExitRoot: big.NewInt(0),
			FeeIdxsCoordinator: make([]common.Idx, 0),
			CollectedFees:      make(map[common.TokenID]*big.Int),
		},
	}
}

func newBlock(blockNum int64) common.BlockData {
	return common.BlockData{
		Block: common.Block{
			EthBlockNum: blockNum,
		},
		Rollup: common.RollupData{
			L1UserTxs: []common.L1Tx{},
		},
	}
}

type contextExtra struct {
	openToForge     int64
	toForgeL1TxsNum int64
	nonces          map[common.Idx]common.Nonce
	idx             int
}

// Context contains the data of the test
type Context struct {
	Instructions          []instruction
	userNames             []string
	Users                 map[string]*User // Name -> *User
	UsersByIdx            map[int]*User
	accountsByIdx         map[int]*Account
	LastRegisteredTokenID common.TokenID
	l1CreatedAccounts     map[string]*Account // (Name, TokenID) -> *Account

	// rollupConstMaxL1UserTx Maximum L1-user transactions allowed to be queued in a batch
	rollupConstMaxL1UserTx int

	idx           int
	currBlock     common.BlockData
	currBatch     common.BatchData
	currBatchNum  int
	Queues        [][]L1Tx
	ToForgeNum    int
	openToForge   int
	currBatchTest struct {
		l1CoordinatorTxs []L1Tx
		l2Txs            []L2Tx
	}
	blockNum int64

	extra contextExtra
}

// NewContext returns a new Context
func NewContext(rollupConstMaxL1UserTx int) *Context {
	currBatchNum := 1 // The protocol defines the first batchNum to be 1
	return &Context{
		Users:                 make(map[string]*User),
		l1CreatedAccounts:     make(map[string]*Account),
		UsersByIdx:            make(map[int]*User),
		accountsByIdx:         make(map[int]*Account),
		LastRegisteredTokenID: 0,

		rollupConstMaxL1UserTx: rollupConstMaxL1UserTx,
		idx:                    common.UserThreshold,
		// We use some placeholder values for StateRoot and ExitTree
		// because these values will never be nil
		currBlock:    newBlock(2), //nolint:gomnd
		currBatch:    newBatchData(currBatchNum),
		currBatchNum: currBatchNum,
		// start with 2 queues, one for toForge, and the other for openToForge
		Queues:      make([][]L1Tx, 2),
		ToForgeNum:  0,
		openToForge: 1,
		//nolint:gomnd
		blockNum: 2, // rollup genesis blockNum
		extra: contextExtra{
			openToForge:     0,
			toForgeL1TxsNum: 0,
			nonces:          make(map[common.Idx]common.Nonce),
			idx:             common.UserThreshold,
		},
	}
}

// Account contains the data related to the account for a specific TokenID of a User
type Account struct {
	Idx      common.Idx
	TokenID  common.TokenID
	Nonce    common.Nonce
	BatchNum int
}

// User contains the data related to a testing user
type User struct {
	Name     string
	BJJ      *babyjub.PrivateKey
	Addr     ethCommon.Address
	Accounts map[common.TokenID]*Account
}

// L1Tx is the data structure used internally for transaction test generation,
// which contains a common.L1Tx data plus some intermediate data for the
// transaction generation.
type L1Tx struct {
	lineNum     int
	fromIdxName string
	toIdxName   string

	L1Tx common.L1Tx
}

// L2Tx is the data structure used internally for transaction test generation,
// which contains a common.L2Tx data plus some intermediate data for the
// transaction generation.
type L2Tx struct {
	lineNum     int
	fromIdxName string
	toIdxName   string
	tokenID     common.TokenID
	L2Tx        common.L2Tx
}

// GenerateBlocks returns an array of BlockData for a given set. It uses the
// users (keys & nonces) of the Context.
func (tc *Context) GenerateBlocks(set string) ([]common.BlockData, error) {
	parser := newParser(strings.NewReader(set))
	parsedSet, err := parser.parse()
	if err != nil {
		return nil, err
	}
	if parsedSet.typ != setTypeBlockchain {
		return nil, fmt.Errorf("Expected set type: %s, found: %s", setTypeBlockchain, parsedSet.typ)
	}

	tc.Instructions = parsedSet.instructions
	tc.userNames = parsedSet.users

	tc.generateKeys(tc.userNames)

	var blocks []common.BlockData
	for _, inst := range parsedSet.instructions {
		switch inst.typ {
		case txTypeCreateAccountDepositCoordinator: // tx source: L1CoordinatorTx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L1Tx{
				FromEthAddr: tc.Users[inst.from].Addr,
				FromBJJ:     tc.Users[inst.from].BJJ.Public(),
				TokenID:     inst.tokenID,
				Amount:      big.NewInt(0),
				LoadAmount:  big.NewInt(0),
				Type:        common.TxTypeCreateAccountDeposit, // as txTypeCreateAccountDepositCoordinator is not valid oustide Til package
			}
			testTx := L1Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				L1Tx:        tx,
			}

			tc.currBatchTest.l1CoordinatorTxs = append(tc.currBatchTest.l1CoordinatorTxs, testTx)
		case common.TxTypeCreateAccountDeposit, common.TxTypeCreateAccountDepositTransfer: // tx source: L1UserTx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L1Tx{
				FromEthAddr: tc.Users[inst.from].Addr,
				FromBJJ:     tc.Users[inst.from].BJJ.Public(),
				TokenID:     inst.tokenID,
				Amount:      big.NewInt(0),
				LoadAmount:  big.NewInt(int64(inst.loadAmount)),
				Type:        inst.typ,
			}
			if inst.typ == common.TxTypeCreateAccountDepositTransfer {
				tx.Amount = big.NewInt(int64(inst.amount))
			}
			testTx := L1Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				toIdxName:   inst.to,
				L1Tx:        tx,
			}
			if err := tc.addToL1Queue(testTx); err != nil {
				return nil, err
			}
		case common.TxTypeDeposit, common.TxTypeDepositTransfer: // tx source: L1UserTx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			if err := tc.checkIfAccountExists(inst.from, inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L1Tx{
				TokenID:    inst.tokenID,
				Amount:     big.NewInt(0),
				LoadAmount: big.NewInt(int64(inst.loadAmount)),
				Type:       inst.typ,
			}
			if inst.typ == common.TxTypeDepositTransfer {
				tx.Amount = big.NewInt(int64(inst.amount))
			}
			testTx := L1Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				toIdxName:   inst.to,
				L1Tx:        tx,
			}
			if err := tc.addToL1Queue(testTx); err != nil {
				return nil, err
			}
		case common.TxTypeTransfer: // L2Tx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L2Tx{
				Amount:      big.NewInt(int64(inst.amount)),
				Fee:         common.FeeSelector(inst.fee),
				Type:        common.TxTypeTransfer,
				EthBlockNum: tc.blockNum,
			}
			tx.BatchNum = common.BatchNum(tc.currBatchNum) // when converted to PoolL2Tx BatchNum parameter is lost
			testTx := L2Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				toIdxName:   inst.to,
				tokenID:     inst.tokenID,
				L2Tx:        tx,
			}
			tc.currBatchTest.l2Txs = append(tc.currBatchTest.l2Txs, testTx)
		case common.TxTypeForceTransfer: // tx source: L1UserTx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L1Tx{
				TokenID:    inst.tokenID,
				Amount:     big.NewInt(int64(inst.amount)),
				LoadAmount: big.NewInt(0),
				Type:       common.TxTypeForceTransfer,
			}
			testTx := L1Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				toIdxName:   inst.to,
				L1Tx:        tx,
			}
			if err := tc.addToL1Queue(testTx); err != nil {
				return nil, err
			}
		case common.TxTypeExit: // tx source: L2Tx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L2Tx{
				ToIdx:       common.Idx(1), // as is an Exit
				Fee:         common.FeeSelector(inst.fee),
				Amount:      big.NewInt(int64(inst.amount)),
				Type:        common.TxTypeExit,
				EthBlockNum: tc.blockNum,
			}
			tx.BatchNum = common.BatchNum(tc.currBatchNum) // when converted to PoolL2Tx BatchNum parameter is lost
			testTx := L2Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				toIdxName:   inst.to,
				tokenID:     inst.tokenID,
				L2Tx:        tx,
			}
			tc.currBatchTest.l2Txs = append(tc.currBatchTest.l2Txs, testTx)
		case common.TxTypeForceExit: // tx source: L1UserTx
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx := common.L1Tx{
				ToIdx:      common.Idx(1), // as is an Exit
				TokenID:    inst.tokenID,
				Amount:     big.NewInt(int64(inst.amount)),
				LoadAmount: big.NewInt(0),
				Type:       common.TxTypeForceExit,
			}
			testTx := L1Tx{
				lineNum:     inst.lineNum,
				fromIdxName: inst.from,
				toIdxName:   inst.to,
				L1Tx:        tx,
			}
			if err := tc.addToL1Queue(testTx); err != nil {
				return nil, err
			}
		case typeNewBatch:
			if err = tc.calculateIdxForL1Txs(true, tc.currBatchTest.l1CoordinatorTxs); err != nil {
				return nil, err
			}
			if err = tc.setIdxs(); err != nil {
				log.Error(err)
				return nil, err
			}
		case typeNewBatchL1:
			// for each L1UserTx of the Queues[ToForgeNum], calculate the Idx
			if err = tc.calculateIdxForL1Txs(false, tc.Queues[tc.ToForgeNum]); err != nil {
				return nil, err
			}
			if err = tc.calculateIdxForL1Txs(true, tc.currBatchTest.l1CoordinatorTxs); err != nil {
				return nil, err
			}
			tc.currBatch.L1Batch = true
			if err = tc.setIdxs(); err != nil {
				log.Error(err)
				return nil, err
			}
			toForgeL1TxsNum := int64(tc.openToForge)
			tc.currBatch.Batch.ForgeL1TxsNum = &toForgeL1TxsNum
			// advance batch
			tc.ToForgeNum++
			if tc.ToForgeNum == tc.openToForge {
				tc.openToForge++
				newQueue := []L1Tx{}
				tc.Queues = append(tc.Queues, newQueue)
			}
		case typeNewBlock:
			blocks = append(blocks, tc.currBlock)
			tc.blockNum++
			tc.currBlock = newBlock(tc.blockNum)
		case typeAddToken:
			newToken := common.Token{
				EthAddr: ethCommon.BigToAddress(big.NewInt(int64(inst.tokenID * 100))), //nolint:gomnd
				// Name:        fmt.Sprintf("Token %d", inst.tokenID),
				// Symbol:      fmt.Sprintf("TK%d", inst.tokenID),
				// Decimals:    18,
				TokenID:     inst.tokenID,
				EthBlockNum: tc.blockNum,
			}
			if inst.tokenID != tc.LastRegisteredTokenID+1 {
				return nil, fmt.Errorf("Line %d: AddToken TokenID should be sequential, expected TokenID: %d, defined TokenID: %d", inst.lineNum, tc.LastRegisteredTokenID+1, inst.tokenID)
			}
			tc.LastRegisteredTokenID++
			tc.currBlock.Rollup.AddedTokens = append(tc.currBlock.Rollup.AddedTokens, newToken)
		default:
			return nil, fmt.Errorf("Line %d: Unexpected type: %s", inst.lineNum, inst.typ)
		}
	}

	return blocks, nil
}

// calculateIdxsForL1Txs calculates new Idx for new created accounts. If
// 'isCoordinatorTxs==true', adds the tx to tc.currBatch.L1CoordinatorTxs.
func (tc *Context) calculateIdxForL1Txs(isCoordinatorTxs bool, txs []L1Tx) error {
	// for each batch.L1CoordinatorTxs of the Queues[ToForgeNum], calculate the Idx
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		if tx.L1Tx.Type == common.TxTypeCreateAccountDeposit || tx.L1Tx.Type == common.TxTypeCreateAccountDepositTransfer {
			if tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID] != nil { // if account already exists, return error
				return fmt.Errorf("Can not create same account twice (same User (%s) & same TokenID (%d)) (this is a design property of Til)", tx.fromIdxName, tx.L1Tx.TokenID)
			}
			tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID] = &Account{
				Idx:      common.Idx(tc.idx),
				TokenID:  tx.L1Tx.TokenID,
				Nonce:    common.Nonce(0),
				BatchNum: tc.currBatchNum,
			}
			tc.l1CreatedAccounts[idxTokenIDToString(tx.fromIdxName, tx.L1Tx.TokenID)] = tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID]
			tc.accountsByIdx[tc.idx] = tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID]
			tc.UsersByIdx[tc.idx] = tc.Users[tx.fromIdxName]
			tc.idx++
		}
		if isCoordinatorTxs {
			tc.currBatch.L1CoordinatorTxs = append(tc.currBatch.L1CoordinatorTxs, tx.L1Tx)
		}
	}
	return nil
}

// setIdxs sets the Idxs to the transactions of the tc.currBatch
func (tc *Context) setIdxs() error {
	// once Idxs are calculated, update transactions to use the new Idxs
	for i := 0; i < len(tc.currBatchTest.l2Txs); i++ {
		testTx := &tc.currBatchTest.l2Txs[i]

		if tc.Users[testTx.fromIdxName].Accounts[testTx.tokenID] == nil {
			return fmt.Errorf("Line %d: %s from User %s for TokenID %d while account not created yet", testTx.lineNum, testTx.L2Tx.Type, testTx.fromIdxName, testTx.tokenID)
		}
		if testTx.L2Tx.Type == common.TxTypeTransfer {
			if _, ok := tc.l1CreatedAccounts[idxTokenIDToString(testTx.toIdxName, testTx.tokenID)]; !ok {
				return fmt.Errorf("Line %d: Can not create Transfer for a non existing account. Batch %d, ToIdx name: %s, TokenID: %d", testTx.lineNum, tc.currBatchNum, testTx.toIdxName, testTx.tokenID)
			}
		}
		tc.Users[testTx.fromIdxName].Accounts[testTx.tokenID].Nonce++
		// next line is commented to avoid Blockchain L2Txs to have
		// Nonce different from 0, as from Blockchain those
		// transactions will come without Nonce
		// testTx.L2Tx.Nonce = tc.Users[testTx.fromIdxName].Accounts[testTx.tokenID].Nonce

		// set real Idx
		testTx.L2Tx.FromIdx = tc.Users[testTx.fromIdxName].Accounts[testTx.tokenID].Idx
		if testTx.L2Tx.Type == common.TxTypeTransfer {
			testTx.L2Tx.ToIdx = tc.Users[testTx.toIdxName].Accounts[testTx.tokenID].Idx
		}
		// in case Type==Exit, ToIdx=1, already set at the
		// GenerateBlocks main switch inside TxTypeExit case

		nTx, err := common.NewL2Tx(&testTx.L2Tx)
		if err != nil {
			return fmt.Errorf("Line %d: %s", testTx.lineNum, err.Error())
		}
		testTx.L2Tx = *nTx

		tc.currBatch.L2Txs = append(tc.currBatch.L2Txs, testTx.L2Tx)
	}

	tc.currBatch.Batch.LastIdx = int64(tc.idx - 1) // `-1` because tc.idx is the next available idx
	tc.currBlock.Rollup.Batches = append(tc.currBlock.Rollup.Batches, tc.currBatch)
	tc.currBatchNum++
	tc.currBatch = newBatchData(tc.currBatchNum)
	tc.currBatchTest.l1CoordinatorTxs = nil
	tc.currBatchTest.l2Txs = nil
	return nil
}

// addToL1Queue adds the L1Tx into the queue that is open and has space
func (tc *Context) addToL1Queue(tx L1Tx) error {
	if len(tc.Queues[tc.openToForge]) >= tc.rollupConstMaxL1UserTx {
		// if current OpenToForge queue reached its Max, move into a
		// new queue
		tc.openToForge++
		newQueue := []L1Tx{}
		tc.Queues = append(tc.Queues, newQueue)
	}
	// Fill L1UserTx specific parameters
	tx.L1Tx.UserOrigin = true
	toForgeL1TxsNum := int64(tc.openToForge)
	tx.L1Tx.ToForgeL1TxsNum = &toForgeL1TxsNum
	tx.L1Tx.EthBlockNum = tc.blockNum
	tx.L1Tx.Position = len(tc.Queues[tc.openToForge])

	// When an L1UserTx is generated, all idxs must be available (except when idx == 0 or idx == 1)
	if tx.L1Tx.Type != common.TxTypeCreateAccountDeposit && tx.L1Tx.Type != common.TxTypeCreateAccountDepositTransfer {
		tx.L1Tx.FromIdx = tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID].Idx
	}
	tx.L1Tx.FromEthAddr = tc.Users[tx.fromIdxName].Addr
	tx.L1Tx.FromBJJ = tc.Users[tx.fromIdxName].BJJ.Public()
	if tx.toIdxName == "" {
		tx.L1Tx.ToIdx = common.Idx(0)
	} else {
		account, ok := tc.Users[tx.toIdxName].Accounts[tx.L1Tx.TokenID]
		if !ok {
			return fmt.Errorf("Line %d: Transfer to User: %s, for TokenID: %d, "+
				"while account not created yet", tx.lineNum, tx.toIdxName, tx.L1Tx.TokenID)
		}
		tx.L1Tx.ToIdx = account.Idx
	}
	if tx.L1Tx.Type == common.TxTypeForceExit {
		tx.L1Tx.ToIdx = common.Idx(1)
	}
	nTx, err := common.NewL1Tx(&tx.L1Tx)
	if err != nil {
		return fmt.Errorf("Line %d: %s", tx.lineNum, err.Error())
	}
	tx.L1Tx = *nTx

	tc.Queues[tc.openToForge] = append(tc.Queues[tc.openToForge], tx)
	tc.currBlock.Rollup.L1UserTxs = append(tc.currBlock.Rollup.L1UserTxs, tx.L1Tx)

	return nil
}

func (tc *Context) checkIfAccountExists(tf string, inst instruction) error {
	if tc.Users[tf].Accounts[inst.tokenID] == nil {
		return fmt.Errorf("%s at User: %s, for TokenID: %d, while account not created yet", inst.typ, tf, inst.tokenID)
	}
	return nil
}
func (tc *Context) checkIfTokenIsRegistered(inst instruction) error {
	if inst.tokenID > tc.LastRegisteredTokenID {
		return fmt.Errorf("Can not process %s: TokenID %d not registered, last registered TokenID: %d", inst.typ, inst.tokenID, tc.LastRegisteredTokenID)
	}
	return nil
}

// GeneratePoolL2Txs returns an array of common.PoolL2Tx from a given set. It
// uses the users (keys) of the Context.
func (tc *Context) GeneratePoolL2Txs(set string) ([]common.PoolL2Tx, error) {
	parser := newParser(strings.NewReader(set))
	parsedSet, err := parser.parse()
	if err != nil {
		return nil, err
	}
	if parsedSet.typ != setTypePoolL2 {
		return nil, fmt.Errorf("Expected set type: %s, found: %s", setTypePoolL2, parsedSet.typ)
	}

	tc.Instructions = parsedSet.instructions
	tc.userNames = parsedSet.users

	tc.generateKeys(tc.userNames)

	txs := []common.PoolL2Tx{}
	for _, inst := range tc.Instructions {
		switch inst.typ {
		case common.TxTypeTransfer, common.TxTypeTransferToEthAddr, common.TxTypeTransferToBJJ:
			if err := tc.checkIfAccountExists(inst.from, inst); err != nil {
				log.Error(err)
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			if inst.typ == common.TxTypeTransfer {
				// if TxTypeTransfer, need to exist the ToIdx account
				if err := tc.checkIfAccountExists(inst.to, inst); err != nil {
					log.Error(err)
					return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
				}
			}
			tc.Users[inst.from].Accounts[inst.tokenID].Nonce++
			// if account of receiver does not exist, don't use
			// ToIdx, and use only ToEthAddr & ToBJJ
			tx := common.PoolL2Tx{
				FromIdx:     tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				TokenID:     inst.tokenID,
				Amount:      big.NewInt(int64(inst.amount)),
				Fee:         common.FeeSelector(inst.fee),
				Nonce:       tc.Users[inst.from].Accounts[inst.tokenID].Nonce,
				State:       common.PoolL2TxStatePending,
				Timestamp:   time.Now(),
				RqToEthAddr: common.EmptyAddr,
				RqToBJJ:     nil,
				Type:        inst.typ,
			}
			if tx.Type == common.TxTypeTransfer {
				tx.ToIdx = tc.Users[inst.to].Accounts[inst.tokenID].Idx
				tx.ToEthAddr = tc.Users[inst.to].Addr
				tx.ToBJJ = tc.Users[inst.to].BJJ.Public()
			} else if tx.Type == common.TxTypeTransferToEthAddr {
				tx.ToIdx = common.Idx(0)
				tx.ToEthAddr = tc.Users[inst.to].Addr
			} else if tx.Type == common.TxTypeTransferToBJJ {
				tx.ToIdx = common.Idx(0)
				tx.ToEthAddr = common.FFAddr
				tx.ToBJJ = tc.Users[inst.to].BJJ.Public()
			}
			nTx, err := common.NewPoolL2Tx(&tx)
			if err != nil {
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx = *nTx
			// perform signature and set it to tx.Signature
			toSign, err := tx.HashToSign()
			if err != nil {
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			sig := tc.Users[inst.from].BJJ.SignPoseidon(toSign)
			tx.Signature = sig.Compress()

			txs = append(txs, tx)
		case common.TxTypeExit:
			tc.Users[inst.from].Accounts[inst.tokenID].Nonce++
			tx := common.PoolL2Tx{
				FromIdx: tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				ToIdx:   common.Idx(1), // as is an Exit
				Fee:     common.FeeSelector(inst.fee),
				TokenID: inst.tokenID,
				Amount:  big.NewInt(int64(inst.amount)),
				Nonce:   tc.Users[inst.from].Accounts[inst.tokenID].Nonce,
				State:   common.PoolL2TxStatePending,
				Type:    common.TxTypeExit,
			}
			nTx, err := common.NewPoolL2Tx(&tx)
			if err != nil {
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			tx = *nTx
			// perform signature and set it to tx.Signature
			toSign, err := tx.HashToSign()
			if err != nil {
				return nil, fmt.Errorf("Line %d: %s", inst.lineNum, err.Error())
			}
			sig := tc.Users[inst.from].BJJ.SignPoseidon(toSign)
			tx.Signature = sig.Compress()
			txs = append(txs, tx)
		default:
			return nil, fmt.Errorf("Line %d: instruction type unrecognized: %s", inst.lineNum, inst.typ)
		}
	}

	return txs, nil
}

// generateKeys generates BabyJubJub & Address keys for the given list of user
// names in a deterministic way. This means, that for the same given
// 'userNames' in a certain order, the keys will be always the same.
func (tc *Context) generateKeys(userNames []string) {
	for i := 1; i < len(userNames)+1; i++ {
		if _, ok := tc.Users[userNames[i-1]]; ok {
			// account already created
			continue
		}
		// babyjubjub key
		var sk babyjub.PrivateKey
		copy(sk[:], []byte(strconv.Itoa(i))) // only for testing

		// eth address
		var key ecdsa.PrivateKey
		key.D = big.NewInt(int64(i)) // only for testing
		key.PublicKey.X, key.PublicKey.Y = ethCrypto.S256().ScalarBaseMult(key.D.Bytes())
		key.Curve = ethCrypto.S256()
		addr := ethCrypto.PubkeyToAddress(key.PublicKey)

		u := User{
			Name:     userNames[i-1],
			BJJ:      &sk,
			Addr:     addr,
			Accounts: make(map[common.TokenID]*Account),
		}
		tc.Users[userNames[i-1]] = &u
	}
}

// L1TxsToCommonL1Txs converts an array of []til.L1Tx to []common.L1Tx
func L1TxsToCommonL1Txs(l1 []L1Tx) []common.L1Tx {
	var r []common.L1Tx
	for i := 0; i < len(l1); i++ {
		r = append(r, l1[i].L1Tx)
	}
	return r
}

// ConfigExtra is the configuration used in FillBlocksExtra to extend the
// blocks returned by til.
type ConfigExtra struct {
	// Address to set as forger for each batch
	BootCoordAddr ethCommon.Address
	// Coordinator user name used to select the corresponding accounts to
	// collect coordinator fees
	CoordUser string
}

// FillBlocksL1UserTxsBatchNum fills the BatchNum of forged L1UserTxs:
// - blocks[].Rollup.L1UserTxs[].BatchNum
func (tc *Context) FillBlocksL1UserTxsBatchNum(blocks []common.BlockData) {
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Rollup.Batches {
			batch := &block.Rollup.Batches[j]
			if batch.L1Batch {
				// Set BatchNum for forged L1UserTxs to til blocks
				bn := batch.Batch.BatchNum
				for k := range blocks {
					block := &blocks[k]
					for l := range block.Rollup.L1UserTxs {
						tx := &block.Rollup.L1UserTxs[l]
						if *tx.ToForgeL1TxsNum == tc.extra.openToForge {
							tx.BatchNum = &bn
						}
					}
				}
				tc.extra.openToForge++
			}
		}
	}
}

// FillBlocksExtra fills extra fields not generated by til in each block, so
// that the blockData is closer to what the HistoryDB stores.  The filled fields are:
// - blocks[].Rollup.Batch.EthBlockNum
// - blocks[].Rollup.Batch.ForgerAddr
// - blocks[].Rollup.Batch.ForgeL1TxsNum
// - blocks[].Rollup.Batch.L1CoordinatorTxs[].TxID
// - blocks[].Rollup.Batch.L1CoordinatorTxs[].BatchNum
// - blocks[].Rollup.Batch.L1CoordinatorTxs[].EthBlockNum
// - blocks[].Rollup.Batch.L1CoordinatorTxs[].Position
// - blocks[].Rollup.Batch.L2Txs[].TxID
// - blocks[].Rollup.Batch.L2Txs[].Position
// - blocks[].Rollup.Batch.L2Txs[].Nonce
// - blocks[].Rollup.Batch.ExitTree
// - blocks[].Rollup.Batch.CreatedAccounts
// - blocks[].Rollup.Batch.FeeIdxCoordinator
// - blocks[].Rollup.Batch.CollectedFees
func (tc *Context) FillBlocksExtra(blocks []common.BlockData, cfg *ConfigExtra) error {
	// Fill extra fields not generated by til in til block
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Rollup.Batches {
			batch := &block.Rollup.Batches[j]
			batch.Batch.EthBlockNum = block.Block.EthBlockNum
			// til doesn't fill the batch forger addr
			batch.Batch.ForgerAddr = cfg.BootCoordAddr
			if batch.L1Batch {
				toForgeL1TxsNumCpy := tc.extra.toForgeL1TxsNum
				// til doesn't fill the ForgeL1TxsNum
				batch.Batch.ForgeL1TxsNum = &toForgeL1TxsNumCpy
				tc.extra.toForgeL1TxsNum++
			}

			batchNum := batch.Batch.BatchNum
			for k := range batch.L1CoordinatorTxs {
				tx := &batch.L1CoordinatorTxs[k]
				tx.BatchNum = &batchNum
				tx.EthBlockNum = batch.Batch.EthBlockNum
			}
		}
	}

	// Fill CreatedAccounts
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Rollup.Batches {
			batch := &block.Rollup.Batches[j]
			l1Txs := []common.L1Tx{}
			if batch.L1Batch {
				for _, tx := range tc.Queues[*batch.Batch.ForgeL1TxsNum] {
					l1Txs = append(l1Txs, tx.L1Tx)
				}
			}
			l1Txs = append(l1Txs, batch.L1CoordinatorTxs...)
			for k := range l1Txs {
				tx := &l1Txs[k]
				if tx.Type == common.TxTypeCreateAccountDeposit ||
					tx.Type == common.TxTypeCreateAccountDepositTransfer {
					user, ok := tc.UsersByIdx[tc.extra.idx]
					if !ok {
						return fmt.Errorf("Created account with idx: %v not found", tc.extra.idx)
					}
					batch.CreatedAccounts = append(batch.CreatedAccounts,
						common.Account{
							Idx:       common.Idx(tc.extra.idx),
							TokenID:   tx.TokenID,
							BatchNum:  batch.Batch.BatchNum,
							PublicKey: user.BJJ.Public(),
							EthAddr:   user.Addr,
							Nonce:     0,
							Balance:   big.NewInt(0),
						})
					tc.extra.idx++
				}
			}
		}
	}

	// Fill expected positions in L1CoordinatorTxs and L2Txs
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Rollup.Batches {
			batch := &block.Rollup.Batches[j]
			position := 0
			if batch.L1Batch {
				position = len(tc.Queues[*batch.Batch.ForgeL1TxsNum])
			}
			for k := range batch.L1CoordinatorTxs {
				tx := &batch.L1CoordinatorTxs[k]
				tx.Position = position
				position++
				nTx, err := common.NewL1Tx(tx)
				if err != nil {
					return err
				}
				*tx = *nTx
			}
			for k := range batch.L2Txs {
				tx := &batch.L2Txs[k]
				tx.Position = position
				position++
				tc.extra.nonces[tx.FromIdx]++
				tx.Nonce = tc.extra.nonces[tx.FromIdx]
				nTx, err := common.NewL2Tx(tx)
				if err != nil {
					return err
				}
				*tx = *nTx
			}
		}
	}

	// Fill ExitTree (only AccountIdx and Balance)
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Rollup.Batches {
			batch := &block.Rollup.Batches[j]
			if batch.L1Batch {
				for _, _tx := range tc.Queues[*batch.Batch.ForgeL1TxsNum] {
					tx := _tx.L1Tx
					if tx.Type == common.TxTypeForceExit {
						batch.ExitTree =
							append(batch.ExitTree,
								common.ExitInfo{
									BatchNum:   batch.Batch.BatchNum,
									AccountIdx: tx.FromIdx,
									Balance:    tx.Amount,
								})
					}
				}
			}
			for k := range batch.L2Txs {
				tx := &batch.L2Txs[k]
				if tx.Type == common.TxTypeExit {
					batch.ExitTree = append(batch.ExitTree, common.ExitInfo{
						BatchNum:   batch.Batch.BatchNum,
						AccountIdx: tx.FromIdx,
						Balance:    tx.Amount,
					})
				}
				fee, err := common.CalcFeeAmount(tx.Amount, tx.Fee)
				if err != nil {
					return err
				}

				// Find the TokenID of the tx
				fromAcc, ok := tc.accountsByIdx[int(tx.FromIdx)]
				if !ok {
					return fmt.Errorf("L2tx.FromIdx idx: %v not found", tx.FromIdx)
				}

				// Find the idx of the CoordUser for the
				// TokenID, and if it exists, add the fee to
				// the collectedFees.  Only consider the
				// coordinator account to receive fee if it was
				// created in this or a previous batch
				if acc, ok := tc.l1CreatedAccounts[idxTokenIDToString(cfg.CoordUser, fromAcc.TokenID)]; ok &&
					common.BatchNum(acc.BatchNum) <= batch.Batch.BatchNum {
					found := false
					for _, idx := range batch.Batch.FeeIdxsCoordinator {
						if idx == common.Idx(acc.Idx) {
							found = true
							break
						}
					}
					if !found {
						batch.Batch.FeeIdxsCoordinator = append(batch.Batch.FeeIdxsCoordinator,
							common.Idx(acc.Idx))
						batch.Batch.CollectedFees[fromAcc.TokenID] = big.NewInt(0)
					}
					collected := batch.Batch.CollectedFees[fromAcc.TokenID]
					collected.Add(collected, fee)
				}
			}
		}
	}
	return nil
}
