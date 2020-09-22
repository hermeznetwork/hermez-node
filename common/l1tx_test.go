package common

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
	for f := 0; f < 256; f++ {
		perc := 0.0
		if f == 0 {
			perc = 0
			//nolint:gomnd
		} else if f <= 32 { //nolint:gomnd
			perc = math.Pow(10, -24+(float64(f)/2)) //nolint:gomnd
		} else if f <= 223 { //nolint:gomnd
			perc = math.Pow(10, -8+(0.041666666666667*(float64(f)-32))) //nolint:gomnd
		} else {
			perc = math.Pow(10, float64(f)-224) //nolint:gomnd
		}
		fmt.Printf("WHEN $1 = %d THEN %e\n", f, perc)
	}
}

func TestL1TxByteParsers(t *testing.T) {
	var pkComp babyjub.PublicKeyComp
	err := pkComp.UnmarshalText([]byte("0x56ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c"))
	require.Nil(t, err)

	pk, err := pkComp.Decompress()
	require.Nil(t, err)

	l1Tx := &L1Tx{
		ToIdx:       3,
		TokenID:     5,
		Amount:      big.NewInt(1),
		LoadAmount:  big.NewInt(2),
		FromIdx:     2,
		FromBJJ:     pk,
		FromEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}

	expected, err := utils.HexDecode("c58d29fa6e86e4fae04ddced660d45bcf3cb237056ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c0000000000020002000100000005000000000003")
	require.Nil(t, err)

	encodedData, err := l1Tx.Bytes(32)
	require.Nil(t, err)
	assert.Equal(t, expected, encodedData)

	decodedData, err := L1TxFromBytes(encodedData)
	require.Nil(t, err)
	assert.Equal(t, l1Tx, decodedData)

	encodedData2, err := decodedData.Bytes(32)
	require.Nil(t, err)
	assert.Equal(t, encodedData, encodedData2)

	// expect error if length!=68
	_, err = L1TxFromBytes(encodedData[:66])
	require.NotNil(t, err)
	_, err = L1TxFromBytes([]byte{})
	require.NotNil(t, err)
	_, err = L1TxFromBytes(nil)
	require.NotNil(t, err)
}
