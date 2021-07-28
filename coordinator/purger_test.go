package coordinator

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newL2DB(t *testing.T) *l2db.L2DB {
	db, err := dbUtils.InitTestSQLDB()
	require.NoError(t, err)
	test.WipeDB(db)
	return l2db.NewL2DB(db, db, 10, 100, 0.0, 1000.0, 24*time.Hour, nil)
}

func newStateDB(t *testing.T) (*statedb.LocalStateDB, *statedb.StateDB) {
	syncDBPath, err := ioutil.TempDir("", "tmpSyncDB")
	require.NoError(t, err)
	deleteme = append(deleteme, syncDBPath)
	syncStateDB, err := statedb.NewStateDB(statedb.Config{Path: syncDBPath, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 48})
	assert.NoError(t, err)
	stateDBPath, err := ioutil.TempDir("", "tmpStateDB")
	require.NoError(t, err)
	deleteme = append(deleteme, stateDBPath)
	stateDB, err := statedb.NewLocalStateDB(statedb.Config{Path: stateDBPath, Keep: 128,
		Type: statedb.TypeTxSelector, NLevels: 0}, syncStateDB)
	require.NoError(t, err)
	return stateDB, syncStateDB
}

func TestCanPurgeCanInvalidate(t *testing.T) {
	cfg := PurgerCfg{
		PurgeBatchDelay:      2,
		PurgeBlockDelay:      6,
		InvalidateBatchDelay: 4,
		InvalidateBlockDelay: 8,
	}
	p := Purger{
		cfg: cfg,
	}
	startBlockNum := int64(1000)
	startBatchNum := int64(10)
	blockNum := startBlockNum
	batchNum := startBatchNum

	assert.True(t, p.CanPurge(blockNum, batchNum))
	p.lastPurgeBlock = startBlockNum
	p.lastPurgeBatch = startBatchNum
	assert.False(t, p.CanPurge(blockNum, batchNum))

	blockNum = startBlockNum + cfg.PurgeBlockDelay - 1
	batchNum = startBatchNum + cfg.PurgeBatchDelay - 1
	assert.False(t, p.CanPurge(blockNum, batchNum))
	blockNum = startBlockNum + cfg.PurgeBlockDelay - 1
	batchNum = startBatchNum + cfg.PurgeBatchDelay
	assert.True(t, p.CanPurge(blockNum, batchNum))
	blockNum = startBlockNum + cfg.PurgeBlockDelay
	batchNum = startBatchNum + cfg.PurgeBatchDelay - 1
	assert.True(t, p.CanPurge(blockNum, batchNum))

	assert.True(t, p.CanInvalidate(blockNum, batchNum))
	p.lastInvalidateBlock = startBlockNum
	p.lastInvalidateBatch = startBatchNum
	assert.False(t, p.CanInvalidate(blockNum, batchNum))

	blockNum = startBlockNum + cfg.InvalidateBlockDelay - 1
	batchNum = startBatchNum + cfg.InvalidateBatchDelay - 1
	assert.False(t, p.CanInvalidate(blockNum, batchNum))
	blockNum = startBlockNum + cfg.InvalidateBlockDelay - 1
	batchNum = startBatchNum + cfg.InvalidateBatchDelay
	assert.True(t, p.CanInvalidate(blockNum, batchNum))
	blockNum = startBlockNum + cfg.InvalidateBlockDelay
	batchNum = startBatchNum + cfg.InvalidateBatchDelay - 1
	assert.True(t, p.CanInvalidate(blockNum, batchNum))
}

func TestPurgeMaybeInvalidateMaybe(t *testing.T) {
	cfg := PurgerCfg{
		PurgeBatchDelay:      2,
		PurgeBlockDelay:      6,
		InvalidateBatchDelay: 4,
		InvalidateBlockDelay: 8,
	}
	p := Purger{
		cfg: cfg,
	}
	l2DB := newL2DB(t)
	stateDB, syncStateDB := newStateDB(t)

	startBlockNum := int64(1000)
	startBatchNum := int64(10)

	p.lastPurgeBlock = startBlockNum
	p.lastPurgeBatch = startBatchNum

	blockNum := startBlockNum + cfg.PurgeBlockDelay - 1
	batchNum := startBatchNum + cfg.PurgeBatchDelay - 1
	ok, err := p.PurgeMaybe(l2DB, blockNum, batchNum)
	require.NoError(t, err)
	assert.False(t, ok)
	// At this point the purger will purge.  The second time it doesn't
	// because it the first time it has updates the last time it did.
	blockNum = startBlockNum + cfg.PurgeBlockDelay - 1
	batchNum = startBatchNum + cfg.PurgeBatchDelay
	ok, err = p.PurgeMaybe(l2DB, blockNum, batchNum)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = p.PurgeMaybe(l2DB, blockNum, batchNum)
	require.NoError(t, err)
	assert.False(t, ok)

	p.lastInvalidateBlock = startBlockNum
	p.lastInvalidateBatch = startBatchNum

	blockNum = startBlockNum + cfg.InvalidateBlockDelay - 1
	batchNum = startBatchNum + cfg.InvalidateBatchDelay - 1
	ok, err = p.InvalidateMaybe(l2DB, stateDB, blockNum, batchNum)
	require.NoError(t, err)
	assert.False(t, ok)
	// At this point the purger will invaidate.  The second time it doesn't
	// because it the first time it has updates the last time it did.
	blockNum = startBlockNum + cfg.InvalidateBlockDelay - 1
	batchNum = startBatchNum + cfg.InvalidateBatchDelay
	ok, err = p.InvalidateMaybe(l2DB, stateDB, blockNum, batchNum)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = p.InvalidateMaybe(l2DB, stateDB, blockNum, batchNum)
	require.NoError(t, err)
	assert.False(t, ok)

	syncStateDB.Close()
	stateDB.StateDB.Close()
	_ = l2DB.DB().Close()
}

func TestIdxsNonce(t *testing.T) {
	inputIdxsNonce := []common.IdxNonce{
		{Idx: 256, Nonce: 1},
		{Idx: 256, Nonce: 2},
		{Idx: 257, Nonce: 3},
		{Idx: 258, Nonce: 5},
		{Idx: 258, Nonce: 2},
	}
	expectedIdxsNonce := map[common.Idx]nonce.Nonce{
		common.Idx(256): nonce.Nonce(2),
		common.Idx(257): nonce.Nonce(3),
		common.Idx(258): nonce.Nonce(5),
	}

	l2txs := make([]common.L2Tx, len(inputIdxsNonce))
	for i, idxNonce := range inputIdxsNonce {
		l2txs[i].FromIdx = idxNonce.Idx
		l2txs[i].Nonce = idxNonce.Nonce
	}
	idxsNonce := idxsNonceFromL2Txs(l2txs)
	assert.Equal(t, len(expectedIdxsNonce), len(idxsNonce))
	for _, idxNonce := range idxsNonce {
		nonce := expectedIdxsNonce[idxNonce.Idx]
		assert.Equal(t, nonce, idxNonce.Nonce)
	}

	pooll2txs := make([]common.PoolL2Tx, len(inputIdxsNonce))
	for i, idxNonce := range inputIdxsNonce {
		pooll2txs[i].FromIdx = idxNonce.Idx
		pooll2txs[i].Nonce = idxNonce.Nonce
	}
	idxsNonce = idxsNonceFromPoolL2Txs(pooll2txs)
	assert.Equal(t, len(expectedIdxsNonce), len(idxsNonce))
	for _, idxNonce := range idxsNonce {
		nonce := expectedIdxsNonce[idxNonce.Idx]
		assert.Equal(t, nonce, idxNonce.Nonce)
	}
}

func TestPoolMarkInvalidOldNonces(t *testing.T) {
	l2DB := newL2DB(t)
	stateDB, syncStateDB := newStateDB(t)

	set0 := `
		Type: Blockchain

		CreateAccountDeposit(0) A: 1000 // Idx=256
		CreateAccountDeposit(0) B: 1000 // Idx=257
		CreateAccountDeposit(0) C: 1000 // Idx=258
		CreateAccountDeposit(0) D: 1000 // Idx=259

		> batchL1
		> batchL1
		> block
	`
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set0)
	require.NoError(t, err)
	tilCfgExtra := til.ConfigExtra{
		CoordUser: "A",
	}
	// Call FillBlocksExtra to fill `Batch.CreatedAccounts`
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	require.Equal(t, 4, len(blocks[0].Rollup.Batches[1].CreatedAccounts)) // sanity check

	for _, acc := range blocks[0].Rollup.Batches[1].CreatedAccounts {
		_, err := stateDB.CreateAccount(acc.Idx, &acc) //nolint:gosec
		require.NoError(t, err)
	}

	setPool0 := `
		Type: PoolL2
		PoolTransfer(0) A-B: 10 (1)
		PoolTransfer(0) A-C: 10 (1)
		PoolTransfer(0) A-D: 10 (1)
		PoolTransfer(0) B-A: 10 (1)
		PoolTransfer(0) B-C: 10 (1)
		PoolTransfer(0) C-A: 10 (1)
	`
	// We expect the following nonces
	nonces0 := map[string]int64{"A": 3, "B": 2, "C": 1, "D": 0}
	l2txs0, err := tc.GeneratePoolL2Txs(setPool0)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(l2txs0))
	for _, tx := range l2txs0 {
		require.NoError(t, l2DB.AddTxTest(&tx)) //nolint:gosec
	}

	// Update the accounts in the StateDB, making the txs in the setPool0
	// invalid
	for name, user := range tc.Users {
		for _, _acc := range user.Accounts {
			require.Equal(t, nonce.Nonce(nonces0[name]), _acc.Nonce) // sanity check
			acc, err := stateDB.GetAccount(_acc.Idx)
			require.NoError(t, err)
			require.Equal(t, nonce.Nonce(0), acc.Nonce) // sanity check
			acc.Nonce = _acc.Nonce
			_, err = stateDB.UpdateAccount(acc.Idx, acc)
			require.NoError(t, err)
		}
	}

	setPool1 := `
		Type: PoolL2
		PoolTransfer(0) A-B: 10 (1)
		PoolTransfer(0) A-C: 10 (1)
		PoolTransfer(0) A-D: 10 (1)
		PoolTransfer(0) B-A: 10 (1)
		PoolTransfer(0) B-C: 10 (1)
		PoolTransfer(0) C-A: 10 (1)
	`
	// We expect the following nonces
	nonces1 := map[string]int64{"A": 6, "B": 4, "C": 2, "D": 0}
	l2txs1, err := tc.GeneratePoolL2Txs(setPool1)
	require.NoError(t, err)
	assert.Equal(t, 6, len(l2txs1))
	for _, tx := range l2txs1 {
		require.NoError(t, l2DB.AddTxTest(&tx)) //nolint:gosec
	}

	for name, user := range tc.Users {
		for _, _acc := range user.Accounts {
			require.Equal(t, nonce.Nonce(nonces1[name]), _acc.Nonce) // sanity check
			acc, err := stateDB.GetAccount(_acc.Idx)
			require.NoError(t, err)
			require.Equal(t, nonce.Nonce(nonces0[name]), acc.Nonce) // sanity check
		}
	}

	// Now we should have 12 txs in the pool, all marked as pending.  Since
	// we updated the stateDB with the nonces after setPool0, the first 6
	// txs will be marked as invalid

	pendingTxs, err := l2DB.GetPendingTxs()
	require.NoError(t, err)
	assert.Equal(t, 12, len(pendingTxs))

	batchNum := common.BatchNum(1)
	err = poolMarkInvalidOldNonces(l2DB, stateDB, batchNum)
	require.NoError(t, err)

	pendingTxs, err = l2DB.GetPendingTxs()
	require.NoError(t, err)
	assert.Equal(t, 6, len(pendingTxs))

	syncStateDB.Close()
	stateDB.StateDB.Close()
	_ = l2DB.DB().Close()
}
