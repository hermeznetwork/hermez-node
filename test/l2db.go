package test

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

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
				Nonce:     nonce.Nonce(i),
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
				tx.RqNonce = nonce.Nonce(i)
			}
			txs = append(txs, tx)
		}
		return txs
	*/
	return nil
}

// GenAuths generates account creation authorizations
func GenAuths(nAuths int, chainID uint16,
	hermezContractAddr ethCommon.Address) []*common.AccountCreationAuth {
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
			BJJ:     bjjPrivK.Public().Compress(),
		}
		// Sign
		h, err := auth.HashToSign(chainID, hermezContractAddr)
		if err != nil {
			panic(err)
		}
		signature, err := ethCrypto.Sign(h, ethPrivK)
		if err != nil {
			panic(err)
		}
		signature[64] += 27
		auth.Signature = signature
		auths = append(auths, auth)
	}
	return auths
}
