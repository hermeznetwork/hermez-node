package test

import (
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
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
func GenPoolTxs(n int, tokens []common.Token) []*common.PoolL2Tx {
	/*
		WARNING: this should be replaced by transaktio
		txs := make([]*common.PoolL2Tx, 0, n)
		privK := babyjub.NewRandPrivKey()
		for i := 256; i < 256+n; i++ {
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
			fee := common.FeeSelector(i % 255) //nolint:gomnd
			token := tokens[i%len(tokens)]
			tx := &common.PoolL2Tx{
				FromIdx:   common.Idx(i),
				ToIdx:     common.Idx(i + 1),
				ToEthAddr: ethCommon.BigToAddress(big.NewInt(int64(i))),
				ToBJJ:     privK.Public(),
				TokenID:   token.TokenID,
				Amount:    big.NewInt(int64(i)),
				Fee:       fee,
				Nonce:     common.Nonce(i),
				State:     state,
				Signature: privK.SignPoseidon(big.NewInt(int64(i))).Compress(),
			}
			var err error
			tx, err = common.NewPoolL2Tx(tx)
			if err != nil {
				panic(err)
			}
			if i%2 == 0 { // Optional parameters: rq
				tx.RqFromIdx = common.Idx(i)
				tx.RqToIdx = common.Idx(i + 1)
				tx.RqToEthAddr = ethCommon.BigToAddress(big.NewInt(int64(i)))
				tx.RqToBJJ = privK.Public()
				tx.RqTokenID = common.TokenID(i)
				tx.RqAmount = big.NewInt(int64(i))
				tx.RqFee = common.FeeSelector(i)
				tx.RqNonce = common.Nonce(i)
			}
			txs = append(txs, tx)
		}
		return txs
	*/
	return nil
}

// GenAuths generates account creation authorizations
func GenAuths(nAuths int) []*common.AccountCreationAuth {
	auths := []*common.AccountCreationAuth{}
	for i := 0; i < nAuths; i++ {
		// Generate keys
		ethPrivK, err := ethCrypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		bjjPrivK := babyjub.NewRandPrivKey()
		// Generate auth
		auth := &common.AccountCreationAuth{
			EthAddr: ethCrypto.PubkeyToAddress(ethPrivK.PublicKey),
			BJJ:     bjjPrivK.Public(),
		}
		// Sign
		h, err := auth.HashToSign()
		if err != nil {
			panic(err)
		}
		signature, err := ethCrypto.Sign(h, ethPrivK)
		if err != nil {
			panic(err)
		}
		auth.Signature = signature
		auths = append(auths, auth)
	}
	return auths
}
