package txselector

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T, chainID uint16, hermezContractAddr ethCommon.Address, coordUser *til.User) *TxSelector {
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.NoError(t, err)
	l2DB := l2db.NewL2DB(db, 10, 100, 0.0, 24*time.Hour, nil)

	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))
	syncStateDB, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeTxSelector, NLevels: 0})
	require.NoError(t, err)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	// use Til Coord keys for tests compatibility
	coordAccount := &CoordAccount{
		Addr:                coordUser.Addr,
		BJJ:                 coordUser.BJJ.Public().Compress(),
		AccountCreationAuth: nil,
	}
	// fmt.Printf("%v\n", coordAccount)
	auth := common.AccountCreationAuth{
		EthAddr: coordUser.Addr,
		BJJ:     coordUser.BJJ.Public().Compress(),
	}
	err = auth.Sign(func(hash []byte) ([]byte, error) {
		return ethCrypto.Sign(hash, coordUser.EthSk)
	}, chainID, hermezContractAddr)
	assert.NoError(t, err)
	coordAccount.AccountCreationAuth = auth.Signature

	txsel, err := NewTxSelector(coordAccount, txselDir, syncStateDB, l2DB)
	require.NoError(t, err)

	test.WipeDB(txsel.l2db.DB())

	return txsel
}

func addAccCreationAuth(t *testing.T, tc *til.Context, txsel *TxSelector, chainID uint16, hermezContractAddr ethCommon.Address, username string) []byte {
	user := tc.Users[username]
	auth := &common.AccountCreationAuth{
		EthAddr: user.Addr,
		BJJ:     user.BJJ.Public().Compress(),
	}
	err := auth.Sign(func(hash []byte) ([]byte, error) {
		return ethCrypto.Sign(hash, user.EthSk)
	}, chainID, hermezContractAddr)
	assert.NoError(t, err)

	err = txsel.l2db.AddAccountCreationAuth(auth)
	assert.NoError(t, err)
	return auth.Signature
}

func addL2Txs(t *testing.T, txsel *TxSelector, poolL2Txs []common.PoolL2Tx) {
	for i := 0; i < len(poolL2Txs); i++ {
		err := txsel.l2db.AddTxTest(&poolL2Txs[i])
		if err != nil {
			log.Error(err)
		}
		require.NoError(t, err)
	}
}

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

	hdb := historydb.NewHistoryDB(db, nil)
	assert.NoError(t, hdb.AddBlock(&common.Block{
		Num: 1,
	}))
	assert.NoError(t, hdb.AddTokens(tokens))
}

func checkBalance(t *testing.T, tc *til.Context, txsel *TxSelector, username string, tokenid int, expected string) {
	// Accounts.Idx does not match with the TxSelector tests as we are not
	// using the Til L1CoordinatorTxs (as are generated by the TxSelector
	// itself when processing the txs, so the Idxs does not match the Til
	// idxs). But the Idx is obtained through StateDB.GetIdxByEthAddrBJJ
	user := tc.Users[username]
	idx, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(user.Addr, user.BJJ.Public().Compress(), common.TokenID(tokenid))
	require.NoError(t, err)
	checkBalanceByIdx(t, txsel, idx, expected)
}

func checkBalanceByIdx(t *testing.T, txsel *TxSelector, idx common.Idx, expected string) {
	acc, err := txsel.localAccountsDB.GetAccount(idx)
	require.NoError(t, err)
	assert.Equal(t, expected, acc.Balance.String())
}

// checkSortedByNonce takes as input testAccNonces map, and the array of
// common.PoolL2Txs, and checks if the nonces correspond to the accumulated
// values of the map. Also increases the Nonces computed on the map.
func checkSortedByNonce(t *testing.T, testAccNonces map[common.Idx]common.Nonce, txs []common.PoolL2Tx) {
	for _, tx := range txs {
		assert.True(t, testAccNonces[tx.FromIdx] == tx.Nonce,
			fmt.Sprintf("Idx: %d, expected: %d, tx.Nonce: %d",
				tx.FromIdx, testAccNonces[tx.FromIdx], tx.Nonce))
		testAccNonces[tx.FromIdx] = testAccNonces[tx.FromIdx] + 1
	}
}

func TestGetL2TxSelectionMinimumFlow0(t *testing.T) {
	chainID := uint16(0)
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	// generate test transactions, the L1CoordinatorTxs generated by Til
	// will be ignored at this test, as will be the TxSelector who
	// generates them when needed
	blocks, err := tc.GenerateBlocks(txsets.SetBlockchainMinimumFlow0)
	assert.NoError(t, err)

	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	txsel := initTest(t, chainID, hermezContractAddr, tc.Users["Coord"])

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()
	testAccNonces := make(map[common.Idx]common.Nonce)

	// add tokens to HistoryDB to avoid breaking FK constrains
	addTokens(t, tc, txsel.l2db.DB())

	tpc := txprocessor.Config{
		NLevels:  16,
		MaxFeeTx: 10,
		MaxTx:    20,
		MaxL1Tx:  10,
		ChainID:  chainID,
	}
	selectionConfig := &SelectionConfig{
		MaxL1UserTxs:      5,
		TxProcessorConfig: tpc,
	}

	// coordIdxs, accAuths, l1UserTxs, l1CoordTxs, l2Txs, err

	log.Debug("block:0 batch:1")
	l1UserTxs := []common.L1Tx{}
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err := txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs))
	assert.Equal(t, common.BatchNum(1), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(255), txsel.localAccountsDB.CurrentIdx())

	log.Debug("block:0 batch:2")
	l1UserTxs = []common.L1Tx{}
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs))
	assert.Equal(t, common.BatchNum(2), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(255), txsel.localAccountsDB.CurrentIdx())

	log.Debug("block:0 batch:3")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[2].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 2, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs))
	assert.Equal(t, common.BatchNum(3), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(257), txsel.localAccountsDB.CurrentIdx())
	checkBalance(t, tc, txsel, "A", 0, "500")
	checkBalance(t, tc, txsel, "C", 1, "0")

	log.Debug("block:0 batch:4")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[3].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 1, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs))
	assert.Equal(t, common.BatchNum(4), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(258), txsel.localAccountsDB.CurrentIdx())
	checkBalance(t, tc, txsel, "A", 0, "500")
	checkBalance(t, tc, txsel, "A", 1, "500")
	checkBalance(t, tc, txsel, "C", 1, "0")

	log.Debug("block:0 batch:5")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[4].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs))
	assert.Equal(t, common.BatchNum(5), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(258), txsel.localAccountsDB.CurrentIdx())
	checkBalance(t, tc, txsel, "A", 0, "500")
	checkBalance(t, tc, txsel, "A", 1, "500")
	checkBalance(t, tc, txsel, "C", 1, "0")

	log.Debug("block:0 batch:6")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[5].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 1, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs))
	assert.Equal(t, common.BatchNum(6), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(259), txsel.localAccountsDB.CurrentIdx())
	checkBalance(t, tc, txsel, "A", 0, "600")
	checkBalance(t, tc, txsel, "A", 1, "500")
	checkBalance(t, tc, txsel, "B", 0, "400")
	checkBalance(t, tc, txsel, "C", 1, "0")

	log.Debug("block:0 batch:7")
	// simulate the PoolL2Txs of the batch7
	batchPoolL2 := `
	Type: PoolL2
	PoolTransferToEthAddr(1) A-B: 200 (126)
	PoolTransferToEthAddr(0) B-C: 100 (126)`
	poolL2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	// add AccountCreationAuths that will be used at the next batch
	accAuthSig0 := addAccCreationAuth(t, tc, txsel, chainID, hermezContractAddr, "B")
	accAuthSig1 := addAccCreationAuth(t, tc, txsel, chainID, hermezContractAddr, "C")
	// add the PoolL2Txs to the l2DB
	addL2Txs(t, txsel, poolL2Txs)
	// check signatures of L2Txs from the L2DB (to check that the
	// parameters of the PoolL2Tx match the original parameters signed
	// before inserting it to the L2DB)
	l2TxsFromDB, err := txsel.l2db.GetPendingTxs()
	require.NoError(t, err)
	assert.True(t, l2TxsFromDB[0].VerifySignature(chainID, tc.Users["A"].BJJ.Public().Compress()))
	assert.True(t, l2TxsFromDB[1].VerifySignature(chainID, tc.Users["B"].BJJ.Public().Compress()))
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[6].Batch.ForgeL1TxsNum])
	coordIdxs, accAuths, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err := txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, []common.Idx{261, 262}, coordIdxs)
	assert.Equal(t, txsel.coordAccount.AccountCreationAuth, accAuths[0])
	assert.Equal(t, txsel.coordAccount.AccountCreationAuth, accAuths[1])
	assert.Equal(t, accAuthSig0, accAuths[2])
	assert.Equal(t, accAuthSig1, accAuths[3])
	assert.Equal(t, 1, len(oL1UserTxs))
	assert.Equal(t, 4, len(oL1CoordTxs))
	assert.Equal(t, 2, len(oL2Txs))
	assert.Equal(t, len(oL1CoordTxs), len(accAuths))
	assert.Equal(t, common.BatchNum(7), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(264), txsel.localAccountsDB.CurrentIdx())
	checkBalanceByIdx(t, txsel, 261, "20") // CoordIdx for TokenID=1
	checkBalanceByIdx(t, txsel, 262, "10") // CoordIdx for TokenID=0
	checkBalance(t, tc, txsel, "A", 0, "600")
	checkBalance(t, tc, txsel, "A", 1, "280")
	checkBalance(t, tc, txsel, "B", 0, "290")
	checkBalance(t, tc, txsel, "B", 1, "200")
	checkBalance(t, tc, txsel, "C", 0, "100")
	checkBalance(t, tc, txsel, "D", 0, "800")
	checkSortedByNonce(t, testAccNonces, oL2Txs)
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(poolL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)

	log.Debug("block:0 batch:8")
	// simulate the PoolL2Txs of the batch8
	batchPoolL2 = `
	Type: PoolL2
	PoolTransfer(0) A-B: 100 (126)
	PoolTransfer(0) C-A: 50 (126)
	PoolTransfer(1) B-C: 100 (126)
	PoolExit(0) A: 100 (126)`
	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, txsel, poolL2Txs)
	// check signatures of L2Txs from the L2DB (to check that the
	// parameters of the PoolL2Tx match the original parameters signed
	// before inserting it to the L2DB)
	l2TxsFromDB, err = txsel.l2db.GetPendingTxs()
	require.NoError(t, err)
	assert.True(t, l2TxsFromDB[0].VerifySignature(chainID, tc.Users["A"].BJJ.Public().Compress()))
	assert.True(t, l2TxsFromDB[1].VerifySignature(chainID, tc.Users["C"].BJJ.Public().Compress()))
	assert.True(t, l2TxsFromDB[2].VerifySignature(chainID, tc.Users["B"].BJJ.Public().Compress()))
	assert.True(t, l2TxsFromDB[3].VerifySignature(chainID, tc.Users["A"].BJJ.Public().Compress()))
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[7].Batch.ForgeL1TxsNum])
	coordIdxs, accAuths, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, []common.Idx{261, 262}, coordIdxs)
	assert.Equal(t, 0, len(accAuths))
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 4, len(oL2Txs))
	assert.Equal(t, len(oL1CoordTxs), len(accAuths))
	assert.Equal(t, common.BatchNum(8), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(264), txsel.localAccountsDB.CurrentIdx())
	checkBalanceByIdx(t, txsel, 261, "30")
	checkBalanceByIdx(t, txsel, 262, "35")
	checkBalance(t, tc, txsel, "A", 0, "430")
	checkBalance(t, tc, txsel, "A", 1, "280")
	checkBalance(t, tc, txsel, "B", 0, "390")
	checkBalance(t, tc, txsel, "B", 1, "90")
	checkBalance(t, tc, txsel, "C", 0, "45")
	checkBalance(t, tc, txsel, "C", 1, "100")
	checkBalance(t, tc, txsel, "D", 0, "800")
	checkSortedByNonce(t, testAccNonces, oL2Txs)
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(poolL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)

	log.Debug("(batch9)block:1 batch:1")
	// simulate the PoolL2Txs of the batch9
	batchPoolL2 = `
	Type: PoolL2
	PoolTransfer(0) D-A: 300 (126)
	PoolTransfer(0) B-D: 100 (126)
	`
	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, txsel, poolL2Txs)
	// check signatures of L2Txs from the L2DB (to check that the
	// parameters of the PoolL2Tx match the original parameters signed
	// before inserting it to the L2DB)
	l2TxsFromDB, err = txsel.l2db.GetPendingTxs()
	require.NoError(t, err)
	assert.True(t, l2TxsFromDB[0].VerifySignature(chainID, tc.Users["D"].BJJ.Public().Compress()))
	assert.True(t, l2TxsFromDB[1].VerifySignature(chainID, tc.Users["B"].BJJ.Public().Compress()))
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[0].Batch.ForgeL1TxsNum])
	coordIdxs, accAuths, oL1UserTxs, oL1CoordTxs, oL2Txs, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, []common.Idx{262}, coordIdxs)
	assert.Equal(t, 0, len(accAuths))
	assert.Equal(t, 4, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 2, len(oL2Txs))
	assert.Equal(t, len(oL1CoordTxs), len(accAuths))
	assert.Equal(t, common.BatchNum(9), txsel.localAccountsDB.CurrentBatch())
	assert.Equal(t, common.Idx(264), txsel.localAccountsDB.CurrentIdx())
	checkBalanceByIdx(t, txsel, 261, "30")
	checkBalanceByIdx(t, txsel, 262, "75")
	checkBalance(t, tc, txsel, "A", 0, "730")
	checkBalance(t, tc, txsel, "A", 1, "280")
	checkBalance(t, tc, txsel, "B", 0, "380")
	checkBalance(t, tc, txsel, "B", 1, "90")
	checkBalance(t, tc, txsel, "C", 0, "845")
	checkBalance(t, tc, txsel, "C", 1, "100")
	checkBalance(t, tc, txsel, "D", 0, "470")
	checkSortedByNonce(t, testAccNonces, oL2Txs)
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(poolL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)
}

func TestPoolL2TxsWithoutEnoughBalance(t *testing.T) {
	set := `
		Type: Blockchain

		CreateAccountDeposit(0) Coord: 0
		CreateAccountDeposit(0) A: 100
		CreateAccountDeposit(0) B: 100

		> batchL1 // freeze L1User{1}
		> batchL1 // forge L1User{1}
		> block
	`

	chainID := uint16(0)
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	// generate test transactions, the L1CoordinatorTxs generated by Til
	// will be ignored at this test, as will be the TxSelector who
	// generates them when needed
	blocks, err := tc.GenerateBlocks(set)
	assert.NoError(t, err)

	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	txsel := initTest(t, chainID, hermezContractAddr, tc.Users["Coord"])

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()

	tpc := txprocessor.Config{
		NLevels:  16,
		MaxFeeTx: 10,
		MaxTx:    20,
		MaxL1Tx:  10,
		ChainID:  chainID,
	}
	selectionConfig := &SelectionConfig{
		MaxL1UserTxs:      5,
		TxProcessorConfig: tpc,
	}
	// batch1
	l1UserTxs := []common.L1Tx{}
	_, _, _, _, _, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)

	expectedTxID0 := "0x028847b86613c0b70be18c8622119ed045b42e4e47d7938fa90bb3d1dc14928965" // 1st TransferToEthAddr
	expectedTxID1 := "0x0200b18773dcf56f770d65870fb02041cb59a088fd35b7c3f3df69f8a250b99a42" // 1st Exit
	expectedTxID2 := "0x029720ff506153f970f120ac638cd7ee759eeff2c2012e7634a78e4fdc05c04a90" // 2nd TransferToEthAddr

	// batch2
	// prepare the PoolL2Txs
	batchPoolL2 := `
	Type: PoolL2
	PoolTransferToEthAddr(0) A-B: 100 (126)
	PoolExit(0) B: 100 (126)`
	poolL2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	// add the PoolL2Txs to the l2DB
	addL2Txs(t, txsel, poolL2Txs)

	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err := txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 3, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 0, len(oL2Txs)) // should be 0 as the 2 PoolL2Txs does not have enough funds
	assert.Equal(t, 2, len(discardedL2Txs))
	assert.Equal(t, expectedTxID0, discardedL2Txs[0].TxID.String())
	assert.Equal(t, expectedTxID1, discardedL2Txs[1].TxID.String())
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)

	// as the PoolL2Txs have not been really processed, restart nonces
	tc.RestartNonces()

	// batch3
	// NOTE: this batch will result with 1 L2Tx, as the PoolExit tx is not
	// possible, as the PoolTransferToEthAddr is not processed yet when
	// checking availability of PoolExit.  This, in a near-future iteration
	// of the TxSelector will return the 2 transactions as valid and
	// selected, as the TxSelector will handle this kind of combinations.
	batchPoolL2 = `
	Type: PoolL2
	PoolTransferToEthAddr(0) A-B: 50 (126)`
	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, txsel, poolL2Txs)

	l1UserTxs = []common.L1Tx{}
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 1, len(oL2Txs)) // see 'NOTE' at the beginning of 'batch3' of this test
	assert.Equal(t, 2, len(discardedL2Txs))
	assert.Equal(t, expectedTxID2, oL2Txs[0].TxID.String())
	assert.Equal(t, expectedTxID0, discardedL2Txs[0].TxID.String())
	assert.Equal(t, expectedTxID1, discardedL2Txs[1].TxID.String())
	assert.Equal(t, common.TxTypeTransferToEthAddr, oL2Txs[0].Type)
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)

	// batch4
	// make the selection of another batch, which should include the
	// initial PoolExit, which now is valid as B has enough Balance
	l1UserTxs = []common.L1Tx{}
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 1, len(oL2Txs))
	assert.Equal(t, 1, len(discardedL2Txs))
	assert.Equal(t, expectedTxID1, oL2Txs[0].TxID.String()) // the Exit that was not accepted at the batch2
	assert.Equal(t, expectedTxID0, discardedL2Txs[0].TxID.String())
	assert.Equal(t, common.TxTypeExit, oL2Txs[0].Type)
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)
}

func TestTransferToBjj(t *testing.T) {
	set := `
		Type: Blockchain
		AddToken(1)

		CreateAccountDeposit(0) Coord: 0
		CreateAccountDeposit(0) A: 1000
		CreateAccountDeposit(0) B: 1000
		CreateAccountDeposit(1) B: 1000

		> batchL1 // freeze L1User{1}
		> batchL1 // forge L1User{1}
		> block
	`

	chainID := uint16(0)
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	assert.NoError(t, err)

	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	txsel := initTest(t, chainID, hermezContractAddr, tc.Users["Coord"])

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()

	addTokens(t, tc, txsel.l2db.DB())

	tpc := txprocessor.Config{
		NLevels:  16,
		MaxFeeTx: 10,
		MaxTx:    20,
		MaxL1Tx:  10,
		ChainID:  chainID,
	}
	selectionConfig := &SelectionConfig{
		MaxL1UserTxs:      5,
		TxProcessorConfig: tpc,
	}
	// batch1 to freeze L1UserTxs that will create some accounts with
	// positive balance
	l1UserTxs := []common.L1Tx{}
	_, _, _, _, _, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)

	// Transfer is ToBJJ to a BJJ-only account that doesn't exist
	// and the coordinator will create it via L1CoordTx.

	batchPoolL2 := `
	Type: PoolL2
	PoolTransferToBJJ(0) A-B: 50 (126)
	`
	poolL2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)

	// add the PoolL2Txs to the l2DB
	addL2Txs(t, txsel, poolL2Txs)

	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err := txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 4, len(oL1UserTxs))
	// We expect the coordinator to add an L1CoordTx to create an account for the recipient of the l2tx
	require.Equal(t, 1, len(oL1CoordTxs))
	assert.Equal(t, poolL2Txs[0].ToEthAddr, oL1CoordTxs[0].FromEthAddr)
	assert.Equal(t, poolL2Txs[0].ToBJJ, oL1CoordTxs[0].FromBJJ)
	// fmt.Printf("DBG l1CoordTx[0]: %+v\n", oL1CoordTxs[0])
	assert.Equal(t, 1, len(oL2Txs))
	assert.Equal(t, 0, len(discardedL2Txs))
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)

	// Now the BJJ-only account for B is already created, so the transfer
	// happens without an L1CoordTx that creates the user account.

	batchPoolL2 = `
	Type: PoolL2
	PoolTransferToBJJ(0) A-B: 50 (126)
	`

	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, txsel, poolL2Txs)

	l1UserTxs = []common.L1Tx{}
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	// Since the BJJ-only account B already exists, the coordinator doesn't add any L1CoordTxs
	assert.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 1, len(oL2Txs))
	assert.Equal(t, 0, len(discardedL2Txs))
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)

	// The transfer now is ToBJJ to a BJJ-only account that doesn't exist
	// and the coordinator will create it via L1CoordTx.   Since it's a
	// transfer of a token for which the coordinator doesn't have a fee
	// account, another L1CoordTx will be created for the coordinator to
	// receive the fees.

	batchPoolL2 = `
	Type: PoolL2
	PoolTransferToBJJ(1) B-A: 50 (126)
	`

	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	addL2Txs(t, txsel, poolL2Txs)

	l1UserTxs = []common.L1Tx{}
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(oL1UserTxs))
	// We expect the coordinator to add an L1CoordTx to create an account
	// to receive the fees by the coordinator and another one for the
	// recipient of the l2tx
	assert.Equal(t, 2, len(oL1CoordTxs))
	// [0] Coordinator account cration for token 1
	assert.Equal(t, tc.Users["Coord"].Addr, oL1CoordTxs[0].FromEthAddr)
	// [1] User A BJJ-only account creation for token 1
	assert.Equal(t, poolL2Txs[0].ToEthAddr, oL1CoordTxs[1].FromEthAddr)
	assert.Equal(t, poolL2Txs[0].ToBJJ, oL1CoordTxs[1].FromBJJ)
	assert.Equal(t, common.TokenID(1), oL1CoordTxs[1].TokenID)

	assert.Equal(t, 1, len(oL2Txs))
	assert.Equal(t, 0, len(discardedL2Txs))
	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)
}

func TestTransferManyFromSameAccount(t *testing.T) {
	set := `
		Type: Blockchain

		CreateAccountDeposit(0) Coord: 0
		CreateAccountDeposit(0) A: 1000
		CreateAccountDeposit(0) B: 1000

		> batchL1 // freeze L1User{1}
		> batchL1 // forge L1User{1}
		> block
	`

	chainID := uint16(0)
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	assert.NoError(t, err)

	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	txsel := initTest(t, chainID, hermezContractAddr, tc.Users["Coord"])

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()

	tpc := txprocessor.Config{
		NLevels:  16,
		MaxFeeTx: 10,
		MaxTx:    10,
		MaxL1Tx:  10,
		ChainID:  chainID,
	}
	selectionConfig := &SelectionConfig{
		MaxL1UserTxs:      5,
		TxProcessorConfig: tpc,
	}
	// batch1 to freeze L1UserTxs
	l1UserTxs := []common.L1Tx{}
	_, _, _, _, _, _, err = txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)

	// 8 transfers from the same account

	batchPoolL2 := `
	Type: PoolL2
	PoolTransfer(0) A-B: 10 (126) // 1
	PoolTransfer(0) A-B: 10 (126) // 2
	PoolTransfer(0) A-B: 10 (126) // 3
	PoolTransfer(0) A-B: 10 (126) // 4
	PoolTransfer(0) A-B: 10 (126) // 5
	PoolTransfer(0) A-B: 10 (126) // 6
	PoolTransfer(0) A-B: 10 (126) // 7
	PoolTransfer(0) A-B: 10 (126) // 8
	PoolTransfer(0) A-B: 10 (126) // 9
	PoolTransfer(0) A-B: 10 (126) // 10
	PoolTransfer(0) A-B: 10 (126) // 11
	`
	poolL2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)
	require.Equal(t, 11, len(poolL2Txs))

	// reorder poolL2Txs so that nonces are not sorted
	poolL2Txs[0], poolL2Txs[7] = poolL2Txs[7], poolL2Txs[0]
	poolL2Txs[1], poolL2Txs[10] = poolL2Txs[10], poolL2Txs[1]

	// add the PoolL2Txs to the l2DB
	addL2Txs(t, txsel, poolL2Txs)
	// batch 2 to crate some accounts with positive balance, and do 8 L2Tx transfers from account A
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	_, _, oL1UserTxs, oL1CoordTxs, oL2Txs, discardedL2Txs, err := txsel.GetL1L2TxSelection(selectionConfig, l1UserTxs)
	require.NoError(t, err)
	assert.Equal(t, 3, len(oL1UserTxs))
	require.Equal(t, 0, len(oL1CoordTxs))
	assert.Equal(t, 7, len(oL2Txs))
	assert.Equal(t, 1, len(discardedL2Txs))

	err = txsel.l2db.StartForging(common.TxIDsFromPoolL2Txs(oL2Txs), txsel.localAccountsDB.CurrentBatch())
	require.NoError(t, err)
}
