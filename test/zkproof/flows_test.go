package zkproof

import (
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteme []string

func addTokens(t *testing.T, tc *til.Context, db *sqlx.DB) {
	var tokens []common.Token
	for i := 0; i < int(tc.LastRegisteredTokenID); i++ {
		tokens = append(tokens, common.Token{
			TokenID:     common.TokenID(i + 1),
			EthBlockNum: 1,
			EthAddr:     ethCommon.BytesToAddress([]byte{byte(i + 1)}),
			Name:        strconv.Itoa(i),
			Symbol:      strconv.Itoa(i),
			Decimals:    18,
		})
	}

	hdb := historydb.NewHistoryDB(db, db, nil)
	assert.NoError(t, hdb.AddBlock(&common.Block{
		Num: 1,
	}))
	assert.NoError(t, hdb.AddTokens(tokens))
}

func addL2Txs(t *testing.T, l2DB *l2db.L2DB, poolL2Txs []common.PoolL2Tx) {
	for i := 0; i < len(poolL2Txs); i++ {
		err := l2DB.AddTxTest(&poolL2Txs[i])
		if err != nil {
			log.Error(err)
		}
		require.NoError(t, err)
	}
}

func addAccCreationAuth(t *testing.T, tc *til.Context, l2DB *l2db.L2DB, chainID uint16,
	hermezContractAddr ethCommon.Address, username string) []byte {
	user := tc.Users[username]
	auth := &common.AccountCreationAuth{
		EthAddr: user.Addr,
		BJJ:     user.BJJ.Public().Compress(),
	}
	err := auth.Sign(func(hash []byte) ([]byte, error) {
		return ethCrypto.Sign(hash, user.EthSk)
	}, chainID, hermezContractAddr)
	assert.NoError(t, err)

	err = l2DB.AddAccountCreationAuth(auth)
	assert.NoError(t, err)
	return auth.Signature
}

func initTxSelector(t *testing.T, chainID uint16, hermezContractAddr ethCommon.Address,
	coordUser *til.User) (*txselector.TxSelector, *l2db.L2DB, *statedb.StateDB) {
	db, err := test.InitTestSQLDB()
	require.NoError(t, err)
	l2DB := l2db.NewL2DB(db, db, 10, 100, 0.0, 1000.0, 24*time.Hour, nil)

	dir, err := ioutil.TempDir("", "tmpSyncDB")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)
	syncStateDB, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 0})
	require.NoError(t, err)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.NoError(t, err)
	deleteme = append(deleteme, txselDir)

	// use Til Coord keys for tests compatibility
	coordAccount := &txselector.CoordAccount{
		Addr:                coordUser.Addr,
		BJJ:                 coordUser.BJJ.Public().Compress(),
		AccountCreationAuth: nil,
	}
	auth := common.AccountCreationAuth{
		EthAddr: coordUser.Addr,
		BJJ:     coordUser.BJJ.Public().Compress(),
	}
	err = auth.Sign(func(hash []byte) ([]byte, error) {
		return ethCrypto.Sign(hash, coordUser.EthSk)
	}, chainID, hermezContractAddr)
	assert.NoError(t, err)
	coordAccount.AccountCreationAuth = auth.Signature

	txsel, err := txselector.NewTxSelector(coordAccount, txselDir, syncStateDB, l2DB)
	require.NoError(t, err)

	test.WipeDB(l2DB.DB())

	return txsel, l2DB, syncStateDB
}

func TestTxSelectorBatchBuilderZKInputsMinimumFlow0(t *testing.T) {
	tc := til.NewContext(ChainID, common.RollupConstMaxL1UserTx)
	// generate test transactions, the L1CoordinatorTxs generated by Til
	// will be ignored at this test, as will be the TxSelector who
	// generates them when needed
	blocks, err := tc.GenerateBlocks(txsets.SetBlockchainMinimumFlow0)
	require.NoError(t, err)

	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	txsel, l2DBTxSel, syncStateDB := initTxSelector(t, ChainID, hermezContractAddr, tc.Users["Coord"])

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.NoError(t, err)
	deleteme = append(deleteme, bbDir)
	bb, err := batchbuilder.NewBatchBuilder(bbDir, syncStateDB, 0, NLevels)
	require.NoError(t, err)

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()

	// add tokens to HistoryDB to avoid breaking FK constrains
	addTokens(t, tc, l2DBTxSel.DB())

	configBatch := &batchbuilder.ConfigBatch{
		// ForgerAddress:
		TxProcessorConfig: txprocConfig,
	}

	// loop over the first 6 batches
	expectedRoots := []string{"0", "0",
		"10303926118213025243660668481827257778714122989909761705455084995854999537039",
		"8530501758307821623834726627056947648600328521261384179220598288701741436285",
		"8530501758307821623834726627056947648600328521261384179220598288701741436285",
		"9061858435528794221929846392270405504056106238451760714188625065949729889651"}
	for i := 0; i < 6; i++ {
		log.Debugf("block:0 batch:%d", i+1)
		var l1UserTxs []common.L1Tx
		if blocks[0].Rollup.Batches[i].Batch.ForgeL1TxsNum != nil {
			l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[i].Batch.ForgeL1TxsNum])
		}
		// TxSelector select the transactions for the next Batch
		coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err :=
			txsel.GetL1L2TxSelection(txprocConfig, l1UserTxs, nil)
		require.NoError(t, err)
		// BatchBuilder build Batch
		zki, err := bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
		require.NoError(t, err)
		assert.Equal(t, expectedRoots[i], bb.LocalStateDB().MT.Root().BigInt().String())
		sendProofAndCheckResp(t, zki)
	}

	log.Debug("block:0 batch:7")
	// simulate the PoolL2Txs of the batch6
	batchPoolL2 := `
	Type: PoolL2
	PoolTransferToEthAddr(1) A-B: 200 (126)
	PoolTransferToEthAddr(0) B-C: 100 (126)`
	l2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	// add AccountCreationAuths that will be used at the next batch
	_ = addAccCreationAuth(t, tc, l2DBTxSel, ChainID, hermezContractAddr, "B")
	_ = addAccCreationAuth(t, tc, l2DBTxSel, ChainID, hermezContractAddr, "C")
	addL2Txs(t, l2DBTxSel, l2Txs) // Add L2s to TxSelector.L2DB
	l1UserTxs := til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[6].Batch.ForgeL1TxsNum])
	// TxSelector select the transactions for the next Batch
	coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err :=
		txsel.GetL1L2TxSelection(txprocConfig, l1UserTxs, nil)
	require.NoError(t, err)
	// BatchBuilder build Batch
	zki, err := bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"4392049343656836675348565048374261353937130287163762821533580216441778455298",
		bb.LocalStateDB().MT.Root().BigInt().String())
	sendProofAndCheckResp(t, zki)
	err = l2DBTxSel.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs),
		txsel.LocalAccountsDB().CurrentBatch())
	require.NoError(t, err)
	err = l2DBTxSel.UpdateTxsInfo(discardedL2Txs)
	require.NoError(t, err)

	log.Debug("block:0 batch:8")
	// simulate the PoolL2Txs of the batch8
	batchPoolL2 = `
	Type: PoolL2
	PoolTransfer(0) A-B: 100 (126)
	PoolTransfer(0) C-A: 50 (126)
	PoolTransfer(1) B-C: 100 (126)
	PoolExit(0) A: 100 (126)`
	l2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, l2DBTxSel, l2Txs) // Add L2s to TxSelector.L2DB
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[7].Batch.ForgeL1TxsNum])
	// TxSelector select the transactions for the next Batch
	coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err =
		txsel.GetL1L2TxSelection(txprocConfig, l1UserTxs, nil)
	require.NoError(t, err)
	// BatchBuilder build Batch
	zki, err = bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"8905191229562583213069132470917469035834300549892959854483573322676101624713",
		bb.LocalStateDB().MT.Root().BigInt().String())
	sendProofAndCheckResp(t, zki)
	err = l2DBTxSel.StartForging(common.TxIDsFromPoolL2Txs(l2Txs),
		txsel.LocalAccountsDB().CurrentBatch())
	require.NoError(t, err)
	err = l2DBTxSel.UpdateTxsInfo(discardedL2Txs)
	require.NoError(t, err)

	log.Debug("(batch9) block:1 batch:1")
	// simulate the PoolL2Txs of the batch9
	batchPoolL2 = `
	Type: PoolL2
	PoolTransfer(0) D-A: 300 (126)
	PoolTransfer(0) B-D: 100 (126)`
	l2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, l2DBTxSel, l2Txs) // Add L2s to TxSelector.L2DB
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[0].Batch.ForgeL1TxsNum])
	// TxSelector select the transactions for the next Batch
	coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err =
		txsel.GetL1L2TxSelection(txprocConfig, l1UserTxs, nil)
	require.NoError(t, err)
	// BatchBuilder build Batch
	zki, err = bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"20593679664586247774284790801579542411781976279024409415159440382607791042723",
		bb.LocalStateDB().MT.Root().BigInt().String())
	sendProofAndCheckResp(t, zki)
	err = l2DBTxSel.StartForging(common.TxIDsFromPoolL2Txs(l2Txs),
		txsel.LocalAccountsDB().CurrentBatch())
	require.NoError(t, err)
	err = l2DBTxSel.UpdateTxsInfo(discardedL2Txs)
	require.NoError(t, err)

	log.Debug("(batch10) block:1 batch:2")
	l2Txs = []common.PoolL2Tx{}
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	// TxSelector select the transactions for the next Batch
	coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err =
		txsel.GetL1L2TxSelection(txprocConfig, l1UserTxs, nil)
	require.NoError(t, err)
	// BatchBuilder build Batch
	zki, err = bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
	require.NoError(t, err)
	// same root as previous batch, as the L1CoordinatorTxs created by the
	// Til set is not created by the TxSelector in this test
	assert.Equal(t,
		"20593679664586247774284790801579542411781976279024409415159440382607791042723",
		bb.LocalStateDB().MT.Root().BigInt().String())
	sendProofAndCheckResp(t, zki)
	err = l2DBTxSel.StartForging(common.TxIDsFromPoolL2Txs(l2Txs),
		txsel.LocalAccountsDB().CurrentBatch())
	require.NoError(t, err)
	err = l2DBTxSel.UpdateTxsInfo(discardedL2Txs)
	require.NoError(t, err)

	bb.LocalStateDB().Close()
	txsel.LocalAccountsDB().Close()
	syncStateDB.Close()
}

// TestZKInputsExitWithFee0 checks the case where there is a PoolTxs of type
// Exit with fee 0 for a TokenID that the Coordinator does not have it
// registered yet
func TestZKInputsExitWithFee0(t *testing.T) {
	tc := til.NewContext(ChainID, common.RollupConstMaxL1UserTx)

	var set = `
	Type: Blockchain
	AddToken(1)

	CreateAccountDeposit(1) A: 1000
	CreateAccountDeposit(1) B: 1000
	CreateAccountDeposit(1) C: 1000
	> batchL1
	> batchL1

	CreateAccountCoordinator(1) Coord
	> batch
	> block
	`
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)

	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	txsel, l2DBTxSel, syncStateDB := initTxSelector(t, ChainID, hermezContractAddr, tc.Users["Coord"])

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.NoError(t, err)
	deleteme = append(deleteme, bbDir)
	bb, err := batchbuilder.NewBatchBuilder(bbDir, syncStateDB, 0, NLevels)
	require.NoError(t, err)

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()
	// add tokens to HistoryDB to avoid breaking FK constrains
	addTokens(t, tc, l2DBTxSel.DB())

	configBatch := &batchbuilder.ConfigBatch{
		TxProcessorConfig: txprocConfig,
	}

	// batch2
	// TxSelector select the transactions for the next Batch
	l1UserTxs := til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err :=
		txsel.GetL1L2TxSelection(txprocConfig, l1UserTxs, nil)
	require.NoError(t, err)
	// BatchBuilder build Batch
	zki, err := bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"3050252508378236752695438107925920517579600844238792454632938959089837319058",
		bb.LocalStateDB().MT.Root().BigInt().String())
	h, err := zki.HashGlobalData()
	require.NoError(t, err)
	assert.Equal(t,
		"136173330006576039857485697813777018179965431269591881328654192642028135989",
		h.String())
	sendProofAndCheckResp(t, zki)

	// batch3
	batchPoolL2 := `
	Type: PoolL2
	PoolExit(1) A: 100 (0)`
	l2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, l2DBTxSel, l2Txs) // Add L2s to TxSelector.L2DB
	coordIdxs, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err :=
		txsel.GetL1L2TxSelection(txprocConfig, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, len(coordIdxs))
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 1, len(oL1CoordTxs))
	assert.Equal(t, 1, len(oL2Txs))
	assert.Equal(t, 0, len(discardedL2Txs))
	// BatchBuilder build Batch
	zki, err = bb.BuildBatch(coordIdxs, configBatch, oL1UserTxs, oL1CoordTxs, oL2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"2941150582529643425331223235752941075548157545257982041291886277157404095484",
		bb.LocalStateDB().MT.Root().BigInt().String())
	h, err = zki.HashGlobalData()
	require.NoError(t, err)
	assert.Equal(t,
		"11526955144859107275861838429358092025337347677758832533226842081116224550335",
		h.String())
	assert.Equal(t, common.EthAddrToBigInt(tc.Users["Coord"].Addr), zki.EthAddr3[0])
	assert.Equal(t, "0", zki.EthAddr3[1].String())
	sendProofAndCheckResp(t, zki)

	bb.LocalStateDB().Close()
	txsel.LocalAccountsDB().Close()
	syncStateDB.Close()
}
