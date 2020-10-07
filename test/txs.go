package test

import (
	"crypto/ecdsa"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/require"
)

// TestContext contains the data of the test
type TestContext struct {
	t                 *testing.T
	Instructions      []Instruction
	accountsNames     []string
	accounts          map[string]*Account
	TokenIDs          []common.TokenID
	l1CreatedAccounts map[string]*Account
}

// NewTestContext returns a new TestContext
func NewTestContext(t *testing.T) *TestContext {
	return &TestContext{
		t:                 t,
		accounts:          make(map[string]*Account),
		l1CreatedAccounts: make(map[string]*Account),
	}
}

// Account contains the data related to a testing account
type Account struct {
	BJJ   *babyjub.PrivateKey
	Addr  ethCommon.Address
	Idx   common.Idx
	Nonce common.Nonce
}

// BlockData contains the information of a Block
type BlockData struct {
	block *common.Block // ethereum block
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
func (tc *TestContext) GenerateBlocks(set string) []BlockData {
	parser := NewParser(strings.NewReader(set))
	parsedSet, err := parser.Parse(false)
	require.Nil(tc.t, err)

	tc.Instructions = parsedSet.Instructions
	tc.accountsNames = parsedSet.Accounts
	tc.TokenIDs = parsedSet.TokenIDs

	tc.generateKeys(tc.accountsNames)

	var blocks []BlockData
	currBatchNum := 0
	var currBlock BlockData
	var currBatch BatchData
	idx := 256
	for _, inst := range parsedSet.Instructions {
		switch inst.Type {
		case common.TxTypeCreateAccountDeposit, common.TxTypeCreateAccountDepositTransfer:
			tx := common.L1Tx{
				// TxID
				FromEthAddr: tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Addr,
				FromBJJ:     tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				LoadAmount:  big.NewInt(int64(inst.LoadAmount)),
				Type:        inst.Type,
			}
			if tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx == common.Idx(0) { // if account.Idx is not set yet, set it and increment idx
				tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx = common.Idx(idx)

				tc.l1CreatedAccounts[idxTokenIDToString(inst.From, inst.TokenID)] = tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)]
				idx++
			}
			if inst.Type == common.TxTypeCreateAccountDepositTransfer {
				tx.Amount = big.NewInt(int64(inst.Amount))
			}
			currBatch.L1UserTxs = append(currBatch.L1UserTxs, tx)
		case common.TxTypeDeposit, common.TxTypeDepositTransfer:
			tx := common.L1Tx{
				// TxID
				FromIdx:     &tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				FromEthAddr: tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Addr,
				FromBJJ:     tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				LoadAmount:  big.NewInt(int64(inst.LoadAmount)),
				Type:        inst.Type,
			}
			if tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx == common.Idx(0) {
				// if account.Idx is not set yet, set it and increment idx
				tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx = common.Idx(idx)

				tc.l1CreatedAccounts[idxTokenIDToString(inst.From, inst.TokenID)] = tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)]
				idx++
			}
			if inst.Type == common.TxTypeDepositTransfer {
				tx.Amount = big.NewInt(int64(inst.Amount))
				// if ToIdx is not set yet, set it and increment idx
				if tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx == common.Idx(0) {
					tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx = common.Idx(idx)

					tc.l1CreatedAccounts[idxTokenIDToString(inst.To, inst.TokenID)] = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)]
					tx.ToIdx = common.Idx(idx)
					idx++
				} else {
					// if Idx account of To already exist, use it for ToIdx
					tx.ToIdx = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx
				}
			}
			currBatch.L1UserTxs = append(currBatch.L1UserTxs, tx)
		case common.TxTypeTransfer:
			tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			// if account of receiver does not exist, create a new CoordinatorL1Tx creating the account
			if _, ok := tc.l1CreatedAccounts[idxTokenIDToString(inst.To, inst.TokenID)]; !ok {
				tx := common.L1Tx{
					FromEthAddr: tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
					FromBJJ:     tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
					TokenID:     inst.TokenID,
					LoadAmount:  big.NewInt(int64(inst.Amount)),
					Type:        common.TxTypeCreateAccountDeposit,
				}
				tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx = common.Idx(idx)
				tc.l1CreatedAccounts[idxTokenIDToString(inst.To, inst.TokenID)] = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)]
				currBatch.L1CoordinatorTxs = append(currBatch.L1CoordinatorTxs, tx)
				idx++
			}
			toEthAddr := new(ethCommon.Address)
			*toEthAddr = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr
			rqToEthAddr := new(ethCommon.Address)
			*rqToEthAddr = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr
			tx := common.L2Tx{
				FromIdx:  tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:    tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx,
				Amount:   big.NewInt(int64(inst.Amount)),
				Fee:      common.FeeSelector(inst.Fee),
				Nonce:    tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				BatchNum: common.BatchNum(currBatchNum),
				Type:     common.TxTypeTransfer,
			}
			nTx, err := common.NewPoolL2Tx(tx.PoolL2Tx())
			if err != nil {
				panic(err)
			}
			nL2Tx, err := nTx.L2Tx()
			if err != nil {
				panic(err)
			}
			tx = *nL2Tx

			currBatch.L2Txs = append(currBatch.L2Txs, tx)
		case common.TxTypeExit:
			tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			tx := common.L2Tx{
				FromIdx: tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:   common.Idx(1), // as is an Exit
				Amount:  big.NewInt(int64(inst.Amount)),
				Nonce:   tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				Type:    common.TxTypeExit,
			}
			nTx, err := common.NewPoolL2Tx(tx.PoolL2Tx())
			if err != nil {
				panic(err)
			}
			nL2Tx, err := nTx.L2Tx()
			if err != nil {
				panic(err)
			}
			tx = *nL2Tx
			currBatch.L2Txs = append(currBatch.L2Txs, tx)
		case common.TxTypeForceExit:
			fromIdx := new(common.Idx)
			*fromIdx = tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx
			tx := common.L1Tx{
				FromIdx: fromIdx,
				ToIdx:   common.Idx(1), // as is an Exit
				TokenID: inst.TokenID,
				Amount:  big.NewInt(int64(inst.Amount)),
				Type:    common.TxTypeExit,
			}
			currBatch.L1UserTxs = append(currBatch.L1UserTxs, tx)
		case TypeNewBatch:
			currBlock.Batches = append(currBlock.Batches, currBatch)
			currBatchNum++
			currBatch = BatchData{}
		case TypeNewBlock:
			currBlock.Batches = append(currBlock.Batches, currBatch)
			currBatchNum++
			currBatch = BatchData{}
			blocks = append(blocks, currBlock)
			currBlock = BlockData{}
		default:
			log.Fatalf("Unexpected type: %s", inst.Type)
		}
	}
	currBlock.Batches = append(currBlock.Batches, currBatch)
	blocks = append(blocks, currBlock)

	return blocks
}

// GeneratePoolL2Txs returns an array of common.PoolL2Tx from a given set. It
// uses the accounts (keys & nonces) of the TestContext.
func (tc *TestContext) GeneratePoolL2Txs(set string) []common.PoolL2Tx {
	parser := NewParser(strings.NewReader(set))
	parsedSet, err := parser.Parse(true)
	require.Nil(tc.t, err)

	tc.Instructions = parsedSet.Instructions
	tc.accountsNames = parsedSet.Accounts
	tc.TokenIDs = parsedSet.TokenIDs

	tc.generateKeys(tc.accountsNames)

	txs := []common.PoolL2Tx{}
	for _, inst := range tc.Instructions {
		switch inst.Type {
		case common.TxTypeTransfer:
			tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			// if account of receiver does not exist, don't use
			// ToIdx, and use only ToEthAddr & ToBJJ
			toIdx := new(common.Idx)
			if _, ok := tc.l1CreatedAccounts[idxTokenIDToString(inst.To, inst.TokenID)]; !ok {
				*toIdx = 0
			} else {
				*toIdx = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx
			}
			// TODO once common.L{x}Txs parameter pointers is undone, update this lines related to pointers usage
			toEthAddr := new(ethCommon.Address)
			*toEthAddr = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr
			rqToEthAddr := new(ethCommon.Address)
			*rqToEthAddr = tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr
			tx := common.PoolL2Tx{
				FromIdx:     tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:       toIdx,
				ToEthAddr:   toEthAddr,
				ToBJJ:       tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				Amount:      big.NewInt(int64(inst.Amount)),
				Fee:         common.FeeSelector(inst.Fee),
				Nonce:       tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				State:       common.PoolL2TxStatePending,
				Timestamp:   time.Now(),
				BatchNum:    nil,
				RqToEthAddr: rqToEthAddr,
				RqToBJJ:     tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
				Type:        common.TxTypeTransfer,
			}
			nTx, err := common.NewPoolL2Tx(&tx)
			if err != nil {
				panic(err)
			}
			tx = *nTx
			// perform signature and set it to tx.Signature
			toSign, err := tx.HashToSign()
			if err != nil {
				panic(err)
			}
			sig := tc.accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.SignPoseidon(toSign)
			tx.Signature = sig

			txs = append(txs, tx)
		case common.TxTypeExit:
			tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			// TODO once common.L{x}Txs parameter pointers is undone, update this lines related to pointers usage
			toIdx := new(common.Idx)
			*toIdx = common.Idx(1) // as is an Exit
			tx := common.PoolL2Tx{
				FromIdx: tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:   toIdx, // as is an Exit
				TokenID: inst.TokenID,
				Amount:  big.NewInt(int64(inst.Amount)),
				Nonce:   tc.accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				Type:    common.TxTypeExit,
			}
			txs = append(txs, tx)
		default:
			log.Fatalf("instruction type unrecognized: %s", inst.Type)
		}
	}

	return txs
}

// generateKeys generates BabyJubJub & Address keys for the given list of
// account names in a deterministic way. This means, that for the same given
// 'accNames' in a certain order, the keys will be always the same.
func (tc *TestContext) generateKeys(accNames []string) map[string]*Account {
	acc := make(map[string]*Account)
	for i := 1; i < len(accNames)+1; i++ {
		if _, ok := tc.accounts[accNames[i-1]]; ok {
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

		a := Account{
			BJJ:   &sk,
			Addr:  addr,
			Nonce: 0,
		}
		tc.accounts[accNames[i-1]] = &a
	}
	return acc
}
