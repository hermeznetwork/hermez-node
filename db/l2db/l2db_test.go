package l2db

import (
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
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

var decimals = uint64(3)
var tokenValue = 1.0 // The price update gives a value of 1.0 USD to the token
var l2DB *L2DB
var l2DBWithACC *L2DB
var historyDB *historydb.HistoryDB
var tc *til.Context
var tokens map[common.TokenID]historydb.TokenWithUSD

var accs map[common.Idx]common.Account

func TestMain(m *testing.M) {
	// init DB
	db, err := dbUtils.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	l2DB = NewL2DB(db, db, 10, 1000, 0.0, 1000.0, 24*time.Hour, nil)
	apiConnCon := dbUtils.NewAPIConnectionController(1, time.Second)
	l2DBWithACC = NewL2DB(db, db, 10, 1000, 0.0, 1000.0, 24*time.Hour, apiConnCon)
	test.WipeDB(l2DB.DB())
	historyDB = historydb.NewHistoryDB(db, db, nil)
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
			CreateAccountDeposit(1) A: 20000
			CreateAccountDeposit(2) A: 20000
			CreateAccountDeposit(1) B: 10000
			CreateAccountDeposit(2) B: 10000
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
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Rollup.AddedTokens {
			token := &block.Rollup.AddedTokens[j]
			token.Name = fmt.Sprintf("Token %d", token.TokenID)
			token.Symbol = fmt.Sprintf("TK%d", token.TokenID)
			token.Decimals = decimals
		}
	}

	tokens = make(map[common.TokenID]historydb.TokenWithUSD)
	// tokensValue = make(map[common.TokenID]float64)
	accs = make(map[common.Idx]common.Account)
	now := time.Now().UTC()
	// Add all blocks except for the last one
	for i := range blocks[:len(blocks)-1] {
		if err := historyDB.AddBlockSCData(&blocks[i]); err != nil {
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
				USD:         &tokenValue,
				USDUpdate:   &now,
			}
			tokens[token.TokenID] = readToken
			// Set value to the tokens
			err := historyDB.UpdateTokenValue(readToken.EthAddr, *readToken.USD)
			if err != nil {
				return tracerr.Wrap(err)
			}
		}
	}
	return nil
}

func generatePoolL2Txs() ([]common.PoolL2Tx, error) {
	// Fee = 126 corresponds to ~10%
	setPool := `
			Type: PoolL2
			PoolTransfer(1) A-B: 6000 (126)
			PoolTransfer(2) A-B: 3000 (126)
			PoolTransfer(1) B-A: 5000 (126)
			PoolTransfer(2) B-A: 10000 (126)
			PoolTransfer(1) A-B: 7000 (126)
			PoolTransfer(2) A-B: 2000 (126)
			PoolTransfer(1) B-A: 8000 (126)
			PoolTransfer(2) B-A: 1000 (126)
			PoolTransfer(1) A-B: 3000 (126)
			PoolTransferToEthAddr(2) B-A: 5000 (126)
			PoolTransferToBJJ(2) B-A: 5000 (126)

			PoolExit(1) A: 5000 (126)
			PoolExit(2) B: 3000 (126)
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
	require.NoError(t, err)
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(poolL2Txs[i].TxID)
		require.NoError(t, err)
		assertTx(t, &poolL2Txs[i], fetchedTx)
		nameZone, offset := fetchedTx.Timestamp.Zone()
		assert.Equal(t, "UTC", nameZone)
		assert.Equal(t, 0, offset)
	}

	// test, that we can update already existing tx
	tx := &poolL2Txs[1]
	fetchedTx, err := l2DB.GetTx(tx.TxID)
	require.NoError(t, err)
	assert.Equal(t, fetchedTx.ToIdx, tx.ToIdx)
	tx.ToIdx = common.Idx(1)
	err = l2DBWithACC.UpdateTxAPI(tx)
	require.NoError(t, err)
	fetchedTx, err = l2DB.GetTx(tx.TxID)
	require.NoError(t, err)
	assert.Equal(t, fetchedTx.ToIdx, common.Idx(1))
}

func TestAddTxAPI(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}

	oldMaxTxs := l2DBWithACC.maxTxs
	// set max number of pending txs that can be kept in the pool to 5
	l2DBWithACC.maxTxs = 5

	poolL2Txs, err := generatePoolL2Txs()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(poolL2Txs), 8)
	for i := range poolL2Txs[:5] {
		err := l2DBWithACC.AddTxAPI(&poolL2Txs[i])
		require.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(poolL2Txs[i].TxID)
		require.NoError(t, err)
		assertTx(t, &poolL2Txs[i], fetchedTx)
		nameZone, offset := fetchedTx.Timestamp.Zone()
		assert.Equal(t, "UTC", nameZone)
		assert.Equal(t, 0, offset)
	}
	err = l2DBWithACC.AddTxAPI(&poolL2Txs[5])
	assert.Equal(t, errPoolFull, tracerr.Unwrap(err))
	// reset maxTxs to original value
	l2DBWithACC.maxTxs = oldMaxTxs

	// set minFeeUSD to a high value than the tx feeUSD to test the error
	// of inserting a tx with lower than min fee
	oldMinFeeUSD := l2DBWithACC.minFeeUSD
	tx := poolL2Txs[5]
	feeAmount, err := common.CalcFeeAmount(tx.Amount, tx.Fee)
	require.NoError(t, err)
	feeAmountUSD := common.TokensToUSD(feeAmount, decimals, tokenValue)
	// set minFeeUSD higher than the tx fee to trigger the error
	l2DBWithACC.minFeeUSD = feeAmountUSD + 1
	err = l2DBWithACC.AddTxAPI(&tx)
	require.Error(t, err)
	assert.Regexp(t, "tx.feeUSD (.*) < minFeeUSD (.*)", err.Error())
	// reset minFeeUSD to original value
	l2DBWithACC.minFeeUSD = oldMinFeeUSD
}

func TestUpdateTxsInfo(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	require.NoError(t, err)
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)

		// once added, change the Info parameter
		poolL2Txs[i].Info = "test"
	}
	// update the txs
	var batchNum common.BatchNum
	err = l2DB.UpdateTxsInfo(poolL2Txs, batchNum)
	require.NoError(t, err)

	for i := range poolL2Txs {
		fetchedTx, err := l2DB.GetTx(poolL2Txs[i].TxID)
		require.NoError(t, err)
		assert.Equal(t, "BatchNum: 0. test", fetchedTx.Info)
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
		amountUSD := common.TokensToUSD(expected.Amount, token.Decimals, *token.USD)
		expected.AbsoluteFee = amountUSD * expected.Fee.Percentage()
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
	require.NoError(t, err)
	var pendingTxs []*common.PoolL2Tx
	// Add case for fields that have been added after the original schema
	poolL2Txs[0].AtomicGroupID = common.AtomicGroupID([common.AtomicGroupIDLen]byte{9})
	poolL2Txs[0].RqNonce = 1
	poolL2Txs[0].RqTokenID = 1
	poolL2Txs[0].RqFromIdx = 678
	poolL2Txs[0].RqToIdx = 679
	poolL2Txs[0].RqAmount = big.NewInt(99999)
	poolL2Txs[0].RqFee = 200
	poolL2Txs[0].RqOffset = 3
	poolL2Txs[0].RqToEthAddr = ethCommon.BigToAddress(big.NewInt(11111111))
	poolL2Txs[0].MaxNumBatch = 123456

	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
		pendingTxs = append(pendingTxs, &poolL2Txs[i])
	}
	fetchedTxs, err := l2DB.GetPendingTxs()
	require.NoError(t, err)
	assert.Equal(t, len(pendingTxs), len(fetchedTxs))
	for i := range fetchedTxs {
		assertTx(t, pendingTxs[i], &fetchedTxs[i])
	}
	// Check AbsoluteFee amount
	for i := range fetchedTxs {
		tx := &fetchedTxs[i]
		feeAmount, err := common.CalcFeeAmount(tx.Amount, tx.Fee)
		require.NoError(t, err)
		feeAmountUSD := common.TokensToUSD(feeAmount,
			tokens[tx.TokenID].Decimals, *tokens[tx.TokenID].USD)
		assert.InEpsilon(t, feeAmountUSD, tx.AbsoluteFee, 0.01)
	}
}

func TestL2DB_GetPoolTxs(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	require.NoError(t, err)
	state := common.PoolL2TxState("pend")
	idx := common.Idx(256)
	fromItem := uint(0)
	limit := uint(10)
	var pendingTxs []*common.PoolL2Tx
	for i := range poolL2Txs {
		if poolL2Txs[i].FromIdx == idx || poolL2Txs[i].ToIdx == idx {
			err := l2DB.AddTxTest(&poolL2Txs[i])
			require.NoError(t, err)
			pendingTxs = append(pendingTxs, &poolL2Txs[i])
		}
	}
	fetchedTxs, _, err := l2DBWithACC.GetPoolTxsAPI(GetPoolTxsAPIRequest{
		Idx:      &idx,
		State:    &state,
		FromItem: &fromItem,
		Limit:    &limit,
		Order:    dbUtils.OrderAsc,
	})
	require.NoError(t, err)
	assert.Equal(t, len(pendingTxs), len(fetchedTxs))
}

func TestStartForging(t *testing.T) {
	// Generate txs
	var fakeBatchNum common.BatchNum = 33
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	poolL2Txs, err := generatePoolL2Txs()
	require.NoError(t, err)
	var startForgingTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
		if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			startForgingTxIDs = append(startForgingTxIDs, poolL2Txs[i].TxID)
		}
		randomizer++
	}
	// Start forging txs
	err = l2DB.StartForging(startForgingTxIDs, fakeBatchNum)
	require.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range startForgingTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
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
	require.NoError(t, err)
	var startForgingTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
		if poolL2Txs[i].State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			startForgingTxIDs = append(startForgingTxIDs, poolL2Txs[i].TxID)
		}
		randomizer++
	}
	// Start forging txs
	err = l2DB.StartForging(startForgingTxIDs, fakeBatchNum)
	require.NoError(t, err)

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
	require.NoError(t, err)

	// Fetch txs and check that they've been updated correctly
	for _, id := range doneForgingTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
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
	require.NoError(t, err)
	var invalidTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
		if poolL2Txs[i].State != common.PoolL2TxStateInvalid && randomizer%2 == 0 {
			randomizer++
			invalidTxIDs = append(invalidTxIDs, poolL2Txs[i].TxID)
		}
	}
	// Invalidate txs
	err = l2DB.InvalidateTxs(invalidTxIDs, fakeBatchNum)
	require.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range invalidTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
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
	require.NoError(t, err)
	// Update Accounts currentNonce
	var updateAccounts []common.IdxNonce
	var currentNonce = nonce.Nonce(1)
	for i := range accs {
		updateAccounts = append(updateAccounts, common.IdxNonce{
			Idx:   accs[i].Idx,
			Nonce: nonce.Nonce(currentNonce),
		})
	}
	// Add txs to DB
	var invalidTxIDs []common.TxID
	for i := range poolL2Txs {
		if poolL2Txs[i].Nonce < currentNonce {
			invalidTxIDs = append(invalidTxIDs, poolL2Txs[i].TxID)
		}
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
	}
	// sanity check
	require.Greater(t, len(invalidTxIDs), 0)

	err = l2DB.InvalidateOldNonces(updateAccounts, fakeBatchNum)
	require.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range invalidTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateInvalid, fetchedTx.State)
		assert.Equal(t, &fakeBatchNum, fetchedTx.BatchNum)
		assert.Equal(t, invalidateOldNoncesInfo, *fetchedTx.Info)
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
	require.NoError(t, err)

	reorgedTxIDs := []common.TxID{}
	nonReorgedTxIDs := []common.TxID{}
	var startForgingTxIDs []common.TxID
	var invalidTxIDs []common.TxID
	var allTxRandomize []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
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
	require.NoError(t, err)

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
	require.NoError(t, err)
	// Done forging txs in reorgBatch --> Reorg
	err = l2DB.DoneForging(doneForgingTxIDs, reorgBatch)
	require.NoError(t, err)

	err = l2DB.Reorg(lastValidBatch)
	require.NoError(t, err)
	for _, id := range reorgedTxIDs {
		tx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
		assert.Nil(t, tx.BatchNum)
		assert.Equal(t, common.PoolL2TxStatePending, tx.State)
	}
	for _, id := range nonReorgedTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
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
	require.NoError(t, err)

	reorgedTxIDs := []common.TxID{}
	nonReorgedTxIDs := []common.TxID{}
	var startForgingTxIDs []common.TxID
	var invalidTxIDs []common.TxID
	var allTxRandomize []common.TxID
	randomizer := 0
	// Add txs to DB
	for i := range poolL2Txs {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
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
	require.NoError(t, err)

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
	require.NoError(t, err)
	// Invalidate txs in reorgBatch --> Reorg
	err = l2DB.InvalidateTxs(invalidTxIDs, reorgBatch)
	require.NoError(t, err)

	err = l2DB.Reorg(lastValidBatch)
	require.NoError(t, err)
	for _, id := range reorgedTxIDs {
		tx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
		assert.Nil(t, tx.BatchNum)
		assert.Equal(t, common.PoolL2TxStatePending, tx.State)
	}
	for _, id := range nonReorgedTxIDs {
		fetchedTx, err := l2DBWithACC.GetTxAPI(id)
		require.NoError(t, err)
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
		require.NoError(t, err)
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
		require.NoError(t, err)
	}
	// Set batchNum keeped txs
	for i := range keepedIDs {
		_, err = l2DB.dbWrite.Exec(
			"UPDATE tx_pool SET batch_num = $1 WHERE tx_id = $2;",
			safeBatchNum, keepedIDs[i],
		)
		require.NoError(t, err)
	}
	// Start forging txs and set batchNum
	err = l2DB.StartForging(doneForgingTxIDs, toDeleteBatchNum)
	require.NoError(t, err)
	// Done forging txs and set batchNum
	err = l2DB.DoneForging(doneForgingTxIDs, toDeleteBatchNum)
	require.NoError(t, err)
	// Invalidate txs and set batchNum
	err = l2DB.InvalidateTxs(invalidTxIDs, toDeleteBatchNum)
	require.NoError(t, err)
	// Update timestamp of afterTTL txs
	deleteTimestamp := time.Unix(time.Now().UTC().Unix()-int64(l2DB.ttl.Seconds()+float64(4*time.Second)), 0)
	for _, id := range afterTTLIDs {
		// Set timestamp
		_, err = l2DB.dbWrite.Exec(
			"UPDATE tx_pool SET timestamp = $1, state = $2 WHERE tx_id = $3;",
			deleteTimestamp, common.PoolL2TxStatePending, id,
		)
		require.NoError(t, err)
	}

	// Purge txs
	err = l2DB.Purge(safeBatchNum)
	require.NoError(t, err)
	// Check results
	for _, id := range deletedIDs {
		_, err := l2DB.GetTx(id)
		assert.Error(t, err)
	}
	for _, id := range keepedIDs {
		_, err := l2DB.GetTx(id)
		require.NoError(t, err)
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
		require.NoError(t, err)
		// Fetch from DB
		auth, err := l2DB.GetAccountCreationAuth(auths[i].EthAddr)
		require.NoError(t, err)
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

func TestManyAuth(t *testing.T) {
	test.WipeDB(l2DB.DB())
	const nAuths = 5
	chainID := uint16(0)
	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	// Generate authorizations
	genAuths := test.GenAuths(nAuths, chainID, hermezContractAddr)
	auths := make([]common.AccountCreationAuth, len(genAuths))
	// Convert to a non-pointer slice
	for i := 0; i < len(genAuths); i++ {
		auths[i] = *genAuths[i]
	}

	// Add a duplicate one to check the not exist condition
	err := l2DB.AddAccountCreationAuth(genAuths[0])
	require.NoError(t, err)

	// Add to the DB
	err = l2DB.AddManyAccountCreationAuth(auths)
	require.NoError(t, err)

	// Assert the result
	for i := 0; i < len(auths); i++ {
		// Fetch from DB
		auth, err := l2DB.GetAccountCreationAuth(auths[i].EthAddr)
		require.NoError(t, err)
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
	require.NoError(t, err)

	// We will work with only 3 txs
	require.GreaterOrEqual(t, len(poolL2Txs), 3)
	txs := poolL2Txs[:3]
	// NOTE: By changing the tx fields, the signature will no longer be
	// valid, but we are not checking the signature here so it's OK.
	// 0. Has ToIdx >= 256 && ToEthAddr == 0 && ToBJJ == 0
	require.GreaterOrEqual(t, int(txs[0].ToIdx), 256)
	txs[0].ToEthAddr = ethCommon.Address{}
	txs[0].ToBJJ = babyjub.PublicKeyComp{}
	// 1. Has ToIdx >= 256 && ToEthAddr != 0 && ToBJJ != 0
	require.GreaterOrEqual(t, int(txs[1].ToIdx), 256)
	txs[1].ToEthAddr = common.FFAddr
	sk := babyjub.NewRandPrivKey()
	txs[1].ToBJJ = sk.Public().Compress()
	// 2. Has ToIdx == 0 && ToEthAddr != 0 && ToBJJ != 0
	txs[2].ToIdx = 0
	txs[2].ToEthAddr = common.FFAddr
	sk = babyjub.NewRandPrivKey()
	txs[2].ToBJJ = sk.Public().Compress()

	for i := 0; i < len(txs); i++ {
		require.NoError(t, txs[i].SetID())
		require.NoError(t, l2DB.AddTxTest(&txs[i]))
	}
	// Verify that the inserts haven't altered any field (specially
	// ToEthAddr and ToBJJ)
	for i := 0; i < len(txs); i++ {
		dbTx, err := l2DB.GetTx(txs[i].TxID)
		require.NoError(t, err)
		assertTx(t, &txs[i], dbTx)
	}
}

func TestPurgeByExternalDelete(t *testing.T) {
	err := prepareHistoryDB(historyDB)
	if err != nil {
		log.Error("Error prepare historyDB", err)
	}
	txs, err := generatePoolL2Txs()
	require.NoError(t, err)

	// We will work with 8 txs
	require.GreaterOrEqual(t, len(txs), 8)
	txs = txs[:8]
	for i := range txs {
		require.NoError(t, l2DB.AddTxTest(&txs[i]))
	}

	// We will recreate this scenario:
	// tx index, status , external_delete
	// 0       , pending, false
	// 1       , pending, false
	// 2       , pending, true // will be deleted
	// 3       , pending, true // will be deleted
	// 4       , fging  , false
	// 5       , fging  , false
	// 6       , fging  , true
	// 7       , fging  , true

	require.NoError(t, l2DB.StartForging(
		[]common.TxID{txs[4].TxID, txs[5].TxID, txs[6].TxID, txs[7].TxID},
		1))
	_, err = l2DB.dbWrite.Exec(
		`UPDATE tx_pool SET external_delete = true WHERE
			tx_id IN ($1, $2, $3, $4)
		;`,
		txs[2].TxID, txs[3].TxID, txs[6].TxID, txs[7].TxID,
	)
	require.NoError(t, err)
	require.NoError(t, l2DB.PurgeByExternalDelete())

	// Query txs that are have been not deleted
	for _, i := range []int{0, 1, 4, 5, 6, 7} {
		txID := txs[i].TxID
		_, err := l2DB.GetTx(txID)
		require.NoError(t, err)
	}

	// Query txs that have been deleted
	for _, i := range []int{2, 3} {
		txID := txs[i].TxID
		_, err := l2DB.GetTx(txID)
		require.Equal(t, sql.ErrNoRows, tracerr.Unwrap(err))
	}
}
