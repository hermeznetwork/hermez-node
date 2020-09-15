package common

import (
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestL1TxCodec(t *testing.T) {
	var pkComp babyjub.PublicKeyComp
	err := pkComp.UnmarshalText([]byte("0x56ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c"))
	require.Nil(t, err)

	pk, err := pkComp.Decompress()
	require.Nil(t, err)

	l1Tx := L1Tx{
		ToIdx:       3,
		TokenID:     5,
		Amount:      big.NewInt(1),
		LoadAmount:  big.NewInt(2),
		FromIdx:     2,
		FromBJJ:     pk,
		FromEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}

	expected, err := utils.HexDecode("c58d29fa6e86e4fae04ddced660d45bcf3cb237056ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c00000002000200010000000500000003")
	require.Nil(t, err)

	encodedData := l1Tx.Bytes(32)
	assert.Equal(t, expected, encodedData)

	decodedData, err := L1TxFromBytes(encodedData)
	require.Nil(t, err)

	encodedData2 := decodedData.Bytes(32)
	assert.Equal(t, encodedData, encodedData2)
}
