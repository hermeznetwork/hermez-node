package common

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewL1Tx(t *testing.T) {
	l1Tx := &L1Tx{
		ToForgeL1TxsNum: int64(123456),
		Position:        71,
		ToIdx:           301,
		TokenID:         5,
		Amount:          big.NewInt(1),
		LoadAmount:      big.NewInt(2),
		FromIdx:         300,
	}
	l1Tx, err := NewL1Tx(l1Tx)
	assert.Nil(t, err)
	assert.Equal(t, "0x01000000000001e240004700", l1Tx.TxID.String())
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

	encodedData, err := l1Tx.Bytes()
	require.Nil(t, err)
	assert.Equal(t, expected, encodedData)

	decodedData, err := L1TxFromBytes(encodedData)
	require.Nil(t, err)
	assert.Equal(t, l1Tx, decodedData)

	encodedData2, err := decodedData.Bytes()
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

func TestL1CoordinatorTxByteParsers(t *testing.T) {
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	require.Nil(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	require.Nil(t, err)
	fromEthAddr := crypto.PubkeyToAddress(*pubKey)
	var pkComp babyjub.PublicKeyComp
	err = pkComp.UnmarshalText([]byte("0x56ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c"))
	require.Nil(t, err)
	pk, err := pkComp.Decompress()
	require.Nil(t, err)
	bytesMessage1 := []byte("\x19Ethereum Signed Message:\n98")
	bytesMessage2 := []byte("I authorize this babyjubjub key for hermez rollup account creation")

	babyjub := pk.Compress()
	var data []byte
	data = append(data, bytesMessage1...)
	data = append(data, bytesMessage2...)
	data = append(data, babyjub[:]...)
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	require.Nil(t, err)
	// Ethereum adds 27 to v
	v := int(signature[64])
	signature[64] = byte(v + 27)

	l1Tx := &L1Tx{
		TokenID:     231,
		FromBJJ:     pk,
		FromEthAddr: fromEthAddr,
	}

	bytesCoordinatorL1, err := l1Tx.BytesCoordinatorTx(signature)
	require.Nil(t, err)
	l1txDecoded, err := L1TxFromCoordinatorBytes(bytesCoordinatorL1)
	require.Nil(t, err)
	assert.Equal(t, l1Tx, l1txDecoded)
}

func TestL1CoordinatorTxByteParsersCompatibility(t *testing.T) {
	var signature []byte
	r, err := hex.DecodeString("da71e5eb097e115405d84d1e7b464009b434b32c014a2df502d1f065ced8bc3b")
	require.Nil(t, err)
	s, err := hex.DecodeString("186d7122ff7f654cfed3156719774898d573900c86599a885a706dbdffe5ea8c")
	require.Nil(t, err)
	v, err := hex.DecodeString("1b")
	require.Nil(t, err)

	signature = append(signature, r[:]...)
	signature = append(signature, s[:]...)
	signature = append(signature, v[:]...)

	var pkComp babyjub.PublicKeyComp
	err = pkComp.UnmarshalText([]byte("0xa2c2807ee39c3b3378738cff85a46a9465bb8fcf44ea597c33da9719be7c259c"))
	require.Nil(t, err)
	// Data from the compatibility test
	expected := "1b186d7122ff7f654cfed3156719774898d573900c86599a885a706dbdffe5ea8cda71e5eb097e115405d84d1e7b464009b434b32c014a2df502d1f065ced8bc3ba2c2807ee39c3b3378738cff85a46a9465bb8fcf44ea597c33da9719be7c259c000000e7"
	pk, err := pkComp.Decompress()
	require.Nil(t, err)

	l1Tx := &L1Tx{
		TokenID: 231,
		FromBJJ: pk,
	}

	encodeData, err := l1Tx.BytesCoordinatorTx(signature)
	require.Nil(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodeData))
}
