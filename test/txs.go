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

type TestContext struct {
	t                 *testing.T
	Instructions      []Instruction
	accountsNames     []string
	accounts          map[string]*Account
	TokenIDs          []common.TokenID
	l1CreatedAccounts map[string]*Account
}

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

// func (tc *TestContext) GenerateBlocks() []BlockData {
//
//         return nil
// }

// GeneratePoolL2Txs returns an array of common.PoolL2Tx from a given set. It
// uses the accounts (keys & nonces) of the TestContext.
func (tc *TestContext) GeneratePoolL2Txs(set string) []common.PoolL2Tx {
	parser := NewParser(strings.NewReader(set))
	parsedSet, err := parser.Parse()
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
			log.Warnf("instruction type unrecognized: %s", inst.Type)
			continue
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

/*
// GenerateTestTxs generates L1Tx & PoolL2Tx in a deterministic way for the
// given ParsedSet.
func GenerateTestTxs(t *testing.T, parsedSet *ParsedSet) ([][]common.L1Tx, [][]common.L1Tx, [][]common.PoolL2Tx, []common.Token) {
	accounts := generateKeys(t, parsedSet.Accounts)
	l1CreatedAccounts := make(map[string]*Account)

	var batchL1Txs []common.L1Tx
	var batchCoordinatorL1Txs []common.L1Tx
	var batchPoolL2Txs []common.PoolL2Tx
	var l1Txs [][]common.L1Tx
	var coordinatorL1Txs [][]common.L1Tx
	var poolL2Txs [][]common.PoolL2Tx
	idx := 256
	for _, inst := range parsedSet.Instructions {
		switch inst.Type {
		case common.TxTypeCreateAccountDeposit:
			tx := common.L1Tx{
				// TxID
				FromEthAddr: accounts[idxTokenIDToString(inst.From, inst.TokenID)].Addr,
				FromBJJ:     accounts[idxTokenIDToString(inst.From, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				Amount:      big.NewInt(0),
				LoadAmount:  big.NewInt(int64(inst.Amount)),
				Type:        common.TxTypeCreateAccountDeposit,
			}
			batchL1Txs = append(batchL1Txs, tx)
			if accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx == common.Idx(0) { // if account.Idx is not set yet, set it and increment idx
				accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx = common.Idx(idx)

				l1CreatedAccounts[idxTokenIDToString(inst.From, inst.TokenID)] = accounts[idxTokenIDToString(inst.From, inst.TokenID)]
				idx++
			}
		case common.TxTypeTransfer:
			// if account of receiver does not exist, create a new CoordinatorL1Tx creating the account
			if _, ok := l1CreatedAccounts[idxTokenIDToString(inst.To, inst.TokenID)]; !ok {
				tx := common.L1Tx{
					FromEthAddr: accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
					FromBJJ:     accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
					TokenID:     inst.TokenID,
					LoadAmount:  big.NewInt(int64(inst.Amount)),
					Type:        common.TxTypeCreateAccountDeposit,
				}
				accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx = common.Idx(idx)
				l1CreatedAccounts[idxTokenIDToString(inst.To, inst.TokenID)] = accounts[idxTokenIDToString(inst.To, inst.TokenID)]
				batchCoordinatorL1Txs = append(batchCoordinatorL1Txs, tx)
				idx++
			}
			tx := common.PoolL2Tx{
				FromIdx:     accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:       accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx,
				ToEthAddr:   accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
				ToBJJ:       accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				Amount:      big.NewInt(int64(inst.Amount)),
				Fee:         common.FeeSelector(inst.Fee),
				Nonce:       accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				State:       common.PoolL2TxStatePending,
				RqToEthAddr: accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
				RqToBJJ:     accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
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
			sig := accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.SignPoseidon(toSign)
			tx.Signature = sig

			accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			batchPoolL2Txs = append(batchPoolL2Txs, tx)

		case common.TxTypeExit, common.TxTypeForceExit:
			tx := common.L1Tx{
				FromIdx: accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:   common.Idx(1), // as is an Exit
				TokenID: inst.TokenID,
				Amount:  big.NewInt(int64(inst.Amount)),
				Type:    common.TxTypeExit,
			}
			batchL1Txs = append(batchL1Txs, tx)
		case TypeNewBatch:
			l1Txs = append(l1Txs, batchL1Txs)
			coordinatorL1Txs = append(coordinatorL1Txs, batchCoordinatorL1Txs)
			poolL2Txs = append(poolL2Txs, batchPoolL2Txs)
			batchL1Txs = []common.L1Tx{}
			batchCoordinatorL1Txs = []common.L1Tx{}
			batchPoolL2Txs = []common.PoolL2Tx{}
		default:
			continue
		}
	}
	l1Txs = append(l1Txs, batchL1Txs)
	coordinatorL1Txs = append(coordinatorL1Txs, batchCoordinatorL1Txs)
	poolL2Txs = append(poolL2Txs, batchPoolL2Txs)
	tokens := []common.Token{}
	for i := 0; i < len(poolL2Txs); i++ {
		for j := 0; j < len(poolL2Txs[i]); j++ {
			id := poolL2Txs[i][j].TokenID
			found := false
			for k := 0; k < len(tokens); k++ {
				if tokens[k].TokenID == id {
					found = true
					break
				}
			}
			if !found {
				tokens = append(tokens, common.Token{
					TokenID:     id,
					EthBlockNum: 1,
					EthAddr:     ethCommon.BigToAddress(big.NewInt(int64(i*10000 + j))),
				})
			}
		}
	}
	return l1Txs, coordinatorL1Txs, poolL2Txs, tokens
}

// GenerateTestTxsFromSet reurns the L1 & L2 transactions for a given Set of
// Instructions code
func GenerateTestTxsFromSet(t *testing.T, set string) ([][]common.L1Tx, [][]common.L1Tx, [][]common.PoolL2Tx, []common.Token) {
	parser := NewParser(strings.NewReader(set))
	parsedSet, err := parser.Parse()
	require.Nil(t, err)

	return GenerateTestTxs(t, parsedSet)
}
*/
