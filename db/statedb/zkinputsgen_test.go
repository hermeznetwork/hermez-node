package statedb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:deadcode,unused
func printZKInputs(t *testing.T, zki *common.ZKInputs) {
	s, err := json.Marshal(zki)
	require.Nil(t, err)
	h, err := zki.HashGlobalData()
	require.Nil(t, err)

	fmt.Println("\nCopy&Paste into js circom test:\n	let zkInput = JSON.parse(`" + string(s) + "`);")
	// fmt.Println("\nZKInputs json:\n	echo '" + string(s) + "' | jq")

	fmt.Printf(`
		const output={
			hashGlobalInputs: "%s",
		};
		await circuit.assertOut(w, output);
		`, h.String())
	fmt.Println("")
}

func generateJsUsers(t *testing.T) []til.User {
	// same values than in the js test
	// skJsHex is equivalent to the 0000...000i js private key in commonjs
	skJsHex := []string{"7eb258e61862aae75c6c1d1f7efae5006ffc9e4d5596a6ff95f3df4ea209ea7f", "c005700f76f4b4cec710805c21595688648524df0a9d467afae537b7a7118819", "b373d14c67fb2a517bf4ac831c93341eec8e1b38dbc14e7d725b292a7cf84707", "2064b68d04a7aaae0ac3b36bf6f1850b380f1423be94a506c531940bd4a48b76"}
	addrHex := []string{"0x7e5f4552091a69125d5dfcb7b8c2659029395bdf", "0x2b5ad5c4795c026514f8317c7a215e218dccd6cf", "0x6813eb9362372eef6200f3b1dbc3f819671cba69", "0x1eff47bc3a10a45d4b230b5d10e37751fe6aa718"}
	var users []til.User
	for i := 0; i < len(skJsHex); i++ {
		skJs, err := hex.DecodeString(skJsHex[i])
		require.Nil(t, err)
		var sk babyjub.PrivateKey
		copy(sk[:], skJs)
		// bjj := sk.Public()
		user := til.User{
			Name: strconv.Itoa(i),
			BJJ:  &sk,
			Addr: ethCommon.HexToAddress(addrHex[i]),
		}
		users = append(users, user)
	}
	assert.Equal(t, "d746824f7d0ac5044a573f51b278acb56d823bec39551d1d7bf7378b68a1b021", users[0].BJJ.Public().String())
	assert.Equal(t, "4d05c307400c65795f02db96b1b81c60386fd53e947d9d3f749f3d99b1853909", users[1].BJJ.Public().String())
	assert.Equal(t, "38ffa002724562eb2a952a2503e206248962406cf16392ff32759b6f2a41fe11", users[2].BJJ.Public().String())
	assert.Equal(t, "c719e6401190be7fa7fbfcd3448fe2755233c01575341a3b09edadf5454f760b", users[3].BJJ.Public().String())

	return users
}

func signL2Tx(t *testing.T, user til.User, l2Tx common.PoolL2Tx) common.PoolL2Tx {
	toSign, err := l2Tx.HashToSign()
	require.Nil(t, err)
	sig := user.BJJ.SignPoseidon(toSign)
	l2Tx.Signature = sig.Compress()
	return l2Tx
}

func TestZKInputsHashTestVector0(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// same values than in the js test
	users := generateJsUsers(t)
	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
	}
	// skip first batch to do the test with BatchNum=1
	_, err = sdb.ProcessTxs(ptc, nil, nil, nil, nil)
	require.Nil(t, err)

	ptOut, err := sdb.ProcessTxs(ptc, nil, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, users[0].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[0].Addr.Hex(), acc.EthAddr.Hex())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(257))
	require.NotNil(t, err)
	ptOut.ZKInputs.FeeIdxs[0] = common.Idx(256).BigInt()

	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	assert.Nil(t, err)
	// value from js test vector
	expectedToHash := "0000000000ff000000000100000000000000000000000000000000000000000000000000000000000000000015ba488d749f6b891d29d0bf3a72481ec812e4d4ecef2bf7a3fc64f3c010444200000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a00000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000010003e87e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000001"
	// checks are splitted to find the difference easier
	assert.Equal(t, expectedToHash[:1000], hex.EncodeToString(toHash)[:1000])
	assert.Equal(t, expectedToHash[1000:2000], hex.EncodeToString(toHash)[1000:2000])
	assert.Equal(t, expectedToHash[2000:], hex.EncodeToString(toHash)[2000:])

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	// value from js test vector
	assert.Equal(t, "4356692423721763303547321618014315464040324829724049399065961225345730555597", h.String())
}

func TestZKInputsHashTestVector1(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// same values than in the js test
	users := generateJsUsers(t)
	// bjj0, err := common.BJJFromStringWithChecksum("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7")
	// assert.Nil(t, err)
	// bjj1, err := common.BJJFromStringWithChecksum("093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d")
	// assert.Nil(t, err)
	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 257,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     137,
			Type:    common.TxTypeTransfer,
		},
	}

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
	}
	// skip first batch to do the test with BatchNum=1
	_, err = sdb.ProcessTxs(ptc, nil, nil, nil, nil)
	require.Nil(t, err)

	ptOut, err := sdb.ProcessTxs(ptc, nil, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, users[0].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[0].Addr.Hex(), acc.EthAddr.Hex())
	acc, err = sdb.GetAccount(common.Idx(257))
	require.Nil(t, err)
	assert.Equal(t, users[1].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[1].Addr.Hex(), acc.EthAddr.Hex())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(258))
	require.NotNil(t, err)
	ptOut.ZKInputs.FeeIdxs[0] = common.Idx(257).BigInt()

	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	assert.Nil(t, err)
	// value from js test vector
	expectedToHash := "0000000000ff0000000001010000000000000000000000000000000000000000000000000000000000000000304a3f3aef4f416cca887aab7265227449077627138345c2eb25bf8ff946b09500000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a00000000000010000000000002b5ad5c4795c026514f8317c7a215e218dccd6cf093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d00000000000028a000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001010000010003e889000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010100000000000000000000000000000000000000000000000000000000000000000001"
	// checks are splitted to find the difference easier
	assert.Equal(t, expectedToHash[:1000], hex.EncodeToString(toHash)[:1000])
	assert.Equal(t, expectedToHash[1000:2000], hex.EncodeToString(toHash)[1000:2000])
	assert.Equal(t, expectedToHash[2000:], hex.EncodeToString(toHash)[2000:])

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	// value from js test vector
	assert.Equal(t, "20293112365009290386650039345314592436395562810005523677125576447132206192598", h.String())
}

// TestZKInputs0:
// - 1 L1Tx
// - 1 L2Tx without fees
// no Coordinator Idxs defined to receive the fees
func TestZKInputs0(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	nLevels := 16

	sdb, err := NewStateDB(dir, TypeBatchBuilder, nLevels)
	assert.Nil(t, err)

	// same values than in the js test
	users := generateJsUsers(t)

	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     0,
			Type:    common.TxTypeTransfer,
		},
	}

	toSign, err := l2Txs[0].HashToSign()
	require.Nil(t, err)
	sig := users[0].BJJ.SignPoseidon(toSign)
	l2Txs[0].Signature = sig.Compress()

	ptc := ProcessTxsConfig{
		NLevels:  uint32(nLevels),
		MaxTx:    3,
		MaxL1Tx:  2,
		MaxFeeTx: 2,
	}
	// skip first batch to do the test with BatchNum=1
	_, err = sdb.ProcessTxs(ptc, nil, nil, nil, nil)
	require.Nil(t, err)

	ptOut, err := sdb.ProcessTxs(ptc, nil, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, users[0].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[0].Addr.Hex(), acc.EthAddr.Hex())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(257))
	require.NotNil(t, err)

	// printZKInputs(t, ptOut.ZKInputs)
	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	require.Nil(t, err)
	assert.Equal(t, "0000000000ff0000000001000000000000000000000000000000000000000000000000000000000000000000071a61ed5a1ac052b0d1086a330c540b55318a07f6b7989573b9bbbb5380d1a900000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a0000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100010003e8000000000000000000000000000000000001", hex.EncodeToString(toHash))

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	assert.Equal(t, "10273997725311869157325593477103834352520120955255334791164183491223442653056", h.String())

	s, err := json.Marshal(ptOut.ZKInputs)
	require.Nil(t, err)

	// the 'expected' data has been checked with the circom circuits
	expected := `{"auxFromIdx":["256","0","0"],"auxToIdx":["0","0","0"],"ay1":["15238403086306505038849621710779816852318505119327426213168494964113886299863","15238403086306505038849621710779816852318505119327426213168494964113886299863","0"],"ay2":["0","15238403086306505038849621710779816852318505119327426213168494964113886299863","0"],"ay3":["0","0"],"balance1":["16000000","16000000","0"],"balance2":["0","15999000","0"],"balance3":["0","0"],"currentNumBatch":"1","ethAddr1":["721457446580647751014191829380889690493307935711","721457446580647751014191829380889690493307935711","0"],"ethAddr2":["0","721457446580647751014191829380889690493307935711","0"],"ethAddr3":["0","0"],"feeIdxs":["0","0"],"feePlanTokens":["0","0"],"fromBjjCompressed":[["1","1","1","0","1","0","1","1","0","1","1","0","0","0","1","0","0","1","0","0","0","0","0","1","1","1","1","1","0","0","1","0","1","0","1","1","1","1","1","0","0","1","0","1","0","0","0","0","1","0","1","0","0","0","1","1","0","0","1","0","0","0","0","0","0","1","0","1","0","0","1","0","1","1","1","0","1","0","1","0","1","1","1","1","1","1","0","0","1","0","0","0","1","0","1","0","0","1","0","0","1","1","0","1","0","0","0","1","1","1","1","0","0","0","1","1","0","1","0","1","1","0","1","0","1","1","0","1","1","0","1","1","0","1","1","0","0","1","0","0","0","0","0","1","1","1","0","1","1","1","0","0","0","0","1","1","0","1","1","1","1","0","0","1","1","1","0","0","1","0","1","0","1","0","1","0","1","0","1","1","1","0","0","0","1","0","1","1","1","0","0","0","1","1","0","1","1","1","1","0","1","1","1","0","1","1","1","1","1","1","1","0","1","1","0","0","1","1","0","1","0","0","0","1","0","0","0","1","0","1","1","0","1","0","0","0","0","1","0","1","0","0","0","0","1","1","0","1","1","0","0","0","0","1","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"fromEthAddr":["721457446580647751014191829380889690493307935711","0","0"],"fromIdx":["0","256","0"],"globalChainID":"0","imAccFeeOut":[["0","0"],["0","0"]],"imExitRoot":["0","0"],"imFinalAccFee":["0","0"],"imInitStateRootFee":"3212803832159212591526550848126062808026208063555125878245901046146545013161","imOnChain":["1","0"],"imOutIdx":["256","256"],"imStateRoot":["2999178063326948609414231200730958862089790119006655219527433501846141543551","3212803832159212591526550848126062808026208063555125878245901046146545013161"],"imStateRootFee":["3212803832159212591526550848126062808026208063555125878245901046146545013161"],"isOld0_1":["1","0","0"],"isOld0_2":["0","0","0"],"loadAmountF":["10400","0","0"],"maxNumBatch":["0","0","0"],"newAccount":["1","0","0"],"newExit":["0","0","0"],"nonce1":["0","0","0"],"nonce2":["0","1","0"],"nonce3":["0","0"],"oldKey1":["0","0","0"],"oldKey2":["0","0","0"],"oldLastIdx":"255","oldStateRoot":"0","oldValue1":["0","0","0"],"oldValue2":["0","0","0"],"onChain":["1","0","0"],"r8x":["0","13339118088097183560380359255316479838355724395928453439485234854234470298884","0"],"r8y":["0","12062876403986777372637801733000285846673058725183957648593976028822138986587","0"],"rqOffset":["0","0","0"],"rqToBjjAy":["0","0","0"],"rqToEthAddr":["0","0","0"],"rqTxCompressedDataV2":["0","0","0"],"s":["0","1429292460142966038093363510339656828866419125109324886747095533117015974779","0"],"siblings1":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"siblings2":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"siblings3":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"sign1":["0","0","0"],"sign2":["0","0","0"],"sign3":["0","0"],"toBjjAy":["0","0","0"],"toEthAddr":["0","0","0"],"toIdx":["0","256","0"],"tokenID1":["1","1","0"],"tokenID2":["0","1","0"],"tokenID3":["0","0"],"txCompressedData":["1461501637330902918203684832716283019659255211535","1483802382529433561627630154640673862706524841487","3322668559"],"txCompressedDataV2":["0","5271525021049092038181634317484288","0"]}`
	assert.Equal(t, expected, string(s))
}

// TestZKInputs1:
// - 2 L1Tx of CreateAccountDeposit
// - 1 L2Tx with fees
func TestZKInputs1(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	nLevels := 16

	sdb, err := NewStateDB(dir, TypeBatchBuilder, nLevels)
	assert.Nil(t, err)

	// same values than in the js test
	users := generateJsUsers(t)

	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	toSign, err := l2Txs[0].HashToSign()
	require.Nil(t, err)
	sig := users[0].BJJ.SignPoseidon(toSign)
	l2Txs[0].Signature = sig.Compress()

	ptc := ProcessTxsConfig{
		NLevels:  uint32(nLevels),
		MaxTx:    3,
		MaxL1Tx:  2,
		MaxFeeTx: 2,
	}
	// skip first batch to do the test with BatchNum=1
	_, err = sdb.ProcessTxs(ptc, nil, nil, nil, nil)
	require.Nil(t, err)

	coordIdxs := []common.Idx{257}
	ptOut, err := sdb.ProcessTxs(ptc, coordIdxs, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, users[0].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[0].Addr.Hex(), acc.EthAddr.Hex())
	assert.Equal(t, "15999899", acc.Balance.String())
	acc, err = sdb.GetAccount(common.Idx(257))
	require.Nil(t, err)
	assert.Equal(t, users[1].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[1].Addr.Hex(), acc.EthAddr.Hex())
	assert.Equal(t, "16000101", acc.Balance.String())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(258))
	require.NotNil(t, err)

	// printZKInputs(t, ptOut.ZKInputs)

	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	require.Nil(t, err)
	assert.Equal(t, "0000000000ff0000000001010000000000000000000000000000000000000000000000000000000000000000036d607b790b93bb1768d5390803b5a4a1f77e46755f57900930b14454faf95c00000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a00000000000010000000000002b5ad5c4795c026514f8317c7a215e218dccd6cf093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d00000000000028a000000000000100000000000000000000000000000000000000000100010003e87e01010000000000000001", hex.EncodeToString(toHash))

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	assert.Equal(t, "11039366437749764484706691779656824178800407917434257418748834397596260744468", h.String())

	s, err := json.Marshal(ptOut.ZKInputs)
	require.Nil(t, err)

	// the 'expected' data has been checked with the circom circuits
	expected := `{"auxFromIdx":["256","257","0"],"auxToIdx":["0","0","0"],"ay1":["15238403086306505038849621710779816852318505119327426213168494964113886299863","4172448640254579435434214421479401747968866348490029667576411173067925161293","15238403086306505038849621710779816852318505119327426213168494964113886299863"],"ay2":["0","0","15238403086306505038849621710779816852318505119327426213168494964113886299863"],"ay3":["4172448640254579435434214421479401747968866348490029667576411173067925161293","0"],"balance1":["16000000","16000000","16000000"],"balance2":["0","0","15998899"],"balance3":["16000000","0"],"currentNumBatch":"1","ethAddr1":["721457446580647751014191829380889690493307935711","247512291986854564435551364600938690683113101007","721457446580647751014191829380889690493307935711"],"ethAddr2":["0","0","721457446580647751014191829380889690493307935711"],"ethAddr3":["247512291986854564435551364600938690683113101007","0"],"feeIdxs":["257","0"],"feePlanTokens":["1","0"],"fromBjjCompressed":[["1","1","1","0","1","0","1","1","0","1","1","0","0","0","1","0","0","1","0","0","0","0","0","1","1","1","1","1","0","0","1","0","1","0","1","1","1","1","1","0","0","1","0","1","0","0","0","0","1","0","1","0","0","0","1","1","0","0","1","0","0","0","0","0","0","1","0","1","0","0","1","0","1","1","1","0","1","0","1","0","1","1","1","1","1","1","0","0","1","0","0","0","1","0","1","0","0","1","0","0","1","1","0","1","0","0","0","1","1","1","1","0","0","0","1","1","0","1","0","1","1","0","1","0","1","1","0","1","1","0","1","1","0","1","1","0","0","1","0","0","0","0","0","1","1","1","0","1","1","1","0","0","0","0","1","1","0","1","1","1","1","0","0","1","1","1","0","0","1","0","1","0","1","0","1","0","1","0","1","1","1","0","0","0","1","0","1","1","1","0","0","0","1","1","0","1","1","1","1","0","1","1","1","0","1","1","1","1","1","1","1","0","1","1","0","0","1","1","0","1","0","0","0","1","0","0","0","1","0","1","1","0","1","0","0","0","0","1","0","1","0","0","0","0","1","1","0","1","1","0","0","0","0","1","0","0"],["1","0","1","1","0","0","1","0","1","0","1","0","0","0","0","0","1","1","0","0","0","0","1","1","1","1","1","0","0","0","0","0","0","0","0","0","0","0","1","0","0","0","1","1","0","0","0","0","1","0","1","0","0","1","1","0","1","0","0","1","1","1","1","0","1","1","1","1","1","0","1","0","0","1","0","0","0","0","0","0","1","1","0","1","1","0","1","1","0","1","1","0","1","0","0","1","1","0","0","0","1","1","0","1","0","0","0","1","1","1","0","1","0","0","1","1","1","0","0","0","0","0","0","0","0","1","1","0","0","0","0","1","1","1","0","0","1","1","1","1","0","1","1","0","1","0","1","0","1","0","1","1","0","1","1","1","1","1","0","0","0","0","1","0","1","0","0","1","1","0","1","1","1","1","1","0","1","0","1","1","1","0","0","1","1","1","1","1","1","1","0","0","0","0","1","0","1","1","1","0","1","1","1","1","1","0","0","1","1","0","1","1","1","1","0","0","1","0","0","1","1","0","0","1","1","0","0","0","1","1","0","1","1","0","1","0","0","0","0","1","1","0","0","1","1","1","0","0","1","0","0","1","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"fromEthAddr":["721457446580647751014191829380889690493307935711","247512291986854564435551364600938690683113101007","0"],"fromIdx":["0","0","256"],"globalChainID":"0","imAccFeeOut":[["0","0"],["0","0"]],"imExitRoot":["0","0"],"imFinalAccFee":["101","0"],"imInitStateRootFee":"10660728613879129016661596154319504485937170756181586060561759832613498905432","imOnChain":["1","1"],"imOutIdx":["256","257"],"imStateRoot":["2999178063326948609414231200730958862089790119006655219527433501846141543551","13160175861809095962915811919507877524206523306071085047160493107056995190544"],"imStateRootFee":["1550190772280924834409423240867892593473592863918771212295716656664630983004"],"isOld0_1":["1","0","0"],"isOld0_2":["0","0","0"],"loadAmountF":["10400","10400","0"],"maxNumBatch":["0","0","0"],"newAccount":["1","1","0"],"newExit":["0","0","0"],"nonce1":["0","0","0"],"nonce2":["0","0","1"],"nonce3":["0","0"],"oldKey1":["0","256","0"],"oldKey2":["0","0","0"],"oldLastIdx":"255","oldStateRoot":"0","oldValue1":["0","9733782510199048326382833205201407219982604211594942097825192094127807440165","0"],"oldValue2":["0","0","0"],"onChain":["1","1","0"],"r8x":["0","0","16795818329006817411605347777151783287113601795569690834743955502344582990059"],"r8y":["0","0","21153110871938011270027204820675751452345817312713349225139208384264949654114"],"rqOffset":["0","0","0"],"rqToBjjAy":["0","0","0"],"rqToEthAddr":["0","0","0"],"rqTxCompressedDataV2":["0","0","0"],"s":["0","0","1550352334297856444344240780544275542131334387040150478134835458055364079268"],"siblings1":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["6975841694765113541634698345295957238501610055097872059913911260522365532165","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"siblings2":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["6975841694765113541634698345295957238501610055097872059913911260522365532165","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"siblings3":[["9827704113668630072730115158977131501210702363656902211840117643154933433410","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"sign1":["0","0","0"],"sign2":["0","0","0"],"sign3":["0","0"],"toBjjAy":["0","0","0"],"toEthAddr":["0","0","0"],"toIdx":["0","0","256"],"tokenID1":["1","1","1"],"tokenID2":["0","0","1"],"tokenID3":["1","0"],"txCompressedData":["1461501637330902918203684832716283019659255211535","1461501637330902918203684832716283019659255211535","869620039695611037216780722449287736442401358123623346624340758723552783"],"txCompressedDataV2":["0","0","3089511010385631938450432878260044363267416251956459471104"]}`
	assert.Equal(t, expected, string(s))
}

// TestZKInputs2:
// - 4 L1Tx of CreateAccountDeposit
// - 2 L2Tx with fees
func TestZKInputs2(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	nLevels := 16

	sdb, err := NewStateDB(dir, TypeBatchBuilder, nLevels)
	assert.Nil(t, err)

	// same values than in the js test
	users := generateJsUsers(t)

	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[2].BJJ.Public(),
			FromEthAddr:   users[2].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[3].BJJ.Public(),
			FromEthAddr:   users[3].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   258,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
		{
			FromIdx: 256,
			ToIdx:   259,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   1,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, users[0], l2Txs[0])
	l2Txs[1] = signL2Tx(t, users[0], l2Txs[1])

	ptc := ProcessTxsConfig{
		NLevels:  uint32(nLevels),
		MaxTx:    10,
		MaxL1Tx:  4,
		MaxFeeTx: 2,
	}
	// skip first batch to do the test with BatchNum=1
	_, err = sdb.ProcessTxs(ptc, nil, nil, nil, nil)
	require.Nil(t, err)

	coordIdxs := []common.Idx{257}
	ptOut, err := sdb.ProcessTxs(ptc, coordIdxs, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, users[0].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[0].Addr.Hex(), acc.EthAddr.Hex())
	assert.Equal(t, "15997798", acc.Balance.String())
	acc, err = sdb.GetAccount(common.Idx(257))
	require.Nil(t, err)
	assert.Equal(t, users[1].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[1].Addr.Hex(), acc.EthAddr.Hex())
	assert.Equal(t, "16000202", acc.Balance.String())
	acc, err = sdb.GetAccount(common.Idx(258))
	require.Nil(t, err)
	assert.Equal(t, users[2].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[2].Addr.Hex(), acc.EthAddr.Hex())
	assert.Equal(t, "16001000", acc.Balance.String())
	acc, err = sdb.GetAccount(common.Idx(259))
	require.Nil(t, err)
	assert.Equal(t, users[3].BJJ.Public().Compress().String(), acc.PublicKey.Compress().String())
	assert.Equal(t, users[3].Addr.Hex(), acc.EthAddr.Hex())
	assert.Equal(t, "16001000", acc.Balance.String())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(260))
	require.NotNil(t, err)

	// printZKInputs(t, ptOut.ZKInputs)

	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	require.Nil(t, err)
	assert.Equal(t, "0000000000ff00000000010300000000000000000000000000000000000000000000000000000000000000000916786f85a645e04f50cb0304768196730037209eebf2625661f060a9b2496800000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a00000000000010000000000002b5ad5c4795c026514f8317c7a215e218dccd6cf093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d00000000000028a00000000000010000000000006813eb9362372eef6200f3b1dbc3f819671cba6911fe412a6f9b7532ff9263f16c4062892406e203252a952aeb62457202a0ff3800000000000028a00000000000010000000000001eff47bc3a10a45d4b230b5d10e37751fe6aa7180b764f45f5aded093b1a347515c0335275e28f44d3fcfba77fbe901140e619c700000000000028a0000000000001000000000000000000000000000000000000000000000000000000000000000000000100010203e87e0100010303e87e0000000000000000000000000000000000000000000000000000000001010000000000000001", hex.EncodeToString(toHash))

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	assert.Equal(t, "12526925720009671154604304301604163669184748177914925908444300255027382653822", h.String())

	s, err := json.Marshal(ptOut.ZKInputs)
	require.Nil(t, err)
	// the 'expected' data has been checked with the circom circuits
	expected := `{"auxFromIdx":["256","257","258","259","0","0","0","0","0","0"],"auxToIdx":["0","0","0","0","0","0","0","0","0","0"],"ay1":["15238403086306505038849621710779816852318505119327426213168494964113886299863","4172448640254579435434214421479401747968866348490029667576411173067925161293","8138547337953155637712635674046872799144091023272207542905444721067900862264","5184476412130556544090991668866402222418200486089222951317185404775371774407","15238403086306505038849621710779816852318505119327426213168494964113886299863","15238403086306505038849621710779816852318505119327426213168494964113886299863","0","0","0","0"],"ay2":["0","0","0","0","8138547337953155637712635674046872799144091023272207542905444721067900862264","5184476412130556544090991668866402222418200486089222951317185404775371774407","0","0","0","0"],"ay3":["4172448640254579435434214421479401747968866348490029667576411173067925161293","0"],"balance1":["16000000","16000000","16000000","16000000","16000000","15998899","0","0","0","0"],"balance2":["0","0","0","0","16000000","16000000","0","0","0","0"],"balance3":["16000000","0"],"currentNumBatch":"1","ethAddr1":["721457446580647751014191829380889690493307935711","247512291986854564435551364600938690683113101007","594179275863704165266696689399235767493667371625","176962662172908264953938498278848696642639144728","721457446580647751014191829380889690493307935711","721457446580647751014191829380889690493307935711","0","0","0","0"],"ethAddr2":["0","0","0","0","594179275863704165266696689399235767493667371625","176962662172908264953938498278848696642639144728","0","0","0","0"],"ethAddr3":["247512291986854564435551364600938690683113101007","0"],"feeIdxs":["257","0"],"feePlanTokens":["1","0"],"fromBjjCompressed":[["1","1","1","0","1","0","1","1","0","1","1","0","0","0","1","0","0","1","0","0","0","0","0","1","1","1","1","1","0","0","1","0","1","0","1","1","1","1","1","0","0","1","0","1","0","0","0","0","1","0","1","0","0","0","1","1","0","0","1","0","0","0","0","0","0","1","0","1","0","0","1","0","1","1","1","0","1","0","1","0","1","1","1","1","1","1","0","0","1","0","0","0","1","0","1","0","0","1","0","0","1","1","0","1","0","0","0","1","1","1","1","0","0","0","1","1","0","1","0","1","1","0","1","0","1","1","0","1","1","0","1","1","0","1","1","0","0","1","0","0","0","0","0","1","1","1","0","1","1","1","0","0","0","0","1","1","0","1","1","1","1","0","0","1","1","1","0","0","1","0","1","0","1","0","1","0","1","0","1","1","1","0","0","0","1","0","1","1","1","0","0","0","1","1","0","1","1","1","1","0","1","1","1","0","1","1","1","1","1","1","1","0","1","1","0","0","1","1","0","1","0","0","0","1","0","0","0","1","0","1","1","0","1","0","0","0","0","1","0","1","0","0","0","0","1","1","0","1","1","0","0","0","0","1","0","0"],["1","0","1","1","0","0","1","0","1","0","1","0","0","0","0","0","1","1","0","0","0","0","1","1","1","1","1","0","0","0","0","0","0","0","0","0","0","0","1","0","0","0","1","1","0","0","0","0","1","0","1","0","0","1","1","0","1","0","0","1","1","1","1","0","1","1","1","1","1","0","1","0","0","1","0","0","0","0","0","0","1","1","0","1","1","0","1","1","0","1","1","0","1","0","0","1","1","0","0","0","1","1","0","1","0","0","0","1","1","1","0","1","0","0","1","1","1","0","0","0","0","0","0","0","0","1","1","0","0","0","0","1","1","1","0","0","1","1","1","1","0","1","1","0","1","0","1","0","1","0","1","1","0","1","1","1","1","1","0","0","0","0","1","0","1","0","0","1","1","0","1","1","1","1","1","0","1","0","1","1","1","0","0","1","1","1","1","1","1","1","0","0","0","0","1","0","1","1","1","0","1","1","1","1","1","0","0","1","1","0","1","1","1","1","0","0","1","0","0","1","1","0","0","1","1","0","0","0","1","1","0","1","1","0","1","0","0","0","0","1","1","0","0","1","1","1","0","0","1","0","0","1","0","0","0","0"],["0","0","0","1","1","1","0","0","1","1","1","1","1","1","1","1","0","0","0","0","0","1","0","1","0","1","0","0","0","0","0","0","0","1","0","0","1","1","1","0","1","0","1","0","0","0","1","0","0","1","0","0","0","1","1","0","1","1","0","1","0","1","1","1","0","1","0","1","0","1","0","0","1","0","1","0","1","0","0","1","0","1","0","1","0","1","0","0","1","0","1","0","0","1","0","0","1","1","0","0","0","0","0","0","0","1","0","0","0","1","1","1","0","1","1","0","0","0","0","0","0","0","1","0","0","1","0","0","1","0","0","1","0","0","0","1","0","1","0","0","0","1","1","0","0","0","0","0","0","0","1","0","0","0","1","1","0","1","1","0","1","0","0","0","1","1","1","1","1","1","0","0","0","1","1","0","0","1","0","0","1","0","0","1","1","1","1","1","1","1","1","1","0","1","0","0","1","1","0","0","1","0","1","0","1","1","1","0","1","1","0","1","1","0","0","1","1","1","1","1","0","1","1","0","0","1","0","1","0","1","0","0","1","0","0","0","0","0","1","0","0","1","1","1","1","1","1","1","1","0","0","0","1","0","0","0"],["1","1","1","0","0","0","1","1","1","0","0","1","1","0","0","0","0","1","1","0","0","1","1","1","0","0","0","0","0","0","1","0","1","0","0","0","1","0","0","0","0","0","0","0","1","0","0","1","0","1","1","1","1","1","0","1","1","1","1","1","1","1","1","0","1","1","1","0","0","1","0","1","1","1","0","1","1","1","1","1","0","0","1","1","1","1","1","1","1","1","0","0","1","0","1","1","0","0","1","0","0","0","1","0","1","1","1","1","0","0","0","1","0","1","0","0","0","1","1","1","1","0","1","0","1","1","1","0","0","1","0","0","1","0","1","0","1","1","0","0","1","1","0","0","0","0","0","0","0","0","1","1","1","0","1","0","1","0","0","0","1","0","1","0","1","1","1","0","0","0","1","0","1","1","0","0","0","1","0","1","1","0","0","0","1","1","0","1","1","1","0","0","1","0","0","1","0","0","0","0","1","0","1","1","0","1","1","1","1","0","1","1","0","1","0","1","1","0","1","0","1","1","1","1","1","0","1","0","0","0","1","0","1","1","1","1","0","0","1","0","0","1","1","0","1","1","1","0","1","1","0","1","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"fromEthAddr":["721457446580647751014191829380889690493307935711","247512291986854564435551364600938690683113101007","594179275863704165266696689399235767493667371625","176962662172908264953938498278848696642639144728","0","0","0","0","0","0"],"fromIdx":["0","0","0","0","256","256","0","0","0","0"],"globalChainID":"0","imAccFeeOut":[["0","0"],["0","0"],["0","0"],["0","0"],["101","0"],["202","0"],["202","0"],["202","0"],["202","0"]],"imExitRoot":["0","0","0","0","0","0","0","0","0"],"imFinalAccFee":["202","0"],"imInitStateRootFee":"11102416505971276127293972570397361714598994868311229974249826487060684688197","imOnChain":["1","1","1","1","0","0","0","0","0"],"imOutIdx":["256","257","258","259","259","259","259","259","259"],"imStateRoot":["2999178063326948609414231200730958862089790119006655219527433501846141543551","13160175861809095962915811919507877524206523306071085047160493107056995190544","6981081575125964439064848745251852976562872673661163359815312764236115200222","12287200745694319621025881114682538954756900570406181328753188227270350941803","10250910571353243150359059549936965649098556707015028714542118863411160731635","11102416505971276127293972570397361714598994868311229974249826487060684688197","11102416505971276127293972570397361714598994868311229974249826487060684688197","11102416505971276127293972570397361714598994868311229974249826487060684688197","11102416505971276127293972570397361714598994868311229974249826487060684688197"],"imStateRootFee":["4110517488865152389773850641074094192962655219669746287253968903130252986728"],"isOld0_1":["1","0","0","0","0","0","0","0","0","0"],"isOld0_2":["0","0","0","0","0","0","0","0","0","0"],"loadAmountF":["10400","10400","10400","10400","0","0","0","0","0","0"],"maxNumBatch":["0","0","0","0","0","0","0","0","0","0"],"newAccount":["1","1","1","1","0","0","0","0","0","0"],"newExit":["0","0","0","0","0","0","0","0","0","0"],"nonce1":["0","0","0","0","0","1","0","0","0","0"],"nonce2":["0","0","0","0","0","0","0","0","0","0"],"nonce3":["0","0"],"oldKey1":["0","256","256","257","0","0","0","0","0","0"],"oldKey2":["0","0","0","0","0","0","0","0","0","0"],"oldLastIdx":"255","oldStateRoot":"0","oldValue1":["0","9733782510199048326382833205201407219982604211594942097825192094127807440165","9733782510199048326382833205201407219982604211594942097825192094127807440165","17760137544353091380819593442189011627672167337464348873007461140156851422987","0","0","0","0","0","0"],"oldValue2":["0","0","0","0","0","0","0","0","0","0"],"onChain":["1","1","1","1","0","0","0","0","0","0"],"r8x":["0","0","0","0","5214090684236684851573641059575591118676476987573336170365486786637312665141","17818752874555175254292826108469784891786151976165952117435881323396526148981","0","0","0","0"],"r8y":["0","0","0","0","3568441662719387930655444674281943536423464418783443425289988425360036364264","7870851773538784486620881929410103548804126936780372631545590676485414360508","0","0","0","0"],"rqOffset":["0","0","0","0","0","0","0","0","0","0"],"rqToBjjAy":["0","0","0","0","0","0","0","0","0","0"],"rqToEthAddr":["0","0","0","0","0","0","0","0","0","0"],"rqTxCompressedDataV2":["0","0","0","0","0","0","0","0","0","0"],"s":["0","0","0","0","974535875596637445705358386750029200510091659451458393452031111972389260285","383599574514235269420771702006588662767158113271594446015439990815783784544","0","0","0","0"],"siblings1":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["6975841694765113541634698345295957238501610055097872059913911260522365532165","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["16340650608214686026509819657637793341613276710241727968165631912246770201551","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["20096881635736058727699182872953735227378445790497970658442161994753383780529","16179075207476077480163273158757030768504935550886919319708484674028389406807","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["20096881635736058727699182872953735227378445790497970658442161994753383780529","9666104145689704191764010810632978949926547298995332970899083004213180315751","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"siblings2":[["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["20096881635736058727699182872953735227378445790497970658442161994753383780529","18206494837933532633746175540777421367331523385545468990813257189787137445167","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["9141974547200611535034738074275487715083927494193481815706633586422651686652","6975841694765113541634698345295957238501610055097872059913911260522365532165","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"siblings3":[["9141974547200611535034738074275487715083927494193481815706633586422651686652","3300045819699360541652863975392028819717279937342922306336031453914844695187","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"],["0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0","0"]],"sign1":["0","0","0","0","0","0","0","0","0","0"],"sign2":["0","0","0","0","0","0","0","0","0","0"],"sign3":["0","0"],"toBjjAy":["0","0","0","0","0","0","0","0","0","0"],"toEthAddr":["0","0","0","0","0","0","0","0","0","0"],"toIdx":["0","0","0","0","258","259","0","0","0","0"],"tokenID1":["1","1","1","1","1","1","0","0","0","0"],"tokenID2":["0","0","0","0","1","1","0","0","0","0"],"tokenID3":["1","0"],"txCompressedData":["1461501637330902918203684832716283019659255211535","1461501637330902918203684832716283019659255211535","1461501637330902918203684832716283019659255211535","1461501637330902918203684832716283019659255211535","869620039695611037216780722449287736442401516579948375153015945811453455","869620039695617314318516109130051572231824803474526991772798003389916687","3322668559","3322668559","3322668559","3322668559"],"txCompressedDataV2":["0","0","0","0","3089511010385631938450432878260044363267416814906412892416","3089511010385654239195631408883185898985689744742895583488","0","0","0","0"]}`
	assert.Equal(t, expected, string(s))
}
