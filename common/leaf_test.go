package common

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	cryptoConstants "github.com/iden3/go-iden3-crypto/constants"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/assert"
)

func TestLeaf(t *testing.T) {
	leaf := &Leaf{
		TokenID: TokenID(1),
		Nonce:   uint64(1234),
		Balance: big.NewInt(1000),
		Ax:      big.NewInt(9876),
		Ay:      big.NewInt(6789),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	b, err := leaf.Bytes()
	assert.Nil(t, err)
	l1, err := LeafFromBytes(b)
	assert.Nil(t, err)
	assert.Equal(t, leaf, l1)

	e, err := leaf.BigInts()
	assert.Nil(t, err)
	assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[4]))

	assert.Equal(t, "1000", e[1].String())
	assert.Equal(t, "9876", e[2].String())
	assert.Equal(t, "6789", e[3].String())

	l2, err := LeafFromBigInts(e)
	assert.Nil(t, err)
	assert.Equal(t, leaf, l2)
	assert.Equal(t, l1, l2)
}

func TestLeafLoop(t *testing.T) {
	// check that for different Address there is no problem
	for i := 0; i < 256; i++ {
		key, err := ethCrypto.GenerateKey()
		assert.Nil(t, err)
		address := ethCrypto.PubkeyToAddress(key.PublicKey)

		leaf := &Leaf{
			TokenID: TokenID(i),
			Nonce:   uint64(i),
			Balance: big.NewInt(1000),
			Ax:      big.NewInt(9876),
			Ay:      big.NewInt(6789),
			EthAddr: address,
		}
		b, err := leaf.Bytes()
		assert.Nil(t, err)
		l1, err := LeafFromBytes(b)
		assert.Nil(t, err)
		assert.Equal(t, leaf, l1)

		e, err := leaf.BigInts()
		assert.Nil(t, err)
		assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[4]))

		l2, err := LeafFromBigInts(e)
		assert.Nil(t, err)
		assert.Equal(t, leaf, l2)
	}
}

func TestLeafErrNotInFF(t *testing.T) {
	z := big.NewInt(0)

	// Q-1 should not give error
	r := new(big.Int).Sub(cryptoConstants.Q, big.NewInt(1))
	e := [5]*big.Int{z, z, r, r, r}
	_, err := LeafFromBigInts(e)
	assert.Nil(t, err)

	// Q should give error
	r = cryptoConstants.Q
	e = [5]*big.Int{z, z, r, r, r}
	_, err = LeafFromBigInts(e)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotInFF, err)

	// Q+1 should give error
	r = new(big.Int).Add(cryptoConstants.Q, big.NewInt(1))
	e = [5]*big.Int{z, z, r, r, r}
	_, err = LeafFromBigInts(e)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotInFF, err)
}

func TestLeafErrNumOverflowNonce(t *testing.T) {
	// check limit
	leaf := &Leaf{
		TokenID: TokenID(1),
		Nonce:   uint64(math.Pow(2, 40) - 1),
		Balance: big.NewInt(1000),
		Ax:      big.NewInt(9876),
		Ay:      big.NewInt(6789),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	_, err := leaf.Bytes()
	assert.Nil(t, err)

	// force value overflow
	leaf.Nonce = uint64(math.Pow(2, 40))
	b, err := leaf.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Nonce", ErrNumOverflow), err)

	_, err = LeafFromBytes(b)
	assert.Nil(t, err)

	b[9] = 1
	_, err = LeafFromBytes(b)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Nonce", ErrNumOverflow), err)
}

func TestLeafErrNumOverflowBalance(t *testing.T) {
	// check limit
	leaf := &Leaf{
		TokenID: TokenID(1),
		Nonce:   uint64(math.Pow(2, 40) - 1),
		Balance: new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil), big.NewInt(1)),
		Ax:      big.NewInt(9876),
		Ay:      big.NewInt(6789),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512895", leaf.Balance.String())

	_, err := leaf.Bytes()
	assert.Nil(t, err)

	// force value overflow
	leaf.Balance = new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil)
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512896", leaf.Balance.String())
	b, err := leaf.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Balance", ErrNumOverflow), err)

	_, err = LeafFromBytes(b)
	assert.Nil(t, err)

	b[56] = 1
	_, err = LeafFromBytes(b)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Balance", ErrNumOverflow), err)
}
