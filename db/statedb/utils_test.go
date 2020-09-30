package statedb

import (
	"io/ioutil"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIdx(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, false, 0)
	assert.Nil(t, err)

	var sk babyjub.PrivateKey
	copy(sk[:], []byte("1234")) // only for testing
	pk := sk.Public()
	var sk2 babyjub.PrivateKey
	copy(sk2[:], []byte("12345")) // only for testing
	pk2 := sk2.Public()
	addr := ethCommon.HexToAddress("0x74E803744B7EEFc272E852f89a05D41515d431f2")
	addr2 := ethCommon.HexToAddress("0x54A0706531cEa2ee8F09bAd22f604e377bb56948")
	idx := common.Idx(1234)
	idx2 := common.Idx(12345)
	idx3 := common.Idx(1233)

	// store the keys for idx by Addr & BJJ
	err = sdb.setIdxByEthAddrBJJ(idx, addr, pk)
	require.Nil(t, err)

	idxR, err := sdb.GetIdxByEthAddrBJJ(addr, pk)
	assert.Nil(t, err)
	assert.Equal(t, idx, idxR)

	// expect error when getting only by EthAddr, as value does not exist
	// in the db for only EthAddr
	_, err = sdb.GetIdxByEthAddr(addr)
	assert.Nil(t, err)
	_, err = sdb.GetIdxByEthAddr(addr2)
	assert.NotNil(t, err)

	// expect to fail
	idxR, err = sdb.GetIdxByEthAddrBJJ(addr2, pk)
	assert.NotNil(t, err)
	assert.Equal(t, common.Idx(0), idxR)
	idxR, err = sdb.GetIdxByEthAddrBJJ(addr, pk2)
	assert.NotNil(t, err)
	assert.Equal(t, common.Idx(0), idxR)

	// try to store bigger idx, will not affect as already exist a smaller
	// Idx for that Addr & BJJ
	err = sdb.setIdxByEthAddrBJJ(idx2, addr, pk)
	assert.Nil(t, err)

	// store smaller idx
	err = sdb.setIdxByEthAddrBJJ(idx3, addr, pk)
	assert.Nil(t, err)

	idxR, err = sdb.GetIdxByEthAddrBJJ(addr, pk)
	assert.Nil(t, err)
	assert.Equal(t, idx3, idxR)

	// by EthAddr should work
	idxR, err = sdb.GetIdxByEthAddr(addr)
	assert.Nil(t, err)
	assert.Equal(t, idx3, idxR)
	// expect error when trying to get Idx by addr2 & pk2
	idxR, err = sdb.GetIdxByEthAddrBJJ(addr2, pk2)
	assert.NotNil(t, err)
	assert.Equal(t, ErrToIdxNotFound, err)
	assert.Equal(t, common.Idx(0), idxR)
}

func TestBJJCompressedTo256BigInt(t *testing.T) {
	var pkComp babyjub.PublicKeyComp
	r := BJJCompressedTo256BigInts(pkComp)
	zero := big.NewInt(0)
	for i := 0; i < 256; i++ {
		assert.Equal(t, zero, r[i])
	}

	pkComp[0] = 3
	r = BJJCompressedTo256BigInts(pkComp)
	one := big.NewInt(1)
	for i := 0; i < 256; i++ {
		if i != 0 && i != 1 {
			assert.Equal(t, zero, r[i])
		} else {
			assert.Equal(t, one, r[i])
		}
	}

	pkComp[31] = 4
	r = BJJCompressedTo256BigInts(pkComp)
	for i := 0; i < 256; i++ {
		if i != 0 && i != 1 && i != 250 {
			assert.Equal(t, zero, r[i])
		} else {
			assert.Equal(t, one, r[i])
		}
	}
}
