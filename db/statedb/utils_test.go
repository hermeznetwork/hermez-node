package statedb

import (
	"io/ioutil"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIdx(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
	assert.NoError(t, err)

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

	tokenID0 := common.TokenID(0)
	tokenID1 := common.TokenID(1)

	// store the keys for idx by Addr & BJJ
	err = sdb.setIdxByEthAddrBJJ(idx, addr, pk.Compress(), tokenID0)
	require.NoError(t, err)

	idxR, err := sdb.GetIdxByEthAddrBJJ(addr, pk.Compress(), tokenID0)
	assert.NoError(t, err)
	assert.Equal(t, idx, idxR)

	// expect error when getting only by EthAddr, as value does not exist
	// in the db for only EthAddr
	_, err = sdb.GetIdxByEthAddr(addr, tokenID0)
	assert.NoError(t, err)
	_, err = sdb.GetIdxByEthAddr(addr2, tokenID0)
	assert.NotNil(t, err)
	// expect error when getting by EthAddr and BJJ, but for another TokenID
	_, err = sdb.GetIdxByEthAddrBJJ(addr, pk.Compress(), tokenID1)
	assert.NotNil(t, err)

	// expect to fail
	idxR, err = sdb.GetIdxByEthAddrBJJ(addr2, pk.Compress(), tokenID0)
	assert.NotNil(t, err)
	assert.Equal(t, common.Idx(0), idxR)
	idxR, err = sdb.GetIdxByEthAddrBJJ(addr, pk2.Compress(), tokenID0)
	assert.NotNil(t, err)
	assert.Equal(t, common.Idx(0), idxR)

	// try to store bigger idx, will not affect as already exist a smaller
	// Idx for that Addr & BJJ
	err = sdb.setIdxByEthAddrBJJ(idx2, addr, pk.Compress(), tokenID0)
	assert.NoError(t, err)

	// store smaller idx
	err = sdb.setIdxByEthAddrBJJ(idx3, addr, pk.Compress(), tokenID0)
	assert.NoError(t, err)

	idxR, err = sdb.GetIdxByEthAddrBJJ(addr, pk.Compress(), tokenID0)
	assert.NoError(t, err)
	assert.Equal(t, idx3, idxR)

	// by EthAddr should work
	idxR, err = sdb.GetIdxByEthAddr(addr, tokenID0)
	assert.NoError(t, err)
	assert.Equal(t, idx3, idxR)
	// expect error when trying to get Idx by addr2 & pk2
	idxR, err = sdb.GetIdxByEthAddrBJJ(addr2, pk2.Compress(), tokenID0)
	assert.NotNil(t, err)
	assert.Equal(t, ErrIdxNotFound, tracerr.Unwrap(err))
	assert.Equal(t, common.Idx(0), idxR)
	// expect error when trying to get Idx by addr with not used TokenID
	_, err = sdb.GetIdxByEthAddr(addr, tokenID1)
	assert.NotNil(t, err)

	sdb.Close()
}
