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
			StateRoot: big.NewInt(0), ExitRoot: big.NewInt(0)},
	}
}

// Context contains the data of the test
type Context struct {
	Instructions          []instruction
	userNames             []string
	Users                 map[string]*User
	lastRegisteredTokenID common.TokenID
	l1CreatedAccounts     map[string]*Account

	// rollupConstMaxL1UserTx Maximum L1-user transactions allowed to be queued in a batch
	rollupConstMaxL1UserTx int

	idx           int
	currBlock     common.BlockData
	currBatch     common.BatchData
	currBatchNum  int
	queues        [][]L1Tx
	toForgeNum    int
	openToForge   int
	currBatchTest struct {
		l1CoordinatorTxs []L1Tx
		l2Txs            []L2Tx
	}
	blockNum int64
}

// NewContext returns a new Context
func NewContext(rollupConstMaxL1UserTx int) *Context {
	currBatchNum := 1 // The protocol defines the first batchNum to be 1
	return &Context{
		Users:                 make(map[string]*User),
		l1CreatedAccounts:     make(map[string]*Account),
		lastRegisteredTokenID: 0,

		rollupConstMaxL1UserTx: rollupConstMaxL1UserTx,
		idx:                    common.UserThreshold,
		// We use some placeholder values for StateRoot and ExitTree
		// because these values will never be nil
		currBatch:    newBatchData(currBatchNum),
		currBatchNum: currBatchNum,
		// start with 2 queues, one for toForge, and the other for openToForge
		queues:      make([][]L1Tx, 2),
		toForgeNum:  0,
		openToForge: 1,
		//nolint:gomnd
		blockNum: 2, // rollup genesis blockNum
	}
}

// Account contains the data related to the account for a specific TokenID of a User
type Account struct {
	Idx   common.Idx
	Nonce common.Nonce
}

// User contains the data related to a testing user
type User struct {
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
				LoadAmount:  big.NewInt(int64(inst.loadAmount)),
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
				Amount: big.NewInt(int64(inst.amount)),
				Fee:    common.FeeSelector(inst.fee),
				Type:   common.TxTypeTransfer,
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
				ToIdx:  common.Idx(1), // as is an Exit
				Amount: big.NewInt(int64(inst.amount)),
				Type:   common.TxTypeExit,
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
			// for each L1UserTx of the queues[ToForgeNum], calculate the Idx
			if err = tc.calculateIdxForL1Txs(false, tc.queues[tc.toForgeNum]); err != nil {
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
			// advance batch
			tc.toForgeNum++
			if tc.toForgeNum == tc.openToForge {
				tc.openToForge++
				newQueue := []L1Tx{}
				tc.queues = append(tc.queues, newQueue)
			}
		case typeNewBlock:
			tc.currBlock.Block = common.Block{
				EthBlockNum: tc.blockNum,
			}
			blocks = append(blocks, tc.currBlock)
			tc.blockNum++
			tc.currBlock = common.BlockData{}
		case typeAddToken:
			newToken := common.Token{
				EthAddr: ethCommon.BigToAddress(big.NewInt(int64(inst.tokenID * 100))), //nolint:gomnd
				// Name:        fmt.Sprintf("Token %d", inst.tokenID),
				// Symbol:      fmt.Sprintf("TK%d", inst.tokenID),
				// Decimals:    18,
				TokenID:     inst.tokenID,
				EthBlockNum: tc.blockNum,
			}
			if inst.tokenID != tc.lastRegisteredTokenID+1 {
				return nil, fmt.Errorf("Line %d: AddToken TokenID should be sequential, expected TokenID: %d, defined TokenID: %d", inst.lineNum, tc.lastRegisteredTokenID+1, inst.tokenID)
			}
			tc.lastRegisteredTokenID++
			tc.currBlock.AddedTokens = append(tc.currBlock.AddedTokens, newToken)
		default:
			return nil, fmt.Errorf("Line %d: Unexpected type: %s", inst.lineNum, inst.typ)
		}
	}

	return blocks, nil
}

// calculateIdxsForL1Txs calculates new Idx for new created accounts. If
// 'isCoordinatorTxs==true', adds the tx to tc.currBatch.L1CoordinatorTxs.
func (tc *Context) calculateIdxForL1Txs(isCoordinatorTxs bool, txs []L1Tx) error {
	// for each batch.L1CoordinatorTxs of the queues[ToForgeNum], calculate the Idx
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		if tx.L1Tx.Type == common.TxTypeCreateAccountDeposit || tx.L1Tx.Type == common.TxTypeCreateAccountDepositTransfer {
			if tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID] != nil { // if account already exists, return error
				return fmt.Errorf("Can not create same account twice (same User & same TokenID) (this is a design property of Til)")
			}
			tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID] = &Account{
				Idx:   common.Idx(tc.idx),
				Nonce: common.Nonce(0),
			}
			tc.l1CreatedAccounts[idxTokenIDToString(tx.fromIdxName, tx.L1Tx.TokenID)] = tc.Users[tx.fromIdxName].Accounts[tx.L1Tx.TokenID]
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
		testTx.L2Tx.Nonce = tc.Users[testTx.fromIdxName].Accounts[testTx.tokenID].Nonce

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
	tc.currBlock.Batches = append(tc.currBlock.Batches, tc.currBatch)
	tc.currBatchNum++
	tc.currBatch = newBatchData(tc.currBatchNum)
	tc.currBatchTest.l1CoordinatorTxs = nil
	tc.currBatchTest.l2Txs = nil
	return nil
}

// addToL1Queue adds the L1Tx into the queue that is open and has space
func (tc *Context) addToL1Queue(tx L1Tx) error {
	if len(tc.queues[tc.openToForge]) >= tc.rollupConstMaxL1UserTx {
		// if current OpenToForge queue reached its Max, move into a
		// new queue
		tc.openToForge++
		newQueue := []L1Tx{}
		tc.queues = append(tc.queues, newQueue)
	}
	// Fill L1UserTx specific parameters
	tx.L1Tx.UserOrigin = true
	toForgeL1TxsNum := int64(tc.openToForge)
	tx.L1Tx.ToForgeL1TxsNum = &toForgeL1TxsNum
	tx.L1Tx.EthBlockNum = tc.blockNum
	tx.L1Tx.Position = len(tc.queues[tc.openToForge])

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

	tc.queues[tc.openToForge] = append(tc.queues[tc.openToForge], tx)
	tc.currBlock.L1UserTxs = append(tc.currBlock.L1UserTxs, tx.L1Tx)

	return nil
}

func (tc *Context) checkIfAccountExists(tf string, inst instruction) error {
	if tc.Users[tf].Accounts[inst.tokenID] == nil {
		return fmt.Errorf("%s at User: %s, for TokenID: %d, while account not created yet", inst.typ, tf, inst.tokenID)
	}
	return nil
}
func (tc *Context) checkIfTokenIsRegistered(inst instruction) error {
	if inst.tokenID > tc.lastRegisteredTokenID {
		return fmt.Errorf("Can not process %s: TokenID %d not registered, last registered TokenID: %d", inst.typ, inst.tokenID, tc.lastRegisteredTokenID)
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
				TokenID: inst.tokenID,
				Amount:  big.NewInt(int64(inst.amount)),
				Nonce:   tc.Users[inst.from].Accounts[inst.tokenID].Nonce,
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
			BJJ:      &sk,
			Addr:     addr,
			Accounts: make(map[common.TokenID]*Account),
		}
		tc.Users[userNames[i-1]] = &u
	}
}
