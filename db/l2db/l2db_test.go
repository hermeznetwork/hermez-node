package l2db

import (
	"math"
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var l2DB *L2DB
var l2DBWithACC *L2DB
var historyDB *historydb.HistoryDB
var tc *til.Context
var tokens map[common.TokenID]historydb.TokenWithUSD
var tokensValue map[common.TokenID]float64
var accs map[common.Idx]common.Account

func TestMain(m *testing.M) {
	// init DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	l2DB = NewL2DB(db, 10, 1000, 24*time.Hour, nil)
	apiConnCon := dbUtils.NewAPICnnectionController(1, time.Second)
	l2DBWithACC = NewL2DB(db, 10, 1000, 24*time.Hour, apiConnCon)
	test.WipeDB(l2DB.DB())
	historyDB = historydb.NewHistoryDB(db, nil)
	// Run tests
	result := m.Run()
	// Close DB
	if err := db.Close(); err != nil {
		log.Error("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func prepareHistoryDB(historyDB *historydb.HistoryDB) error {
	// Reset DB
	test.WipeDB(l2DB.DB())
	// Generate pool txs using til
	setBlockchain := `
			Type: Blockchain

			AddToken(1)
			AddToken(2)
			CreateAccountDeposit(1) A: 2000
			CreateAccountDeposit(2) A: 2000
			CreateAccountDeposit(1) B: 1000
			CreateAccountDeposit(2) B: 1000
			> batchL1 
			> batchL1
			> block
			> block
			`

	tc = til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(setBlockchain)
	if err != nil {
		return tracerr.Wrap(err)
	}

	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	if err != nil {
		return tracerr.Wrap(err)
	}
	tokens = make(map[common.TokenID]historydb.TokenWithUSD)
	tokensValue = make(map[common.TokenID]float64)
	accs = make(map[common.Idx]common.Account)
	value := 5 * 5.389329
	now := time.Now().UTC()
	// Add all blocks except for the last one
	for i := range blocks[:len(blocks)-1] {
		err = historyDB.AddBlockSCData(&blocks[i])
		if err != nil {
			return tracerr.Wrap(err)
		}
		for _, batch := range blocks[i].Rollup.Batches {
			for _, account := range batch.CreatedAccounts {
				accs[account.Idx] = account
			}
		}
		for _, token := range blocks[i].Rollup.AddedTokens {
			readToken := historydb.TokenWithUSD{
				TokenID:     token.TokenID,
				EthBlockNum: token.EthBlockNum,
				EthAddr:     token.EthAddr,
				Name:        token.Name,
				Symbol:      token.Symbol,
				Decimals:    token.Decimals,
			}
			tokensValue[token.TokenID] = value / math.Pow(10, float64(token.Decimals))
			readToken.USDUpdate = &now
			readToken.USD = &value
			tokens[token.TokenID] = readToken
		}
		// Set value to the tokens (tokens have no symbol)
		tokenSymbol := ""
		err := historyDB.UpdateTokenValue(tokenSymbol, value)
		if err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

func generatePoolL2Txs() ([]common.PoolL2Tx, error) {
	setPool := `
			Type: PoolL2
			PoolTransfer(1) A-B: 6 (4)
			PoolTransfer(2) A-B: 3 (1)
			PoolTransfer(1) B-A: 5 (2)
			PoolTransfer(2) B-A: 10 (3)
			PoolTransfer(1) A-B: 7 (2)
			PoolTransfer(2) A-B: 2 (1)
			PoolTransfer(1) B-A: 8 (2)
			PoolTransfer(2) B-A: 1 (1)
			PoolTransfer(1) A-B: 3 (1)
			PoolTransferToEthAddr(2) B-A: 5 (2)
			PoolTransferToBJJ(2) B-A: 5 (2)

			PoolExit(1) A: 5 (2)
			PoolExit(2) B: 3 (1)
		`
	poolL2Txs, err := tc.GeneratePoolL2Txs(setPool)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return poolL2Txs, nil
}

func TestAddTxTest(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(poolL2Txs[i].TxID)
		assert.NoError(t, err)
		assertTx(t, &poolL2Txs[i], fetchedTx)
		nameZone, offset := fetchedTx.Timestamp.Zone()
		assert.Equal(t, "UTC", nameZone)
		assert.Equal(t, 0, offset)
	}
}
func TestUpdateTxsInfo(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)

		// once added, change the Info parameter
		poolL2Txs[i].Info = "test"
	}
	// update the txs
	err = l2DB.UpdateTxsInfo(poolL2Txs)
	require.NoError(t, err)

	for i := range poolL2Txs {
		fetchedTx, err := l2DB.GetTx(poolL2Txs[i].TxID)
		assert.NoError(t, err)
		assert.Equal(t, "test", fetchedTx.Info)
	}
}

func assertTx(t *testing.T, expected, actual *common.PoolL2Tx) {
	// Check that timestamp has been set within the last 3 seconds
	assert.Less(t, time.Now().UTC().Unix()-3, actual.Timestamp.Unix())
	assert.GreaterOrEqual(t, time.Now().UTC().Unix(), actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp
	// Check absolute fee
	// find token
	token := tokens[expected.TokenID]
	// If the token has value in USD setted
	if token.USDUpdate != nil {
		assert.Less(t, token.USDUpdate.Unix()-3, actual.AbsoluteFeeUpdate.Unix())
		expected.AbsoluteFeeUpdate = actual.AbsoluteFeeUpdate
		// Set expected fee
		f := new(big.Float).SetInt(expected.Amount)
		amountF, _ := f.Float64()
		expected.AbsoluteFee = *token.USD * amountF * expected.Fee.Percentage()
		test.AssertUSD(t, &expected.AbsoluteFee, &actual.AbsoluteFee)
	}
	assert.Equal(t, expected, actual)
}

// NO UPDATE: benchmarks will be done after impl is finished
// func BenchmarkAddTxTest(b *testing.B) {
// 	const nInserts = 20
// 	test.WipeDB(l2DB.DB())
// 	txs := test.GenPoolTxs(nInserts, tokens)
// 	now := time.Now()
// 	for _, tx := range txs {
// 		_ = l2DB.AddTxTest(tx)
// 	}
// 	elapsedTime := time.Since(now)
// 	log.Info("Time to insert 2048 txs:", elapsedTime)
// }

func TestGetPending(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	var pendingTxs []*common.PoolL2Tx
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		pendingTxs = append(pendingTxs, &poolL2Txs[i])
	}
	fetchedTxs, err := l2DB.GetPendingTxs()
	assert.NoError(t, err)
	assert.Equal(t, len(pendingTxs), len(fetchedTxs))
	for i := range fetchedTxs {
		assertTx(t, pendingTxs[i], &fetchedTxs[i])
	}
}

func TestStartForging(t *testing.T) {
	// Generate txs
	var fakeBatchNum common.BatchNum = 33
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	var startForgingTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			startForgingTxIDs = append(startForgingTxIDs, poolL2Txs[i].TxID)
		}
		randomizer++
	}
	// Start forging txs
	err = l2DB.StartForging(startForgingTxIDs, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range startForgingTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateForging, fetchedTx.State)
		assert.Equal(t, &fakeBatchNum, fetchedTx.BatchNum)
	}
}

func TestDoneForging(t *testing.T) {
	// Generate txs
	var fakeBatchNum common.BatchNum = 33
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	var startForgingTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			startForgingTxIDs = append(startForgingTxIDs, poolL2Txs[i].TxID)
		}
		randomizer++
	}
	// Start forging txs
	err = l2DB.StartForging(startForgingTxIDs, fakeBatchNum)
	assert.NoError(t, err)

	var doneForgingTxIDs []common.TxID
	randomizer = 0
	for _, txID := range startForgingTxIDs {
		if randomizer%2 == 0 {
			doneForgingTxIDs = append(doneForgingTxIDs, txID)
		}
		randomizer++
	}
	// Done forging txs
	err = l2DB.DoneForging(doneForgingTxIDs, fakeBatchNum)
	assert.NoError(t, err)

	// Fetch txs and check that they've been updated correctly
	for _, id := range doneForgingTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateForged, fetchedTx.State)
		assert.Equal(t, &fakeBatchNum, fetchedTx.BatchNum)
	}
}

func TestInvalidate(t *testing.T) {
	// Generate txs
	var fakeBatchNum common.BatchNum = 33
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	var invalidTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		if poolL2Txs[i].State != common.PoolL2TxStateInvalid && randomizer%2 == 0 {
			randomizer++
			invalidTxIDs = append(invalidTxIDs, poolL2Txs[i].TxID)
		}
	}
	// Invalidate txs
	err = l2DB.InvalidateTxs(invalidTxIDs, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range invalidTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateInvalid, fetchedTx.State)
		assert.Equal(t, &fakeBatchNum, fetchedTx.BatchNum)
	}
}

func TestInvalidateOldNonces(t *testing.T) {
	// Generate txs
	var fakeBatchNum common.BatchNum = 33
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)
	// Update Accounts currentNonce
	var updateAccounts []common.IdxNonce
	var currentNonce = common.Nonce(1)
	for i := range accs {
		updateAccounts = append(updateAccounts, common.IdxNonce{
			Idx:   accs[i].Idx,
			Nonce: common.Nonce(currentNonce),
		})
	}
	// Add txs to DB
	var invalidTxIDs []common.TxID
	for i := range poolL2Txs {
		if poolL2Txs[i].Nonce < currentNonce {
			invalidTxIDs = append(invalidTxIDs, poolL2Txs[i].TxID)
		}
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
	}
	// sanity check
	require.Greater(t, len(invalidTxIDs), 0)

	err = l2DB.InvalidateOldNonces(updateAccounts, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range invalidTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateInvalid, fetchedTx.State)
		assert.Equal(t, &fakeBatchNum, fetchedTx.BatchNum)
	}
}

// TestReorg: first part of the test with reorg
// With invalidated transactions BEFORE reorgBatch
// And forged transactions in reorgBatch
func TestReorg(t *testing.T) {
	// Generate txs
	const lastValidBatch common.BatchNum = 20
	const reorgBatch common.BatchNum = lastValidBatch + 1
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)

	reorgedTxIDs := []common.TxID{}
	nonReorgedTxIDs := []common.TxID{}
	var startForgingTxIDs []common.TxID
	var invalidTxIDs []common.TxID
	var allTxRandomize []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			startForgingTxIDs = append(startForgingTxIDs, poolL2Txs[i].TxID)
			allTxRandomize = append(allTxRandomize, poolL2Txs[i].TxID)
		} else if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%3 == 0 {
			invalidTxIDs = append(invalidTxIDs, poolL2Txs[i].TxID)
			allTxRandomize = append(allTxRandomize, poolL2Txs[i].TxID)
		}
		randomizer++
	}
	// Start forging txs
	err = l2DB.StartForging(startForgingTxIDs, lastValidBatch)
	assert.NoError(t, err)

	var doneForgingTxIDs []common.TxID
	randomizer = 0
	for _, txID := range allTxRandomize {
		invalidTx := false
		for i := range invalidTxIDs {
			if invalidTxIDs[i] == txID {
				invalidTx = true
				nonReorgedTxIDs = append(nonReorgedTxIDs, txID)
			}
		}
		if !invalidTx {
			if randomizer%2 == 0 {
				doneForgingTxIDs = append(doneForgingTxIDs, txID)
				reorgedTxIDs = append(reorgedTxIDs, txID)
			} else {
				nonReorgedTxIDs = append(nonReorgedTxIDs, txID)
			}
			randomizer++
		}
	}

	// Invalidate txs BEFORE reorgBatch --> nonReorg
	err = l2DB.InvalidateTxs(invalidTxIDs, lastValidBatch)
	assert.NoError(t, err)
	// Done forging txs in reorgBatch --> Reorg
	err = l2DB.DoneForging(doneForgingTxIDs, reorgBatch)
	assert.NoError(t, err)

	err = l2DB.Reorg(lastValidBatch)
	assert.NoError(t, err)
	for _, id := range reorgedTxIDs {
		tx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Nil(t, tx.BatchNum)
		assert.Equal(t, common.PoolL2TxStatePending, tx.State)
	}
	for _, id := range nonReorgedTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Equal(t, lastValidBatch, *fetchedTx.BatchNum)
	}
}

// TestReorg: second part of test with reorg
// With invalidated transactions in reorgBatch
// And forged transactions BEFORE reorgBatch
func TestReorg2(t *testing.T) {
	// Generate txs
	const lastValidBatch common.BatchNum = 20
	const reorgBatch common.BatchNum = lastValidBatch + 1
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)

	reorgedTxIDs := []common.TxID{}
	nonReorgedTxIDs := []common.TxID{}
	var startForgingTxIDs []common.TxID
	var invalidTxIDs []common.TxID
	var allTxRandomize []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		assert.NoError(t, err)
		if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			startForgingTxIDs = append(startForgingTxIDs, poolL2Txs[i].TxID)
			allTxRandomize = append(allTxRandomize, poolL2Txs[i].TxID)
		} else if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%3 == 0 {
			invalidTxIDs = append(invalidTxIDs, poolL2Txs[i].TxID)
			allTxRandomize = append(allTxRandomize, poolL2Txs[i].TxID)
		}
		randomizer++
	}
	// Start forging txs
	err = l2DB.StartForging(startForgingTxIDs, lastValidBatch)
	assert.NoError(t, err)

	var doneForgingTxIDs []common.TxID
	randomizer = 0
	for _, txID := range allTxRandomize {
		invalidTx := false
		for i := range invalidTxIDs {
			if invalidTxIDs[i] == txID {
				invalidTx = true
				reorgedTxIDs = append(reorgedTxIDs, txID)
			}
		}
		if !invalidTx {
			if randomizer%2 == 0 {
				doneForgingTxIDs = append(doneForgingTxIDs, txID)
			}
			nonReorgedTxIDs = append(nonReorgedTxIDs, txID)
			randomizer++
		}
	}
	// Done forging txs BEFORE reorgBatch --> nonReorg
	err = l2DB.DoneForging(doneForgingTxIDs, lastValidBatch)
	assert.NoError(t, err)
	// Invalidate txs in reorgBatch --> Reorg
	err = l2DB.InvalidateTxs(invalidTxIDs, reorgBatch)
	assert.NoError(t, err)

	err = l2DB.Reorg(lastValidBatch)
	assert.NoError(t, err)
	for _, id := range reorgedTxIDs {
		tx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Nil(t, tx.BatchNum)
		assert.Equal(t, common.PoolL2TxStatePending, tx.State)
	}
	for _, id := range nonReorgedTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		assert.NoError(t, err)
		assert.Equal(t, lastValidBatch, *fetchedTx.BatchNum)
	}
}

func TestPurge(t *testing.T) {
	// Generate txs
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	// generatePoolL2Txs
	generateTx := int(l2DB.maxTxs/8 + 1)
	var poolL2Tx []common.PoolL2Tx
	for i := 0; i < generateTx; i++ {
		poolL2TxAux, err := generatePoolL2Txs()
		assert.NoError(t, err)
		poolL2Tx = append(poolL2Tx, poolL2TxAux...)
	}

	afterTTLIDs := []common.TxID{}
	keepedIDs := []common.TxID{}
	var deletedIDs []common.TxID
	var invalidTxIDs []common.TxID
	var doneForgingTxIDs []common.TxID
	const toDeleteBatchNum common.BatchNum = 30
	safeBatchNum := toDeleteBatchNum + l2DB.safetyPeriod + 1
	// Add txs to the DB
	for i := 0; i < len(poolL2Tx); i++ {
		tx := poolL2Tx[i]
		if i%2 == 0 { // keep tx
			keepedIDs = append(keepedIDs, tx.TxID)
		} else { // delete after safety period
			if i%3 == 0 {
				doneForgingTxIDs = append(doneForgingTxIDs, tx.TxID)
			} else if i%5 == 0 {
				invalidTxIDs = append(invalidTxIDs, tx.TxID)
			} else {
				afterTTLIDs = append(afterTTLIDs, tx.TxID)
			}
			deletedIDs = append(deletedIDs, poolL2Tx[i].TxID)
		}
		err := l2DB.AddTxTest(&tx)
		assert.NoError(t, err)
	}
	// Set batchNum keeped txs
	for i := range keepedIDs {
		_, err = l2DB.db.Exec(
			"UPDATE tx_pool SET batch_num = $1 WHERE tx_id = $2;",
			safeBatchNum, keepedIDs[i],
		)
		assert.NoError(t, err)
	}
	// Start forging txs and set batchNum
	err = l2DB.StartForging(doneForgingTxIDs, toDeleteBatchNum)
	assert.NoError(t, err)
	// Done forging txs and set batchNum
	err = l2DB.DoneForging(doneForgingTxIDs, toDeleteBatchNum)
	assert.NoError(t, err)
	// Invalidate txs and set batchNum
	err = l2DB.InvalidateTxs(invalidTxIDs, toDeleteBatchNum)
	assert.NoError(t, err)
	// Update timestamp of afterTTL txs
	deleteTimestamp := time.Unix(time.Now().UTC().Unix()-int64(l2DB.ttl.Seconds()+float64(4*time.Second)), 0)
	for _, id := range afterTTLIDs {
		// Set timestamp
		_, err = l2DB.db.Exec(
			"UPDATE tx_pool SET timestamp = $1, state = $2 WHERE tx_id = $3;",
			deleteTimestamp, common.PoolL2TxStatePending, id,
		)
		assert.NoError(t, err)
	}

	// Purge txs
	err = l2DB.Purge(safeBatchNum)
	assert.NoError(t, err)
	// Check results
	for _, id := range deletedIDs {
		_, err := l2DB.GetTx(id)
		assert.Error(t, err)
	}
	for _, id := range keepedIDs {
		_, err := l2DB.GetTx(id)
		assert.NoError(t, err)
	}
}

func TestAuth(t *testing.T) {
	test.WipeDB(l2DB.DB())
	const nAuths = 5
	chainID := uint16(0)
	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	// Generate authorizations
	auths := test.GenAuths(nAuths, chainID, hermezContractAddr)
	for i := 0; i < len(auths); i++ {
		// Add to the DB
		err := l2DB.AddAccountCreationAuth(auths[i])
		assert.NoError(t, err)
		// Fetch from DB
		auth, err := l2DB.GetAccountCreationAuth(auths[i].EthAddr)
		assert.NoError(t, err)
		// Check fetched vs generated
		assert.Equal(t, auths[i].EthAddr, auth.EthAddr)
		assert.Equal(t, auths[i].BJJ, auth.BJJ)
		assert.Equal(t, auths[i].Signature, auth.Signature)
		assert.Equal(t, auths[i].Timestamp.Unix(), auths[i].Timestamp.Unix())
		nameZone, offset := auths[i].Timestamp.Zone()
		assert.Equal(t, "UTC", nameZone)
		assert.Equal(t, 0, offset)
	}
}

func TestAddGet(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	assert.NoError(t, err)

	// We will work with only 3 txs
	require.GreaterOrEqual(t, len(poolL2Txs), 3)
	txs := poolL2Txs[:3]
	// NOTE: By changing the tx fields, the signature will no longer be
	// valid, but we are not checking the signautre here so it's OK.
	// 0. Has ToIdx >= 256 && ToEthAddr == 0 && ToBJJ == 0
	require.GreaterOrEqual(t, int(txs[0].ToIdx), 256)
	txs[0].ToEthAddr = ethCommon.Address{}
	txs[0].ToBJJ = babyjub.PublicKeyComp{}
	// 1. Has ToIdx >= 256 && ToEthAddr != 0 && ToBJJ != 0
	require.GreaterOrEqual(t, int(txs[1].ToIdx), 256)
	require.NotEqual(t, txs[1].ToEthAddr, ethCommon.Address{})
	require.NotEqual(t, txs[1].ToBJJ, babyjub.PublicKeyComp{})
	// 2. Has ToIdx == 0 && ToEthAddr != 0 && ToBJJ != 0
	txs[2].ToIdx = 0
	require.NotEqual(t, txs[2].ToEthAddr, ethCommon.Address{})
	require.NotEqual(t, txs[2].ToBJJ, babyjub.PublicKeyComp{})

	for i := 0; i < len(txs); i++ {
		require.NoError(t, txs[i].SetID())
		require.NoError(t, l2DB.AddTxTest(&txs[i]))
	}
	// Verify that the inserts haven't altered any field (specially
	// ToEthAddr and ToBJJ)
	for i := 0; i < len(txs); i++ {
		dbTx, err := l2DB.GetTx(txs[i].TxID)
		require.NoError(t, err)
		// Ignore Timestamp, AbsoluteFee, AbsoluteFeeUpdate
		txs[i].Timestamp = dbTx.Timestamp
		txs[i].AbsoluteFee = dbTx.AbsoluteFee
		txs[i].AbsoluteFeeUpdate = dbTx.AbsoluteFeeUpdate
		assert.Equal(t, txs[i], *dbTx)
	}
}
