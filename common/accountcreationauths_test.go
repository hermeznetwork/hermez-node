package common

import (
	"encoding/hex"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCreationAuth(t *testing.T) {
	// Ethereum key
	ethSk, err := ethCrypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	require.NoError(t, err)
	ethAddr := ethCrypto.PubkeyToAddress(ethSk.PublicKey)

	// BabyJubJub key
	var sk babyjub.PrivateKey
	_, err = hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)

	chainID := uint16(0)
	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	a := AccountCreationAuth{
		EthAddr: ethAddr,
		BJJ:     sk.Public().Compress(),
	}
	msg, err := a.HashToSign(chainID, hermezContractAddr)
	assert.NoError(t, err)
	assert.Equal(t, "cb5a7e44329ff430c81fec49fb2ac6741f02d5ec96cbcb618a6991f0a9c80ffd", hex.EncodeToString(msg))

	// sign AccountCreationAuth with eth key
	sig, err := ethCrypto.Sign(msg, ethSk)
	assert.NoError(t, err)
	a.Signature = sig

	assert.True(t, a.VerifySignature(chainID, hermezContractAddr))
}
