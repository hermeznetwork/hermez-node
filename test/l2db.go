package test

import (
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jmoiron/sqlx"
)

// CleanL2DB deletes 'tx_pool' and 'account_creation_auth' from the given DB
func CleanL2DB(db *sqlx.DB) {
	if _, err := db.Exec("DELETE FROM tx_pool;"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("DELETE FROM account_creation_auth;"); err != nil {
		panic(err)
	}
}

// GenPoolTxs generates L2 pool txs.
// WARNING: This tx doesn't follow the protocol (signature, txID, ...)
// it's just to test getting/setting from/to the DB.
func GenPoolTxs(n int) []*common.PoolL2Tx {
	txs := make([]*common.PoolL2Tx, 0, n)
	privK := babyjub.NewRandPrivKey()
	for i := 0; i < n; i++ {
		var state common.PoolL2TxState
		//nolint:gomnd
		if i%4 == 0 {
			state = common.PoolL2TxStatePending
			//nolint:gomnd
		} else if i%4 == 1 {
			state = common.PoolL2TxStateInvalid
			//nolint:gomnd
		} else if i%4 == 2 {
			state = common.PoolL2TxStateForging
			//nolint:gomnd
		} else if i%4 == 3 {
			state = common.PoolL2TxStateForged
		}
		f := new(big.Float).SetInt(big.NewInt(int64(i)))
		amountF, _ := f.Float64()
		tx := &common.PoolL2Tx{
			TxID:        common.TxID(common.Hash([]byte(strconv.Itoa(i)))),
			FromIdx:     common.Idx(i),
			ToIdx:       common.Idx(i + 1),
			ToEthAddr:   ethCommon.BigToAddress(big.NewInt(int64(i))),
			ToBJJ:       privK.Public(),
			TokenID:     common.TokenID(i),
			Amount:      big.NewInt(int64(i)),
			AmountFloat: amountF,
			//nolint:gomnd
			Fee:       common.FeeSelector(i % 255),
			Nonce:     common.Nonce(i),
			State:     state,
			Signature: privK.SignPoseidon(big.NewInt(int64(i))),
			Timestamp: time.Now().UTC(),
		}
		if i%2 == 0 { // Optional parameters: rq
			tx.RqFromIdx = common.Idx(i)
			tx.RqToIdx = common.Idx(i + 1)
			tx.RqToEthAddr = ethCommon.BigToAddress(big.NewInt(int64(i)))
			tx.RqToBJJ = privK.Public()
			tx.RqTokenID = common.TokenID(i)
			tx.RqAmount = big.NewInt(int64(i))
			tx.RqFee = common.FeeSelector(i)
			tx.RqNonce = uint64(i)
		}
		if i%3 == 0 { // Optional parameters: things that get updated "a posteriori"
			tx.BatchNum = 489
			tx.AbsoluteFee = 39.12345
			tx.AbsoluteFeeUpdate = time.Now().UTC()
		}
		txs = append(txs, tx)
	}
	return txs
}

// GenAuths generates account creation authorizations
func GenAuths(nAuths int) []*common.AccountCreationAuth {
	auths := []*common.AccountCreationAuth{}
	for i := 0; i < nAuths; i++ {
		privK := babyjub.NewRandPrivKey()
		auths = append(auths, &common.AccountCreationAuth{
			EthAddr:   ethCommon.BigToAddress(big.NewInt(int64(i))),
			BJJ:       privK.Public(),
			Signature: []byte(strconv.Itoa(i)),
			Timestamp: time.Now(),
		})
	}
	return auths
}
