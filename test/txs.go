package test

import (
	"crypto/ecdsa"
	"math/big"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type Account struct {
	BJJ   *babyjub.PrivateKey
	Addr  ethCommon.Address
	Idx   common.Idx
	Nonce common.Nonce
}

// GenerateKeys generates BabyJubJub & Address keys for the given list of
// account names in a deterministic way. This means, that for the same given
// 'accNames' the keys will be always the same.
func GenerateKeys(t *testing.T, accNames []string) map[string]*Account {
	acc := make(map[string]*Account)
	for i := 1; i < len(accNames)+1; i++ {
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
		acc[accNames[i-1]] = &a
	}
	return acc
}

// GenerateTestTxs generates L1Tx & PoolL2Tx in a deterministic way for the
// given Instructions.
func GenerateTestTxs(t *testing.T, instructions Instructions) ([][]*common.L1Tx, [][]*common.L1Tx, [][]*common.PoolL2Tx) {
	accounts := GenerateKeys(t, instructions.Accounts)
	l1CreatedAccounts := make(map[string]*Account)

	var batchL1txs []*common.L1Tx
	var batchCoordinatorL1txs []*common.L1Tx
	var batchL2txs []*common.PoolL2Tx
	var l1txs [][]*common.L1Tx
	var coordinatorL1txs [][]*common.L1Tx
	var l2txs [][]*common.PoolL2Tx
	idx := 1
	for _, inst := range instructions.Instructions {
		switch inst.Type {
		case common.TxTypeCreateAccountDeposit:
			tx := common.L1Tx{
				// TxID
				FromEthAddr: accounts[idxTokenIDToString(inst.From, inst.TokenID)].Addr,
				FromBJJ:     accounts[idxTokenIDToString(inst.From, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				LoadAmount:  big.NewInt(int64(inst.Amount)),
				Type:        common.TxTypeCreateAccountDeposit,
			}
			batchL1txs = append(batchL1txs, &tx)
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
				batchCoordinatorL1txs = append(batchCoordinatorL1txs, &tx)
				idx++
			}

			tx := common.PoolL2Tx{
				// TxID: nil,
				FromIdx:     accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:       accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx,
				ToEthAddr:   accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
				ToBJJ:       accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
				TokenID:     inst.TokenID,
				Amount:      big.NewInt(int64(inst.Amount)),
				Fee:         common.FeeSelector(inst.Fee),
				Nonce:       accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				State:       common.PoolL2TxStatePending,
				Timestamp:   time.Now(),
				BatchNum:    0,
				RqToEthAddr: accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
				RqToBJJ:     accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
				Type:        common.TxTypeTransfer,
			}
			// perform signature and set it to tx.Signature
			toSign, err := tx.HashToSign()
			if err != nil {
				panic(err)
			}
			sig := accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.SignPoseidon(toSign)
			tx.Signature = sig

			accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			batchL2txs = append(batchL2txs, &tx)

		case common.TxTypeExit, common.TxTypeForceExit:
			tx := common.L1Tx{
				FromIdx: accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:   common.Idx(1), // as is an Exit
				TokenID: inst.TokenID,
				Amount:  big.NewInt(int64(inst.Amount)),
				Type:    common.TxTypeExit,
			}
			batchL1txs = append(batchL1txs, &tx)
		case TypeNewBatch:
			l1txs = append(l1txs, batchL1txs)
			coordinatorL1txs = append(coordinatorL1txs, batchCoordinatorL1txs)
			l2txs = append(l2txs, batchL2txs)
			batchL1txs = []*common.L1Tx{}
			batchCoordinatorL1txs = []*common.L1Tx{}
			batchL2txs = []*common.PoolL2Tx{}
		default:
			continue
		}

	}
	l1txs = append(l1txs, batchL1txs)
	coordinatorL1txs = append(coordinatorL1txs, batchCoordinatorL1txs)
	l2txs = append(l2txs, batchL2txs)

	return l1txs, coordinatorL1txs, l2txs
}
