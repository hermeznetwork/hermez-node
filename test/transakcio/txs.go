package transakcio

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
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// TestContext contains the data of the test
type TestContext struct {
	Instructions          []instruction
	accountsNames         []string
	Users                 map[string]*User
	lastRegisteredTokenID common.TokenID
	l1CreatedAccounts     map[string]*Account
}

// NewTestContext returns a new TestContext
func NewTestContext() *TestContext {
	return &TestContext{
		Users:             make(map[string]*User),
		l1CreatedAccounts: make(map[string]*Account),
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

// BlockData contains the information of a Block
type BlockData struct {
	// block *common.Block // ethereum block
	// L1UserTxs that were submitted in the block
	L1UserTxs        []common.L1Tx
	Batches          []BatchData
	RegisteredTokens []common.Token
}

// BatchData contains the information of a Batch
type BatchData struct {
	L1Batch bool // TODO: Remove once Batch.ForgeL1TxsNum is a pointer
	// L1UserTxs that were forged in the batch
	L1UserTxs        []common.L1Tx
	L1CoordinatorTxs []common.L1Tx
	L2Txs            []common.L2Tx
	CreatedAccounts  []common.Account
	ExitTree         []common.ExitInfo
	Batch            *common.Batch
}

// GenerateBlocks returns an array of BlockData for a given set. It uses the
// accounts (keys & nonces) of the TestContext.
func (tc *TestContext) GenerateBlocks(set string) ([]BlockData, error) {
	parser := newParser(strings.NewReader(set))
	parsedSet, err := parser.parse()
	if err != nil {
		return nil, err
	}

	tc.Instructions = parsedSet.instructions
	tc.accountsNames = parsedSet.accounts

	tc.generateKeys(tc.accountsNames)

	var blocks []BlockData
	currBatchNum := 0
	var currBlock BlockData
	var currBatch BatchData
	idx := 256
	for _, inst := range parsedSet.instructions {
		switch inst.typ {
		case common.TxTypeCreateAccountDeposit, common.TxTypeCreateAccountDepositTransfer, txTypeCreateAccountDepositCoordinator:
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				return nil, err
			}
			tx := common.L1Tx{
				// TxID
				FromEthAddr: tc.Users[inst.from].Addr,
				FromBJJ:     tc.Users[inst.from].BJJ.Public(),
				TokenID:     inst.tokenID,
				LoadAmount:  big.NewInt(int64(inst.loadAmount)),
				Type:        inst.typ,
			}
			if tc.Users[inst.from].Accounts[inst.tokenID] == nil { // if account is not set yet, set it and increment idx
				tc.Users[inst.from].Accounts[inst.tokenID] = &Account{
					Idx:   common.Idx(idx),
					Nonce: common.Nonce(0),
				}

				tc.l1CreatedAccounts[idxTokenIDToString(inst.from, inst.tokenID)] = tc.Users[inst.from].Accounts[inst.tokenID]
				idx++
			}
			if inst.typ == common.TxTypeCreateAccountDepositTransfer {
				tx.Amount = big.NewInt(int64(inst.amount))
			}
			if inst.typ == txTypeCreateAccountDepositCoordinator {
				tx.Type = common.TxTypeCreateAccountDeposit // as txTypeCreateAccountDepositCoordinator is not valid oustide Transakcio package
				currBatch.L1CoordinatorTxs = append(currBatch.L1CoordinatorTxs, tx)
			} else {
				currBatch.L1UserTxs = append(currBatch.L1UserTxs, tx)
			}
		case common.TxTypeDeposit, common.TxTypeDepositTransfer:
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				return nil, err
			}
			if tc.Users[inst.from].Accounts[inst.tokenID] == nil {
				return nil, fmt.Errorf("Deposit at User %s for TokenID %d while account not created yet", inst.from, inst.tokenID)
			}
			tx := common.L1Tx{
				// TxID
				FromIdx:     tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				FromEthAddr: tc.Users[inst.from].Addr,
				FromBJJ:     tc.Users[inst.from].BJJ.Public(),
				TokenID:     inst.tokenID,
				LoadAmount:  big.NewInt(int64(inst.loadAmount)),
				Type:        inst.typ,
			}
			if tc.Users[inst.from].Accounts[inst.tokenID].Idx == common.Idx(0) {
				// if account.Idx is not set yet, set it and increment idx
				tc.Users[inst.from].Accounts[inst.tokenID].Idx = common.Idx(idx)

				tc.l1CreatedAccounts[idxTokenIDToString(inst.from, inst.tokenID)] = tc.Users[inst.from].Accounts[inst.tokenID]
				idx++
			}
			if inst.typ == common.TxTypeDepositTransfer {
				tx.Amount = big.NewInt(int64(inst.amount))
				// if ToIdx is not set yet, set it and increment idx
				if tc.Users[inst.to].Accounts[inst.tokenID].Idx == common.Idx(0) {
					tc.Users[inst.to].Accounts[inst.tokenID].Idx = common.Idx(idx)

					tc.l1CreatedAccounts[idxTokenIDToString(inst.to, inst.tokenID)] = tc.Users[inst.to].Accounts[inst.tokenID]
					tx.ToIdx = common.Idx(idx)
					idx++
				} else {
					// if Idx account of To already exist, use it for ToIdx
					tx.ToIdx = tc.Users[inst.to].Accounts[inst.tokenID].Idx
				}
			}
			currBatch.L1UserTxs = append(currBatch.L1UserTxs, tx)
		case common.TxTypeTransfer:
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				return nil, err
			}
			if tc.Users[inst.from].Accounts[inst.tokenID] == nil {
				return nil, fmt.Errorf("Transfer from User %s for TokenID %d while account not created yet", inst.from, inst.tokenID)
			}
			// if account of receiver does not exist, create a new CoordinatorL1Tx creating the account
			if _, ok := tc.l1CreatedAccounts[idxTokenIDToString(inst.to, inst.tokenID)]; !ok {
				return nil, fmt.Errorf("Can not create Transfer for a non existing account. Batch %d, Instruction: %s", currBatchNum, inst)
			}
			tc.Users[inst.from].Accounts[inst.tokenID].Nonce++
			tx := common.L2Tx{
				FromIdx: tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				ToIdx:   tc.Users[inst.to].Accounts[inst.tokenID].Idx,
				Amount:  big.NewInt(int64(inst.amount)),
				Fee:     common.FeeSelector(inst.fee),
				Nonce:   tc.Users[inst.from].Accounts[inst.tokenID].Nonce,
				Type:    common.TxTypeTransfer,
			}
			nTx, err := common.NewPoolL2Tx(tx.PoolL2Tx())
			if err != nil {
				return nil, err
			}
			tx = nTx.L2Tx()
			tx.BatchNum = common.BatchNum(currBatchNum) // when converted to PoolL2Tx BatchNum parameter is lost

			currBatch.L2Txs = append(currBatch.L2Txs, tx)
		case common.TxTypeExit:
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				return nil, err
			}
			tc.Users[inst.from].Accounts[inst.tokenID].Nonce++
			tx := common.L2Tx{
				FromIdx: tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				ToIdx:   common.Idx(1), // as is an Exit
				Amount:  big.NewInt(int64(inst.amount)),
				Nonce:   tc.Users[inst.from].Accounts[inst.tokenID].Nonce,
				Type:    common.TxTypeExit,
			}
			nTx, err := common.NewPoolL2Tx(tx.PoolL2Tx())
			if err != nil {
				return nil, err
			}
			tx = nTx.L2Tx()
			currBatch.L2Txs = append(currBatch.L2Txs, tx)
		case common.TxTypeForceExit:
			if err := tc.checkIfTokenIsRegistered(inst); err != nil {
				return nil, err
			}
			tx := common.L1Tx{
				FromIdx: tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				ToIdx:   common.Idx(1), // as is an Exit
				TokenID: inst.tokenID,
				Amount:  big.NewInt(int64(inst.amount)),
				Type:    common.TxTypeExit,
			}
			currBatch.L1UserTxs = append(currBatch.L1UserTxs, tx)
		case typeNewBatch:
			currBlock.Batches = append(currBlock.Batches, currBatch)
			currBatchNum++
			currBatch = BatchData{}
		case typeNewBlock:
			currBlock.Batches = append(currBlock.Batches, currBatch)
			currBatchNum++
			currBatch = BatchData{}
			blocks = append(blocks, currBlock)
			currBlock = BlockData{}
		case typeRegisterToken:
			newToken := common.Token{
				TokenID:     inst.tokenID,
				EthBlockNum: int64(len(blocks)),
			}
			if inst.tokenID != tc.lastRegisteredTokenID+1 {
				return nil, fmt.Errorf("RegisterToken TokenID should be sequential, expected TokenID: %d, defined TokenID: %d", tc.lastRegisteredTokenID+1, inst.tokenID)
			}
			tc.lastRegisteredTokenID++
			currBlock.RegisteredTokens = append(currBlock.RegisteredTokens, newToken)
		default:
			return nil, fmt.Errorf("Unexpected type: %s", inst.typ)
		}
	}
	currBlock.Batches = append(currBlock.Batches, currBatch)
	blocks = append(blocks, currBlock)

	return blocks, nil
}
func (tc *TestContext) checkIfTokenIsRegistered(inst instruction) error {
	if inst.tokenID > tc.lastRegisteredTokenID {
		return fmt.Errorf("Can not process %s: TokenID %d not registered, last registered TokenID: %d", inst.typ, inst.tokenID, tc.lastRegisteredTokenID)
	}
	return nil
}

// GeneratePoolL2Txs returns an array of common.PoolL2Tx from a given set. It
// uses the accounts (keys & nonces) of the TestContext.
func (tc *TestContext) GeneratePoolL2Txs(set string) ([]common.PoolL2Tx, error) {
	parser := newParser(strings.NewReader(set))
	parsedSet, err := parser.parse()
	if err != nil {
		return nil, err
	}

	tc.Instructions = parsedSet.instructions
	tc.accountsNames = parsedSet.accounts

	tc.generateKeys(tc.accountsNames)

	txs := []common.PoolL2Tx{}
	for _, inst := range tc.Instructions {
		switch inst.typ {
		case common.TxTypeTransfer:
			if tc.Users[inst.from].Accounts[inst.tokenID] == nil {
				return nil, fmt.Errorf("Transfer from User %s for TokenID %d while account not created yet", inst.from, inst.tokenID)
			}
			if tc.Users[inst.to].Accounts[inst.tokenID] == nil {
				return nil, fmt.Errorf("Transfer to User %s for TokenID %d while account not created yet", inst.to, inst.tokenID)
			}
			tc.Users[inst.from].Accounts[inst.tokenID].Nonce++
			// if account of receiver does not exist, don't use
			// ToIdx, and use only ToEthAddr & ToBJJ
			tx := common.PoolL2Tx{
				FromIdx:     tc.Users[inst.from].Accounts[inst.tokenID].Idx,
				ToIdx:       tc.Users[inst.to].Accounts[inst.tokenID].Idx,
				ToEthAddr:   tc.Users[inst.to].Addr,
				ToBJJ:       tc.Users[inst.to].BJJ.Public(),
				TokenID:     inst.tokenID,
				Amount:      big.NewInt(int64(inst.amount)),
				Fee:         common.FeeSelector(inst.fee),
				Nonce:       tc.Users[inst.from].Accounts[inst.tokenID].Nonce,
				State:       common.PoolL2TxStatePending,
				Timestamp:   time.Now(),
				RqToEthAddr: common.EmptyAddr,
				RqToBJJ:     nil,
				Type:        common.TxTypeTransfer,
			}
			nTx, err := common.NewPoolL2Tx(&tx)
			if err != nil {
				return nil, err
			}
			tx = *nTx
			// perform signature and set it to tx.Signature
			toSign, err := tx.HashToSign()
			if err != nil {
				return nil, err
			}
			sig := tc.Users[inst.to].BJJ.SignPoseidon(toSign)
			tx.Signature = sig

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
			txs = append(txs, tx)
		default:
			return nil, fmt.Errorf("instruction type unrecognized: %s", inst.typ)
		}
	}

	return txs, nil
}

// generateKeys generates BabyJubJub & Address keys for the given list of
// account names in a deterministic way. This means, that for the same given
// 'accNames' in a certain order, the keys will be always the same.
func (tc *TestContext) generateKeys(accNames []string) {
	for i := 1; i < len(accNames)+1; i++ {
		if _, ok := tc.Users[accNames[i-1]]; ok {
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
		tc.Users[accNames[i-1]] = &u
	}
}
