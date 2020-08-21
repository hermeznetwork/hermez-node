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
func GenerateTestTxs(t *testing.T, instructions Instructions) ([]*common.L1Tx, []*common.PoolL2Tx) {
	accounts := GenerateKeys(t, instructions.Accounts)

	// debug
	// fmt.Println("accounts:")
	// for n, a := range accounts {
	//         fmt.Printf("	%s: bjj:%s - addr:%s\n", n, a.BJJ.Public().String()[:10], a.Addr.Hex()[:10])
	// }

	var l1txs []*common.L1Tx
	var l2txs []*common.PoolL2Tx
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
			l1txs = append(l1txs, &tx)
			if accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx == common.Idx(0) { // if account.Idx is not set yet, set it and increment idx
				accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx = common.Idx(idx)
				idx++
			}
		case common.TxTypeTransfer:
			tx := common.PoolL2Tx{
				// TxID: nil,
				FromIdx:   accounts[idxTokenIDToString(inst.From, inst.TokenID)].Idx,
				ToIdx:     accounts[idxTokenIDToString(inst.To, inst.TokenID)].Idx,
				ToEthAddr: accounts[idxTokenIDToString(inst.To, inst.TokenID)].Addr,
				ToBJJ:     accounts[idxTokenIDToString(inst.To, inst.TokenID)].BJJ.Public(),
				TokenID:   inst.TokenID,
				Amount:    big.NewInt(int64(inst.Amount)),
				Fee:       common.FeeSelector(inst.Fee),
				Nonce:     accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce,
				State:     common.PoolL2TxStatePending,
				Timestamp: time.Now(),
				BatchNum:  0,
				Type:      common.TxTypeTransfer,
			}
			// TODO once signature function is ready, perform
			// signature and set it to tx.Signature

			accounts[idxTokenIDToString(inst.From, inst.TokenID)].Nonce++
			l2txs = append(l2txs, &tx)
		default:
			continue
		}

	}

	return l1txs, l2txs
}
