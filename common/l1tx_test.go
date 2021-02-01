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

func TestNewL1UserTx(t *testing.T) {
	toForge := int64(123456)
	l1Tx := &L1Tx{
		ToForgeL1TxsNum: &toForge,
		Position:        71,
		UserOrigin:      true,
		ToIdx:           301,
		TokenID:         5,
		Amount:          big.NewInt(1),
		DepositAmount:   big.NewInt(2),
		FromIdx:         Idx(300),
	}
	l1Tx, err := NewL1Tx(l1Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x00a6cbae3b8661fb75b0919ca6605a02cfb04d9c6dd16870fa0fcdf01befa32768", l1Tx.TxID.String())
}

func TestNewL1CoordinatorTx(t *testing.T) {
	batchNum := BatchNum(51966)
	l1Tx := &L1Tx{
		Position:      88,
		UserOrigin:    false,
		ToIdx:         301,
		TokenID:       5,
		Amount:        big.NewInt(1),
		DepositAmount: big.NewInt(2),
		FromIdx:       Idx(300),
		BatchNum:      &batchNum,
	}
	l1Tx, err := NewL1Tx(l1Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x01274482d73df4dab34a1b6740adfca347a462513aa14e82f27b12f818d1b68c84", l1Tx.TxID.String())
}

func TestL1TxCompressedData(t *testing.T) {
	tx := L1Tx{
		FromIdx: 2,
		ToIdx:   3,
		Amount:  big.NewInt(4),
		TokenID: 5,
	}
	chainID := uint16(0)
	txCompressedData, err := tx.TxCompressedData(chainID)
	assert.NoError(t, err)

	// test vector value generated from javascript implementation
	expectedStr := "7307597389635308713748674793997299267459594577423"
	assert.Equal(t, expectedStr, txCompressedData.String())
	assert.Equal(t, "0500040000000000030000000000020000c60be60f", hex.EncodeToString(txCompressedData.Bytes()))
}

func TestBytesDataAvailability(t *testing.T) {
	tx := L1Tx{
		FromIdx: 2,
		ToIdx:   3,
		Amount:  big.NewInt(4),
		TokenID: 5,
	}
	txCompressedData, err := tx.BytesDataAvailability(32)
	assert.NoError(t, err)
	assert.Equal(t, "0000000200000003000000", hex.EncodeToString(txCompressedData))

	tx = L1Tx{
		FromIdx:         2,
		ToIdx:           3,
		EffectiveAmount: big.NewInt(4),
		TokenID:         5,
	}
	txCompressedData, err = tx.BytesDataAvailability(32)
	assert.NoError(t, err)
	assert.Equal(t, "0000000200000003000400", hex.EncodeToString(txCompressedData))
}

func TestL1TxFromDataAvailability(t *testing.T) {
	tx := L1Tx{
		FromIdx: 2,
		ToIdx:   3,
		Amount:  big.NewInt(4),
	}
	txCompressedData, err := tx.BytesDataAvailability(32)
	assert.NoError(t, err)
	l1tx, err := L1TxFromDataAvailability(txCompressedData, 32)
	require.NoError(t, err)
	assert.Equal(t, tx.FromIdx, l1tx.FromIdx)
	assert.Equal(t, tx.ToIdx, l1tx.ToIdx)

	tx = L1Tx{
		FromIdx:         2,
		ToIdx:           3,
		EffectiveAmount: big.NewInt(4),
	}
	txCompressedData, err = tx.BytesDataAvailability(32)
	assert.NoError(t, err)
	l1tx, err = L1TxFromDataAvailability(txCompressedData, 32)
	require.NoError(t, err)
	assert.Equal(t, tx.FromIdx, l1tx.FromIdx)
	assert.Equal(t, tx.ToIdx, l1tx.ToIdx)
	assert.Equal(t, tx.EffectiveAmount, l1tx.EffectiveAmount)
}

func TestL1userTxByteParsers(t *testing.T) {
	var pkComp babyjub.PublicKeyComp
	pkCompL := []byte("0x56ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c")
	err := pkComp.UnmarshalText(pkCompL)
	require.NoError(t, err)

	l1Tx := &L1Tx{
		UserOrigin:    true,
		ToIdx:         3,
		TokenID:       5,
		Amount:        big.NewInt(1),
		DepositAmount: big.NewInt(2),
		FromIdx:       2,
		FromBJJ:       pkComp,
		FromEthAddr:   ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}

	encodedData, err := l1Tx.BytesUser()
	require.NoError(t, err)
	decodedData, err := L1UserTxFromBytes(encodedData)
	require.NoError(t, err)
	assert.Equal(t, l1Tx, decodedData)
	encodedData2, err := decodedData.BytesUser()
	require.NoError(t, err)
	assert.Equal(t, encodedData, encodedData2)

	// expect error if length!=68
	_, err = L1UserTxFromBytes(encodedData[:66])
	require.NotNil(t, err)
	_, err = L1UserTxFromBytes([]byte{})
	require.NotNil(t, err)
	_, err = L1UserTxFromBytes(nil)
	require.NotNil(t, err)
}

func TestL1TxByteParsersCompatibility(t *testing.T) {
	// Data from compatibility test
	var pkComp babyjub.PublicKeyComp
	pkCompB, err := hex.DecodeString("0dd02deb2c81068e7a0f7e327df80b4ab79ee1f41a7def613e73a20c32eece5a")
	require.NoError(t, err)
	pkCompL := SwapEndianness(pkCompB)
	err = pkComp.UnmarshalText([]byte(hex.EncodeToString(pkCompL)))
	require.NoError(t, err)

	depositAmount := new(big.Int)
	depositAmount.SetString("100000000000000000000", 10)
	l1Tx := &L1Tx{
		ToIdx:         87865485,
		TokenID:       2098076,
		Amount:        big.NewInt(2400000000000000000),
		DepositAmount: depositAmount,
		FromIdx:       Idx(29767899),
		FromBJJ:       pkComp,
		FromEthAddr:   ethCommon.HexToAddress("0x85dab5b9e2e361d0c208d77be90efcc0439b0a53"),
		UserOrigin:    true,
	}

	expected, err := utils.HexDecode("85dab5b9e2e361d0c208d77be90efcc0439b0a530dd02deb2c81068e7a0f7e327df80b4ab79ee1f41a7def613e73a20c32eece5a000001c638db8be880f00020039c0000053cb88d")
	require.NoError(t, err)

	encodedData, err := l1Tx.BytesUser()
	require.NoError(t, err)
	assert.Equal(t, expected, encodedData)
}

func TestL1CoordinatorTxByteParsers(t *testing.T) {
	hermezAddress := ethCommon.HexToAddress("0xD6C850aeBFDC46D7F4c207e445cC0d6B0919BDBe")
	chainID := big.NewInt(1337)
	chainIDBytes := ethCommon.LeftPadBytes(chainID.Bytes(), 2)

	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	require.NoError(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	require.NoError(t, err)
	fromEthAddr := crypto.PubkeyToAddress(*pubKey)
	var pkComp babyjub.PublicKeyComp
	pkCompL := []byte("56ca90f80d7c374ae7485e9bcc47d4ac399460948da6aeeb899311097925a72c")
	err = pkComp.UnmarshalText(pkCompL)
	require.NoError(t, err)
	bytesMessage1 := []byte("\x19Ethereum Signed Message:\n120")
	bytesMessage2 := []byte("I authorize this babyjubjub key for hermez rollup account creation")

	babyjubB := SwapEndianness(pkComp[:])
	var data []byte
	data = append(data, bytesMessage1...)
	data = append(data, bytesMessage2...)
	data = append(data, babyjubB[:]...)
	data = append(data, chainIDBytes...)
	data = append(data, hermezAddress.Bytes()...)
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	require.NoError(t, err)
	// Ethereum adds 27 to v
	v := int(signature[64])
	signature[64] = byte(v + 27)

	l1Tx := &L1Tx{
		TokenID:       231,
		FromBJJ:       pkComp,
		FromEthAddr:   fromEthAddr,
		Amount:        big.NewInt(0),
		DepositAmount: big.NewInt(0),
	}

	bytesCoordinatorL1, err := l1Tx.BytesCoordinatorTx(signature)
	require.NoError(t, err)
	l1txDecoded, err := L1CoordinatorTxFromBytes(bytesCoordinatorL1, chainID, hermezAddress)
	require.NoError(t, err)
	assert.Equal(t, l1Tx, l1txDecoded)
	bytesCoordinatorL12, err := l1txDecoded.BytesCoordinatorTx(signature)
	require.NoError(t, err)
	assert.Equal(t, bytesCoordinatorL1, bytesCoordinatorL12)

	// expect error if length!=68
	_, err = L1CoordinatorTxFromBytes(bytesCoordinatorL1[:66], chainID, hermezAddress)
	require.NotNil(t, err)
	_, err = L1CoordinatorTxFromBytes([]byte{}, chainID, hermezAddress)
	require.NotNil(t, err)
	_, err = L1CoordinatorTxFromBytes(nil, chainID, hermezAddress)
	require.NotNil(t, err)
}

func TestL1CoordinatorTxByteParsersCompatibility(t *testing.T) {
	// Data from compatibility test
	var signature []byte
	r, err := hex.DecodeString("da71e5eb097e115405d84d1e7b464009b434b32c014a2df502d1f065ced8bc3b")
	require.NoError(t, err)
	s, err := hex.DecodeString("186d7122ff7f654cfed3156719774898d573900c86599a885a706dbdffe5ea8c")
	require.NoError(t, err)
	v, err := hex.DecodeString("1b")
	require.NoError(t, err)

	signature = append(signature, r[:]...)
	signature = append(signature, s[:]...)
	signature = append(signature, v[:]...)

	var pkComp babyjub.PublicKeyComp
	pkCompB, err := hex.DecodeString("a2c2807ee39c3b3378738cff85a46a9465bb8fcf44ea597c33da9719be7c259c")
	require.NoError(t, err)
	pkCompL := SwapEndianness(pkCompB)
	err = pkComp.UnmarshalText([]byte(hex.EncodeToString(pkCompL)))
	require.NoError(t, err)
	// Data from the compatibility test
	require.NoError(t, err)
	l1Tx := &L1Tx{
		TokenID: 231,
		FromBJJ: pkComp,
	}

	encodeData, err := l1Tx.BytesCoordinatorTx(signature)
	require.NoError(t, err)

	expected, err := utils.HexDecode("1b186d7122ff7f654cfed3156719774898d573900c86599a885a706dbdffe5ea8cda71e5eb097e115405d84d1e7b464009b434b32c014a2df502d1f065ced8bc3ba2c2807ee39c3b3378738cff85a46a9465bb8fcf44ea597c33da9719be7c259c000000e7")
	require.NoError(t, err)

	assert.Equal(t, expected, encodeData)
}

func TestL1TxID(t *testing.T) {
	// L1UserTx
	i64_1 := int64(1)
	i64_2 := int64(2)
	tx0 := L1Tx{
		UserOrigin:      true,
		ToForgeL1TxsNum: &i64_1,
		Position:        1,
	}
	err := tx0.SetID()
	require.NoError(t, err)
	assert.Equal(t, TxIDPrefixL1UserTx, tx0.TxID[0])

	// differ ToForgeL1TxsNum
	tx1 := L1Tx{
		UserOrigin:      true,
		ToForgeL1TxsNum: &i64_2,
		Position:        1,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)

	// differ Position
	tx1 = L1Tx{
		UserOrigin:      true,
		ToForgeL1TxsNum: &i64_1,
		Position:        2,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)

	// L1CoordinatorTx
	bn1 := BatchNum(1)
	bn2 := BatchNum(2)
	tx0 = L1Tx{
		UserOrigin: false,
		BatchNum:   &bn1,
		Position:   1,
	}
	err = tx0.SetID()
	require.NoError(t, err)
	assert.Equal(t, TxIDPrefixL1CoordTx, tx0.TxID[0])

	// differ BatchNum
	tx1 = L1Tx{
		UserOrigin: false,
		BatchNum:   &bn2,
		Position:   1,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)

	// differ Position
	tx1 = L1Tx{
		UserOrigin: false,
		BatchNum:   &bn1,
		Position:   2,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)
}
