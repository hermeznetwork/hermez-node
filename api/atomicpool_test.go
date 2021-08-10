package api

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicPool(t *testing.T) {
	// Generate N "wallets" (account + private key)
	const nAccounts = 4 // don't change this value
	const usedToken = 0 // this test will use only a token
	accounts := make([]common.Account, nAccounts)
	accountUpdates := make([]common.AccountUpdate, nAccounts)
	privateKeys := make(map[common.Idx]*babyjub.PrivateKey, nAccounts)
	for i := 0; i < nAccounts; i++ {
		addr, privKey := generateKeys(i + 1234567)
		idx := common.Idx(i) + 5000
		account := common.Account{
			Idx:      idx,
			TokenID:  tc.tokens[usedToken].TokenID,
			BatchNum: 1,
			BJJ:      privKey.Public().Compress(),
			EthAddr:  addr,
		}
		balance, ok := big.NewInt(0).SetString("1000000000000000000", 10)
		require.True(t, ok)
		accountUpdate := common.AccountUpdate{
			Idx:      idx,
			BatchNum: 1,
			Nonce:    0,
			Balance:  balance,
		}
		accounts[i] = account
		accountUpdates[i] = accountUpdate
		privateKeys[idx] = &privKey
	}
	// Add accounts to HistoryDB
	err := api.historyDB.AddAccounts(accounts)
	assert.NoError(t, err)
	err = api.historyDB.AddAccountUpdates(accountUpdates)
	assert.NoError(t, err)

	txsToClean := []common.TxID{}
	signAndTransformTxs := func(txs []common.PoolL2Tx) (common.AtomicGroup, []testPoolTxReceive) {
		for i := 0; i < len(txs); i++ {
			// Set TxID and type
			_, err := common.NewPoolL2Tx(&txs[i])
			txsToClean = append(txsToClean, txs[i].TxID)
			assert.NoError(t, err)
			// Sign
			toSign, err := txs[i].HashToSign(0)
			assert.NoError(t, err)
			sig := privateKeys[txs[i].FromIdx].SignPoseidon(toSign)
			txs[i].Signature = sig.Compress()
		}
		txsToSend, txsToReceive := genTestPoolTxs(txs, []historydb.TokenWithUSD{tc.tokens[usedToken]}, accounts)
		atomicGroup := common.AtomicGroup{Txs: txsToSend}
		atomicGroup.SetAtomicGroupID()
		return atomicGroup, txsToReceive
	}
	assertTxs := func(txsToReceive []testPoolTxReceive, atomicGroupID common.AtomicGroupID) {
		// Fetch txs one by one
		for _, tx := range txsToReceive {
			const path = apiURL + "transactions-pool/"
			fetchedTx := testPoolTxReceive{}
			require.NoError(
				t, doGoodReq(
					"GET",
					path+tx.TxID.String(),
					nil, &fetchedTx,
				),
			)
			assertPoolTx(t, tx, fetchedTx)
		}
		// Fetch all the group using GET /atomic-pool/{id}
		const path = apiURL + "atomic-pool/"
		fetchedTxs := []testPoolTxReceive{}
		require.NoError(
			t, doGoodReq(
				"GET",
				path+atomicGroupID.String(),
				nil, &fetchedTxs,
			),
		)
		assert.Equal(t, len(txsToReceive), len(fetchedTxs))
		for i, tx := range txsToReceive {
			assertPoolTx(t, tx, fetchedTxs[i])
		}
	}

	const path = apiURL + "atomic-pool"
	// Test correct atomic group (ciclic)
	/*
		A  ──────────► B
		▲              │
		│              │
		└────── C ◄────┘
	*/
	// Generate txs
	txs := []common.PoolL2Tx{}
	baseTx := common.PoolL2Tx{
		TokenID:     tc.tokens[usedToken].TokenID,
		Amount:      big.NewInt(10000000000),
		Fee:         200,
		Nonce:       0,
		RqTokenID:   tc.tokens[usedToken].TokenID,
		RqAmount:    big.NewInt(10000000000),
		RqFee:       200,
		RqNonce:     0,
		MaxNumBatch: 9999999,
	}
	for i := 0; i < nAccounts; i++ {
		tx := baseTx
		tx.FromIdx = accounts[i].Idx
		tx.ToIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqFromIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqToIdx = accounts[(i+2)%nAccounts].Idx
		if i != nAccounts-1 {
			tx.RqOffset = 1
		} else {
			tx.RqOffset = 5
		}
		txs = append(txs, tx)
	}
	// Sign and format txs
	atomicGroup, txsToReceive := signAndTransformTxs(txs)
	// Send txs
	jsonTxBytes, err := json.Marshal(atomicGroup)
	require.NoError(t, err)
	jsonTxReader := bytes.NewReader(jsonTxBytes)
	fetchedTxIDs := []common.TxID{}
	err = doGoodReq("POST", path, jsonTxReader, &fetchedTxIDs)
	assert.NoError(t, err)
	// Check response
	expectedTxIDs := []common.TxID{}
	for _, tx := range txs {
		expectedTxIDs = append(expectedTxIDs, tx.TxID)
	}
	assert.Equal(t, expectedTxIDs, fetchedTxIDs)
	// Check txs in the DB
	assertTxs(txsToReceive, atomicGroup.ID)

	// test that we can't update atomic tx
	// this test is checking, that this request will return error
	// bcs of bad signature. Bad signature returned, bcs Rq* fields
	// are part of the signature, but they are not part of json, which was sent
	// we need to keep this test in case Rq* fields will be part of the PoolL2Tx json
	txRepeated1 := atomicGroup.Txs[1]
	jsonTxBytes, err = json.Marshal(txRepeated1)
	require.NoError(t, err)
	jsonTxReader = bytes.NewReader(jsonTxBytes)
	fetchedTxID := common.TxID{}
	require.Error(t, doGoodReq(
		"PUT",
		apiURL+"transactions-pool/"+txRepeated1.TxID.String(),
		jsonTxReader, &fetchedTxID))

	// Test only one tx with fee
	// Generate txs
	txs = []common.PoolL2Tx{}
	baseTx.Nonce = 1
	baseTx.RqNonce = 1 // Nonce incremented just to avoid TxID conflicts
	baseTx.Fee = 0
	baseTx.RqFee = 0
	for i := 0; i < nAccounts; i++ {
		tx := baseTx
		tx.FromIdx = accounts[i].Idx
		tx.ToIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqFromIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqToIdx = accounts[(i+2)%nAccounts].Idx
		if i == 0 {
			tx.Fee = 200
		} else if i == nAccounts-1 {
			tx.RqFee = 200
		}
		if i != nAccounts-1 {
			tx.RqOffset = 1
		} else {
			tx.RqOffset = 5
		}
		txs = append(txs, tx)
	}
	// Sign and format txs
	atomicGroup, txsToReceive = signAndTransformTxs(txs)
	// Send txs
	jsonTxBytes, err = json.Marshal(atomicGroup)
	require.NoError(t, err)
	jsonTxReader = bytes.NewReader(jsonTxBytes)
	fetchedTxIDs = []common.TxID{}
	err = doGoodReq("POST", path, jsonTxReader, &fetchedTxIDs)
	assert.NoError(t, err)
	// Check response
	expectedTxIDs = []common.TxID{}
	for _, tx := range txs {
		expectedTxIDs = append(expectedTxIDs, tx.TxID)
	}
	assert.Equal(t, expectedTxIDs, fetchedTxIDs)
	// Check txs in the DB
	assertTxs(txsToReceive, atomicGroup.ID)

	// Test wrong atomic group id
	txs = []common.PoolL2Tx{}
	baseTx.Nonce = 2
	baseTx.RqNonce = 2 // Nonce incremented just to avoid TxID conflicts
	for i := 0; i < nAccounts; i++ {
		tx := baseTx
		tx.FromIdx = accounts[i].Idx
		tx.ToIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqFromIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqToIdx = accounts[(i+2)%nAccounts].Idx
		if i == 0 {
			tx.Fee = 5
			tx.RqOffset = 1
		} else if i == nAccounts-1 {
			tx.RqFee = 5
			tx.RqOffset = 5
		}
		txs = append(txs, tx)
	}
	// Sign and format txs
	atomicGroup, _ = signAndTransformTxs(txs)
	atomicGroup.ID = common.AtomicGroupID([32]byte{1, 2, 3, 4})
	// Send txs
	jsonTxBytes, err = json.Marshal(atomicGroup)
	require.NoError(t, err)
	jsonTxReader = bytes.NewReader(jsonTxBytes)
	err = doBadReq("POST", path, jsonTxReader, 400)
	assert.NoError(t, err)

	// Test fee too low
	// Generate txs
	txs = []common.PoolL2Tx{}
	baseTx.Nonce = 3
	baseTx.RqNonce = 3 // Nonce incremented just to avoid TxID conflicts
	baseTx.Fee = 0
	baseTx.RqFee = 0
	for i := 0; i < nAccounts; i++ {
		tx := baseTx
		tx.FromIdx = accounts[i].Idx
		tx.ToIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqFromIdx = accounts[(i+1)%nAccounts].Idx
		tx.RqToIdx = accounts[(i+2)%nAccounts].Idx
		if i == 0 {
			tx.Fee = 5
			tx.RqOffset = 1
		} else if i == nAccounts-1 {
			tx.RqFee = 5
			tx.RqOffset = 5
		}
		txs = append(txs, tx)
	}
	// Sign and format txs
	atomicGroup, _ = signAndTransformTxs(txs)
	// Send txs
	jsonTxBytes, err = json.Marshal(atomicGroup)
	require.NoError(t, err)
	jsonTxReader = bytes.NewReader(jsonTxBytes)
	err = doBadReq("POST", path, jsonTxReader, 400)
	assert.NoError(t, err)

	// Test group that is not atomic #1
	/* Note that in this example, txs B and C could be forged without A

	   A  ──────────► B ───────► C
	                  ▲          │
	                  └──────────┘
	*/
	// Generate txs
	txs = []common.PoolL2Tx{}
	// Acyclic part: A  ──────────► B
	A := baseTx
	A.FromIdx = accounts[nAccounts-1].Idx
	A.ToIdx = accounts[0].Idx
	A.RqFromIdx = accounts[0].Idx
	A.RqToIdx = accounts[1].Idx
	A.RqOffset = 1
	txs = append(txs, A)
	/* Cyclic part:
	B ───────► C
	▲          │
	└──────────┘
	*/
	nAccountsMinus1 := nAccounts - 1
	for i := 0; i < nAccountsMinus1; i++ {
		tx := baseTx
		tx.FromIdx = accounts[i].Idx
		tx.ToIdx = accounts[(i+1)%nAccountsMinus1].Idx
		tx.RqFromIdx = accounts[(i+1)%nAccountsMinus1].Idx
		tx.RqToIdx = accounts[(i+2)%nAccountsMinus1].Idx
		if i != nAccountsMinus1-1 {
			tx.RqOffset = 1
		} else {
			tx.RqOffset = 6
		}
		txs = append(txs, tx)
	}
	// Sign and format txs
	atomicGroup, _ = signAndTransformTxs(txs)
	// Send txs
	jsonTxBytes, err = json.Marshal(atomicGroup)
	require.NoError(t, err)
	jsonTxReader = bytes.NewReader(jsonTxBytes)
	err = doBadReq("POST", path, jsonTxReader, 400)
	assert.NoError(t, err)

	// Clean historyDB: the added account shouldn't be there for other tests
	for _, account := range accounts {
		_, err := api.historyDB.DB().DB.Exec(
			fmt.Sprintf("delete from account where idx = %d;", account.Idx),
		)
		assert.NoError(t, err)
	}
	// clean l2DB: the added txs shouldn't be there for other tests
	for _, txID := range txsToClean {
		_, err := api.historyDB.DB().DB.Exec("delete from tx_pool where tx_id = $1;", txID)
		assert.NoError(t, err)
	}
}

func generateKeys(random int) (ethCommon.Address, babyjub.PrivateKey) {
	var key ecdsa.PrivateKey
	key.D = big.NewInt(int64(random)) // only for testing
	key.PublicKey.X, key.PublicKey.Y = ethCrypto.S256().ScalarBaseMult(key.D.Bytes())
	key.Curve = ethCrypto.S256()
	addr := ethCrypto.PubkeyToAddress(key.PublicKey)
	// BJJ private key
	var sk babyjub.PrivateKey
	var iBytes [8]byte
	binary.LittleEndian.PutUint64(iBytes[:], uint64(random))
	copy(sk[:], iBytes[:]) // only for testing
	return addr, sk
}

func TestIsAtomic(t *testing.T) {
	// NOT atomic cases
	// Empty group
	txs := []common.PoolL2Tx{}
	assert.False(t, isSingleAtomicGroup(txs))

	// Case missing tx: 1 ==> 2 ==> 3 ==> (4: not provided)
	txs = []common.PoolL2Tx{
		{TxID: common.TxID{1}, RqOffset: 1},
		{TxID: common.TxID{2}, RqOffset: 1},
		{TxID: common.TxID{3}, RqOffset: 1},
	}
	assert.False(t, isSingleAtomicGroup(txs))

	// Case loneley tx: 1 ==> 2 ==> 3 ==> 1 <== (4: no buddy references 4th tx)
	txs = []common.PoolL2Tx{
		{TxID: common.TxID{1}, RqOffset: 1},
		{TxID: common.TxID{2}, RqOffset: 1},
		{TxID: common.TxID{3}, RqOffset: 6}, // 6 represents -2
		{TxID: common.TxID{4}, RqOffset: 5}, // 5 represents -3
	}
	assert.False(t, isSingleAtomicGroup(txs))

	// Case two groups: 1 <==> 2  3 <==> 4
	txs = []common.PoolL2Tx{
		{TxID: common.TxID{1}, RqOffset: 1},
		{TxID: common.TxID{2}, RqOffset: 7}, // 7 represents -1
		{TxID: common.TxID{3}, RqOffset: 1},
		{TxID: common.TxID{4}, RqOffset: 7}, // 7 represents -1
	}
	assert.False(t, isSingleAtomicGroup(txs))

	// Atomic cases
	// Case circular: 1 ==> 2 ==> 3 ==> 4 ==> 1
	txs = []common.PoolL2Tx{
		{TxID: common.TxID{1}, RqOffset: 1},
		{TxID: common.TxID{2}, RqOffset: 1},
		{TxID: common.TxID{3}, RqOffset: 1},
		{TxID: common.TxID{4}, RqOffset: 5}, // 5 represents -3
	}
	assert.True(t, isSingleAtomicGroup(txs))
}
