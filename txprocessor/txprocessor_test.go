package txprocessor

import (
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteme []string

func TestMain(m *testing.M) {
	exitVal := 0
	exitVal = m.Run()
	for _, dir := range deleteme {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
	os.Exit(exitVal)
}

func checkBalance(t *testing.T, tc *til.Context, statedb *statedb.StateDB, username string,
	tokenid int, expected string) {
	idx := tc.Users[username].Accounts[common.TokenID(tokenid)].Idx
	acc, err := statedb.GetAccount(idx)
	require.NoError(t, err)
	assert.Equal(t, expected, acc.Balance.String())
}

func checkBalanceByIdx(t *testing.T, statedb *statedb.StateDB, idx common.Idx, expected string) {
	acc, err := statedb.GetAccount(idx)
	require.NoError(t, err)
	assert.Equal(t, expected, acc.Balance.String())
}

func TestComputeEffectiveAmounts(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
	assert.NoError(t, err)

	set := `
		Type: Blockchain
		AddToken(1)
	
		CreateAccountDeposit(0) A: 10
		CreateAccountDeposit(0) B: 10
		CreateAccountDeposit(1) C: 10
		> batchL1
		> batchL1
		> block
	`
	chainID := uint16(0)
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)
	_, err = txProcessor.ProcessTxs(nil, blocks[0].Rollup.L1UserTxs, nil, nil)
	require.NoError(t, err)

	tx := common.L1Tx{
		FromIdx:       256,
		ToIdx:         257,
		Amount:        big.NewInt(10),
		DepositAmount: big.NewInt(0),
		FromEthAddr:   tc.Users["A"].Addr,
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(0), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(10), tx.EffectiveAmount)

	// expect error due not enough funds
	tx = common.L1Tx{
		FromIdx:       256,
		ToIdx:         257,
		Amount:        big.NewInt(11),
		DepositAmount: big.NewInt(0),
		FromEthAddr:   tc.Users["A"].Addr,
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(0), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// expect no-error as there are enough funds in a
	// CreateAccountDepositTransfer transction
	tx = common.L1Tx{
		FromIdx:       0,
		ToIdx:         257,
		Amount:        big.NewInt(10),
		DepositAmount: big.NewInt(10),
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(10), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(10), tx.EffectiveAmount)

	// expect error due not enough funds in a CreateAccountDepositTransfer
	// transction
	tx = common.L1Tx{
		FromIdx:       0,
		ToIdx:         257,
		Amount:        big.NewInt(11),
		DepositAmount: big.NewInt(10),
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(10), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// expect error due not same TokenID
	tx = common.L1Tx{
		FromIdx:       256,
		ToIdx:         258,
		Amount:        big.NewInt(5),
		DepositAmount: big.NewInt(0),
		FromEthAddr:   tc.Users["A"].Addr,
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(0), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// expect error due not same EthAddr
	tx = common.L1Tx{
		FromIdx:       256,
		ToIdx:         257,
		Amount:        big.NewInt(8),
		DepositAmount: big.NewInt(0),
		FromEthAddr:   tc.Users["B"].Addr,
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(0), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// expect on TxTypeDepositTransfer EffectiveAmount=0, but
	// EffectiveDepositAmount!=0, due not enough funds to make the transfer
	tx = common.L1Tx{
		FromIdx:       256,
		ToIdx:         257,
		Amount:        big.NewInt(20),
		DepositAmount: big.NewInt(8),
		FromEthAddr:   tc.Users["A"].Addr,
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(8), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// expect on TxTypeDepositTransfer EffectiveAmount=0, but
	// EffectiveDepositAmount!=0, due different EthAddr from FromIdx
	// address
	tx = common.L1Tx{
		FromIdx:       256,
		ToIdx:         257,
		Amount:        big.NewInt(8),
		DepositAmount: big.NewInt(8),
		FromEthAddr:   tc.Users["B"].Addr,
		UserOrigin:    true,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(8), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// CreateAccountDepositTransfer for TokenID=1 when receiver does not
	// have an account for that TokenID, expect that the
	// EffectiveDepositAmount=DepositAmount, but EffectiveAmount==0
	tx = common.L1Tx{
		FromIdx:       0,
		ToIdx:         257,
		Amount:        big.NewInt(8),
		DepositAmount: big.NewInt(8),
		FromEthAddr:   tc.Users["A"].Addr,
		TokenID:       2,
		UserOrigin:    true,
		Type:          common.TxTypeCreateAccountDepositTransfer,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(8), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	// DepositTransfer for TokenID=1 when receiver does not have an account
	// for that TokenID, expect that the
	// EffectiveDepositAmount=DepositAmount, but EffectiveAmount=0
	tx = common.L1Tx{
		FromIdx:       258,
		ToIdx:         256,
		Amount:        big.NewInt(8),
		DepositAmount: big.NewInt(8),
		FromEthAddr:   tc.Users["C"].Addr,
		TokenID:       1,
		UserOrigin:    true,
		Type:          common.TxTypeDepositTransfer,
	}
	txProcessor.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(8), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)

	statedb.Close()
}

func TestProcessTxsBalances(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
	assert.NoError(t, err)

	chainID := uint16(0)
	// generate test transactions from test.SetBlockchainMinimumFlow0 code
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(txsets.SetBlockchainMinimumFlow0)
	require.NoError(t, err)

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)

	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	_, err = txProcessor.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)
	assert.Equal(t, "0", txProcessor.state.MT.Root().BigInt().String())

	log.Debug("block:0 batch:2")
	l1UserTxs := []common.L1Tx{}
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = txProcessor.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, "0", txProcessor.state.MT.Root().BigInt().String())

	log.Debug("block:0 batch:3")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[2].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	_, err = txProcessor.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "A", 0, "500")
	assert.Equal(t,
		"10303926118213025243660668481827257778714122989909761705455084995854999537039",
		txProcessor.state.MT.Root().BigInt().String())

	log.Debug("block:0 batch:4")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[3].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[3].L2Txs)
	_, err = txProcessor.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[3].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "A", 0, "500")
	checkBalance(t, tc, statedb, "A", 1, "500")
	assert.Equal(t,
		"8530501758307821623834726627056947648600328521261384179220598288701741436285",
		txProcessor.state.MT.Root().BigInt().String())

	log.Debug("block:0 batch:5")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[4].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[4].L2Txs)
	_, err = txProcessor.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[4].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "A", 0, "500")
	checkBalance(t, tc, statedb, "A", 1, "500")
	assert.Equal(t,
		"8530501758307821623834726627056947648600328521261384179220598288701741436285",
		txProcessor.state.MT.Root().BigInt().String())

	log.Debug("block:0 batch:6")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[5].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[5].L2Txs)
	_, err = txProcessor.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[5].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "A", 0, "600")
	checkBalance(t, tc, statedb, "A", 1, "500")
	checkBalance(t, tc, statedb, "B", 0, "400")
	assert.Equal(t,
		"9061858435528794221929846392270405504056106238451760714188625065949729889651",
		txProcessor.state.MT.Root().BigInt().String())

	coordIdxs := []common.Idx{261, 263}
	log.Debug("block:0 batch:7")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[6].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[6].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[6].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "Coord", 0, "10")
	checkBalance(t, tc, statedb, "Coord", 1, "20")
	checkBalance(t, tc, statedb, "A", 0, "600")
	checkBalance(t, tc, statedb, "A", 1, "280")
	checkBalance(t, tc, statedb, "B", 0, "290")
	checkBalance(t, tc, statedb, "B", 1, "200")
	checkBalance(t, tc, statedb, "C", 0, "100")
	checkBalance(t, tc, statedb, "D", 0, "800")
	assert.Equal(t,
		"4392049343656836675348565048374261353937130287163762821533580216441778455298",
		txProcessor.state.MT.Root().BigInt().String())

	log.Debug("block:0 batch:8")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[7].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[7].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[7].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "Coord", 0, "35")
	checkBalance(t, tc, statedb, "Coord", 1, "30")
	checkBalance(t, tc, statedb, "A", 0, "430")
	checkBalance(t, tc, statedb, "A", 1, "280")
	checkBalance(t, tc, statedb, "B", 0, "390")
	checkBalance(t, tc, statedb, "B", 1, "90")
	checkBalance(t, tc, statedb, "C", 0, "45")
	checkBalance(t, tc, statedb, "C", 1, "100")
	checkBalance(t, tc, statedb, "D", 0, "800")
	assert.Equal(t,
		"8905191229562583213069132470917469035834300549892959854483573322676101624713",
		txProcessor.state.MT.Root().BigInt().String())

	coordIdxs = []common.Idx{262}
	log.Debug("block:1 batch:1")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[0].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, l1UserTxs, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "Coord", 1, "30")
	checkBalance(t, tc, statedb, "Coord", 0, "35")
	checkBalance(t, tc, statedb, "A", 0, "730")
	checkBalance(t, tc, statedb, "A", 1, "280")
	checkBalance(t, tc, statedb, "B", 0, "380")
	checkBalance(t, tc, statedb, "B", 1, "90")
	checkBalance(t, tc, statedb, "C", 0, "845")
	checkBalance(t, tc, statedb, "C", 1, "100")
	checkBalance(t, tc, statedb, "D", 0, "470")
	assert.Equal(t,
		"12063160053709941400160547588624831667157042937323422396363359123696668555050",
		txProcessor.state.MT.Root().BigInt().String())

	coordIdxs = []common.Idx{}
	log.Debug("block:1 batch:2")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, l1UserTxs, blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"20375835796927052406196249140510136992262283055544831070430919054949353249481",
		txProcessor.state.MT.Root().BigInt().String())

	// use Set of PoolL2 txs
	poolL2Txs, err := tc.GeneratePoolL2Txs(txsets.SetPoolL2MinimumFlow1)
	assert.NoError(t, err)

	_, err = txProcessor.ProcessTxs(coordIdxs, []common.L1Tx{}, []common.L1Tx{}, poolL2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, statedb, "Coord", 1, "30")
	checkBalance(t, tc, statedb, "Coord", 0, "35")
	checkBalance(t, tc, statedb, "A", 0, "510")
	checkBalance(t, tc, statedb, "A", 1, "170")
	checkBalance(t, tc, statedb, "B", 0, "480")
	checkBalance(t, tc, statedb, "B", 1, "190")
	checkBalance(t, tc, statedb, "C", 0, "845")
	checkBalance(t, tc, statedb, "C", 1, "100")
	checkBalance(t, tc, statedb, "D", 0, "360")
	checkBalance(t, tc, statedb, "F", 0, "100")

	statedb.Close()
}

func TestProcessTxsSynchronizer(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
	assert.NoError(t, err)

	chainID := uint16(0)
	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(txsets.SetBlockchain0)
	require.NoError(t, err)

	assert.Equal(t, 31, len(blocks[0].Rollup.L1UserTxs))
	assert.Equal(t, 4, len(blocks[0].Rollup.Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 0, len(blocks[0].Rollup.Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 22, len(blocks[0].Rollup.Batches[2].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Rollup.Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 65, len(blocks[1].Rollup.Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Rollup.Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[1].Rollup.Batches[1].L2Txs))

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  32,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	ptOut, err := txProcessor.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)
	assert.Equal(t, 4, len(ptOut.CreatedAccounts))
	assert.Equal(t, 0, len(ptOut.CollectedFees))

	log.Debug("block:0 batch:2")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = txProcessor.ProcessTxs(coordIdxs, blocks[0].Rollup.L1UserTxs,
		blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 31, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err := statedb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:3")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = txProcessor.ProcessTxs(coordIdxs, nil, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "2", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "1", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = statedb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	// before processing expect l2Txs[0:2].Nonce==0
	assert.Equal(t, nonce.Nonce(0), l2Txs[0].Nonce)
	assert.Equal(t, nonce.Nonce(0), l2Txs[1].Nonce)
	assert.Equal(t, nonce.Nonce(0), l2Txs[2].Nonce)

	// Idx of user 'X'
	idxX1 := tc.Users["X"].Accounts[common.TokenID(1)].Idx
	// Idx of user 'Y'
	idxY1 := tc.Users["Y"].Accounts[common.TokenID(1)].Idx
	// Idx of user 'Z'
	idxZ1 := tc.Users["Z"].Accounts[common.TokenID(1)].Idx
	accX1, err := statedb.GetAccount(idxX1)
	require.NoError(t, err)
	assert.Equal(t, "25000000000000000010", accX1.Balance.String())
	accY1, err := statedb.GetAccount(idxY1)
	require.NoError(t, err)
	assert.Equal(t, "25000000000000000010", accY1.Balance.String())
	accZ1, err := statedb.GetAccount(idxZ1)
	require.NoError(t, err)
	assert.Equal(t, "25000000000000000015", accZ1.Balance.String())

	ptOut, err = txProcessor.ProcessTxs(coordIdxs, nil, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	// after processing expect l2Txs[0:2].Nonce!=0 and has expected value
	assert.Equal(t, nonce.Nonce(5), l2Txs[0].Nonce)
	assert.Equal(t, nonce.Nonce(6), l2Txs[1].Nonce)
	assert.Equal(t, nonce.Nonce(7), l2Txs[2].Nonce)
	// the 'ForceExit(1)' is not computed yet, as the batch is without L1UserTxs
	assert.Equal(t, 4, len(ptOut.ExitInfos))
	assert.Equal(t, 1, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "1", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	assert.Equal(t, big.NewInt(5), ptOut.ExitInfos[0].Balance)
	assert.Equal(t, big.NewInt(13), ptOut.ExitInfos[1].Balance)
	assert.Equal(t, big.NewInt(12), ptOut.ExitInfos[2].Balance)
	assert.Equal(t, big.NewInt(16), ptOut.ExitInfos[3].Balance)

	acc, err = statedb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	accX1, err = statedb.GetAccount(idxX1)
	require.NoError(t, err)
	assert.Equal(t, "24999999999999999976", accX1.Balance.String())
	accY1, err = statedb.GetAccount(idxY1)
	require.NoError(t, err)
	assert.Equal(t, "25000000000000000004", accY1.Balance.String())
	accZ1, err = statedb.GetAccount(idxZ1)
	require.NoError(t, err)
	assert.Equal(t, "25000000000000000024", accZ1.Balance.String())

	log.Debug("block:1 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	ptOut, err = txProcessor.ProcessTxs(coordIdxs, blocks[1].Rollup.L1UserTxs,
		blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	// 1, as previous batch was without L1UserTxs, and has pending the
	// 'ForceExit(1) A: 5', and the 2 exit transactions get grouped under 1
	// ExitInfo
	assert.Equal(t, 1, len(ptOut.ExitInfos))
	assert.Equal(t, 1, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = statedb.GetAccount(idxA1)
	assert.NoError(t, err)
	assert.Equal(t, "77", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = statedb.GetAccount(idxB0)
	require.NoError(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = statedb.GetAccount(common.Idx(256))
	require.NoError(t, err)
	assert.Equal(t, "2", acc.Balance.String())

	statedb.Close()
}

func TestProcessTxsBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: 32})
	assert.NoError(t, err)

	chainID := uint16(0)
	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(txsets.SetBlockchain0)
	require.NoError(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  32,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	ptOut, err := txProcessor.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)
	// expect 0 at CreatedAccount, as is only computed when StateDB.Type==TypeSynchronizer
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))

	log.Debug("block:0 batch:2")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = txProcessor.ProcessTxs(coordIdxs, blocks[0].Rollup.L1UserTxs,
		blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err := statedb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:3")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = txProcessor.ProcessTxs(coordIdxs, nil, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err = statedb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, nil, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	acc, err = statedb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	log.Debug("block:1 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, blocks[1].Rollup.L1UserTxs,
		blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	acc, err = statedb.GetAccount(idxA1)
	assert.NoError(t, err)
	assert.Equal(t, "77", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = statedb.GetAccount(idxB0)
	require.NoError(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = statedb.GetAccount(common.Idx(256))
	require.NoError(t, err)
	assert.Equal(t, common.TokenID(0), acc.TokenID)
	assert.Equal(t, "2", acc.Balance.String())
	acc, err = statedb.GetAccount(common.Idx(257))
	require.NoError(t, err)
	assert.Equal(t, common.TokenID(1), acc.TokenID)
	assert.Equal(t, "2", acc.Balance.String())

	assert.Equal(t,
		"8499500340673457131709907313180428395258466720027480159049632090608270570263",
		statedb.MT.Root().BigInt().String())

	statedb.Close()
}

func TestProcessTxsRootTestVectors(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: 32})
	assert.NoError(t, err)

	// same values than in the js test
	bjj0, err := common.BJJFromStringWithChecksum(
		"21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7")
	assert.NoError(t, err)
	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       bjj0,
			FromEthAddr:   ethCommon.HexToAddress("0x7e5f4552091a69125d5dfcb7b8c2659029395bdf"),
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

	chainID := uint16(0)
	config := Config{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)
	_, err = txProcessor.ProcessTxs(nil, l1Txs, nil, l2Txs)
	require.NoError(t, err)
	assert.Equal(t,
		"16181420716631932805604732887923905079487577323947343079740042260791593140221",
		statedb.MT.Root().BigInt().String())

	statedb.Close()
}

func TestCreateAccountDepositMaxValue(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	nLevels := 16
	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: nLevels})
	assert.NoError(t, err)

	users := txsets.GenerateJsUsers(t)

	daMaxF40 := common.Float40(0xFFFFFFFFFF)
	daMaxBI, err := daMaxF40.BigInt()
	require.NoError(t, err)
	assert.Equal(t, "343597383670000000000000000000000000000000", daMaxBI.String())

	daMax1F40 := common.Float40(0xFFFFFFFFFE)
	require.NoError(t, err)
	daMax1BI, err := daMax1F40.BigInt()
	require.NoError(t, err)
	assert.Equal(t, "343597383660000000000000000000000000000000", daMax1BI.String())

	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: daMaxBI,
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: daMax1BI,
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}

	chainID := uint16(0)
	config := Config{
		NLevels:  uint32(nLevels),
		MaxTx:    3,
		MaxL1Tx:  2,
		MaxFeeTx: 2,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)

	_, err = txProcessor.ProcessTxs(nil, l1Txs, nil, nil)
	require.NoError(t, err)

	// check balances
	acc, err := statedb.GetAccount(common.Idx(256))
	require.NoError(t, err)
	assert.Equal(t, daMaxBI, acc.Balance)
	acc, err = statedb.GetAccount(common.Idx(257))
	require.NoError(t, err)
	assert.Equal(t, daMax1BI, acc.Balance)

	statedb.Close()
}

func initTestMultipleCoordIdxForTokenID(t *testing.T) (*TxProcessor, *til.Context,
	[]common.BlockData, *statedb.StateDB) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: 32})
	assert.NoError(t, err)

	chainID := uint16(1)

	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)

	set := `
		Type: Blockchain

		CreateAccountDeposit(0) A: 200

		> batchL1 // freeze L1User{1}

		CreateAccountCoordinator(0) Coord
		CreateAccountCoordinator(0) B

		Transfer(0) A-B: 100 (126)

		> batchL1 // forge L1User{1}, forge L1Coord{4}, forge L2{2}
		> block
	`
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)
	// batch1
	_, err = txProcessor.ProcessTxs(nil, nil, nil, nil) // to simulate the first batch from the Til set
	require.NoError(t, err)

	return txProcessor, tc, blocks, statedb
}

func TestMultipleCoordIdxForTokenID(t *testing.T) {
	// Check that ProcessTxs always uses the first occurrence of the
	// CoordIdx for each TokenID

	coordIdxs := []common.Idx{257, 257, 257}
	txProcessor, tc, blocks, statedb := initTestMultipleCoordIdxForTokenID(t)
	l1UserTxs := til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l1CoordTxs := blocks[0].Rollup.Batches[1].L1CoordinatorTxs
	l1CoordTxs = append(l1CoordTxs, l1CoordTxs[0]) // duplicate the CoordAccount for TokenID=0
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err := txProcessor.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	checkBalanceByIdx(t, txProcessor.state, 256, "90")  // A
	checkBalanceByIdx(t, txProcessor.state, 257, "10")  // Coord0
	checkBalanceByIdx(t, txProcessor.state, 258, "100") // B
	checkBalanceByIdx(t, txProcessor.state, 259, "0")   // Coord0

	// reset StateDB values
	coordIdxs = []common.Idx{259, 257}
	statedb.Close()
	txProcessor, tc, blocks, statedb = initTestMultipleCoordIdxForTokenID(t)
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l1CoordTxs = blocks[0].Rollup.Batches[1].L1CoordinatorTxs
	l1CoordTxs = append(l1CoordTxs, l1CoordTxs[0]) // duplicate the CoordAccount for TokenID=0
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	checkBalanceByIdx(t, txProcessor.state, 256, "90")  // A
	checkBalanceByIdx(t, txProcessor.state, 257, "0")   // Coord0
	checkBalanceByIdx(t, txProcessor.state, 258, "100") // B
	checkBalanceByIdx(t, txProcessor.state, 259, "10")  // Coord0

	// reset StateDB values
	coordIdxs = []common.Idx{257, 259}
	statedb.Close()
	txProcessor, tc, blocks, statedb = initTestMultipleCoordIdxForTokenID(t)
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l1CoordTxs = blocks[0].Rollup.Batches[1].L1CoordinatorTxs
	l1CoordTxs = append(l1CoordTxs, l1CoordTxs[0]) // duplicate the CoordAccount for TokenID=0
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = txProcessor.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	checkBalanceByIdx(t, txProcessor.state, 256, "90")  // A
	checkBalanceByIdx(t, txProcessor.state, 257, "10")  // Coord0
	checkBalanceByIdx(t, txProcessor.state, 258, "100") // B
	checkBalanceByIdx(t, txProcessor.state, 259, "0")   // Coord0

	statedb.Close()
}

func testTwoExits(t *testing.T, stateDBType statedb.TypeStateDB) ([]*ProcessTxOutput,
	[]*ProcessTxOutput, []*ProcessTxOutput) {
	// In the first part we generate a batch with two force exits for the
	// same account of 20 each.  The txprocessor output should be a single
	// exitInfo with balance of 40.
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	nLevels := 16
	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: stateDBType, NLevels: nLevels})
	assert.NoError(t, err)

	chainID := uint16(1)

	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)

	// Two exits for the same account.  The tx processor should output a
	// single exit with the accumulated exit balance
	set := `
		Type: Blockchain

		CreateAccountDeposit(0) A: 100

		> batchL1 // freeze L1User{1}
		> batchL1 // forge L1User{1}

		ForceExit(0) A: 20
		ForceExit(0) A: 20

		> batchL1 // freeze L1User{2}
		> batchL1 // forge L1User{2}
		> block
	`
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &til.ConfigExtra{})
	require.NoError(t, err)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	// Sanity check
	require.Equal(t, 1, len(blocks[0].Rollup.Batches[1].L1UserTxs))
	require.Equal(t, 2, len(blocks[0].Rollup.Batches[3].L1UserTxs))

	config := Config{
		NLevels:  uint32(nLevels),
		MaxTx:    3,
		MaxL1Tx:  2,
		MaxFeeTx: 2,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(sdb, config)
	ptOuts := []*ProcessTxOutput{}
	for _, block := range blocks {
		for _, batch := range block.Rollup.Batches {
			ptOut, err := txProcessor.ProcessTxs(nil, batch.L1UserTxs, nil, nil)
			require.NoError(t, err)
			ptOuts = append(ptOuts, ptOut)
		}
	}

	acc, err := sdb.GetAccount(256)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(60), acc.Balance)

	// In the second part we start a fresh statedb and generate a batch
	// with one force exit for the same account as before.  The txprocessor
	// output should be a single exitInfo with balance of 40, and the exit
	// merkle tree proof should be equal to the previous one.

	dir2, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir2)

	sdb2, err := statedb.NewStateDB(statedb.Config{Path: dir2, Keep: 128,
		Type: stateDBType, NLevels: nLevels})
	assert.NoError(t, err)

	tc = til.NewContext(chainID, common.RollupConstMaxL1UserTx)

	// Single exit with balance of both exits in previous set.  The exit
	// root should match.
	set2 := `
		Type: Blockchain

		CreateAccountDeposit(0) A: 100

		> batchL1 // freeze L1User{1}
		> batchL1 // forge L1User{1}

		ForceExit(0) A: 40

		> batchL1 // freeze L1User{2}
		> batchL1 // forge L1User{2}
		> block
	`
	blocks, err = tc.GenerateBlocks(set2)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &til.ConfigExtra{})
	require.NoError(t, err)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	txProcessor = NewTxProcessor(sdb2, config)
	ptOuts2 := []*ProcessTxOutput{}
	for _, block := range blocks {
		for _, batch := range block.Rollup.Batches {
			ptOut, err := txProcessor.ProcessTxs(nil, batch.L1UserTxs, nil, nil)
			require.NoError(t, err)
			ptOuts2 = append(ptOuts2, ptOut)
		}
	}

	// In the third part we start a fresh statedb and generate a batch with
	// two force exit for the same account as before but where the 1st Exit
	// is with all the amount, and the 2nd Exit is with more amount than
	// the available balance.  The txprocessor output should be a single
	// exitInfo with balance of 40, and the exit merkle tree proof should
	// be equal to the previous ones.

	dir3, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir3)

	sdb3, err := statedb.NewStateDB(statedb.Config{Path: dir3, Keep: 128,
		Type: stateDBType, NLevels: nLevels})
	assert.NoError(t, err)

	tc = til.NewContext(chainID, common.RollupConstMaxL1UserTx)

	// Single exit with balance of both exits in previous set.  The exit
	// root should match.
	set3 := `
		Type: Blockchain

		CreateAccountDeposit(0) A: 100

		> batchL1 // freeze L1User{1}
		> batchL1 // forge L1User{1}

		ForceExit(0) A: 40
		ForceExit(0) A: 100

		> batchL1 // freeze L1User{2}
		> batchL1 // forge L1User{2}
		> block
	`
	blocks, err = tc.GenerateBlocks(set3)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &til.ConfigExtra{})
	require.NoError(t, err)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	txProcessor = NewTxProcessor(sdb3, config)
	ptOuts3 := []*ProcessTxOutput{}
	for _, block := range blocks {
		for _, batch := range block.Rollup.Batches {
			ptOut, err := txProcessor.ProcessTxs(nil, batch.L1UserTxs, nil, nil)
			require.NoError(t, err)
			ptOuts3 = append(ptOuts3, ptOut)
		}
	}

	sdb3.Close()
	sdb2.Close()
	sdb.Close()

	return ptOuts, ptOuts2, ptOuts3
}

func TestTwoExitsSynchronizer(t *testing.T) {
	ptOuts, ptOuts2, ptOuts3 := testTwoExits(t, statedb.TypeSynchronizer)

	assert.Equal(t, 1, len(ptOuts[3].ExitInfos))
	assert.Equal(t, big.NewInt(40), ptOuts[3].ExitInfos[0].Balance)

	assert.Equal(t, ptOuts[3].ExitInfos[0].MerkleProof, ptOuts2[3].ExitInfos[0].MerkleProof)
	assert.Equal(t, ptOuts[3].ExitInfos[0].MerkleProof, ptOuts3[3].ExitInfos[0].MerkleProof)
}

func TestExitOf0Amount(t *testing.T) {
	// Test to check that when doing an Exit with amount 0 the Exit Root
	// does not change (as there is no new Exit Leaf created)

	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: 32})
	assert.NoError(t, err)

	chainID := uint16(1)

	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)

	set := `
		Type: Blockchain

		CreateAccountDeposit(0) A: 100
		CreateAccountDeposit(0) B: 100

		> batchL1 // batch1: freeze L1User{2}
		> batchL1 // batch2: forge L1User{2}

		ForceExit(0) A: 10
		ForceExit(0) B: 0

		> batchL1 // batch3: freeze L1User{2}
		> batchL1 // batch4: forge L1User{2}

		ForceExit(0) A: 10

		> batchL1 // batch5: freeze L1User{1}
		> batchL1 // batch6: forge L1User{1}

		ForceExit(0) A: 0
		> batchL1 // batch7: freeze L1User{1}
		> batchL1 // batch8: forge L1User{1}
		> block
	`
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &til.ConfigExtra{})
	require.NoError(t, err)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	// Sanity check
	require.Equal(t, 2, len(blocks[0].Rollup.Batches[1].L1UserTxs))
	require.Equal(t, 2, len(blocks[0].Rollup.Batches[3].L1UserTxs))
	require.Equal(t, big.NewInt(10), blocks[0].Rollup.Batches[3].L1UserTxs[0].Amount)
	require.Equal(t, big.NewInt(0), blocks[0].Rollup.Batches[3].L1UserTxs[1].Amount)

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)

	// For this test are only processed the batches with transactions:
	// - Batch2, equivalent to Batches[1]
	// - Batch4, equivalent to Batches[3]
	// - Batch6, equivalent to Batches[5]
	// - Batch8, equivalent to Batches[7]

	// process Batch2:
	_, err = txProcessor.ProcessTxs(nil, blocks[0].Rollup.Batches[1].L1UserTxs, nil, nil)
	require.NoError(t, err)
	// process Batch4:
	ptOut, err := txProcessor.ProcessTxs(nil, blocks[0].Rollup.Batches[3].L1UserTxs, nil, nil)
	require.NoError(t, err)
	assert.Equal(t,
		"17688031540912620894848983912708704736922099609001460827147265569563156468242",
		ptOut.ZKInputs.Metadata.NewExitRootRaw.BigInt().String())
	exitRootBatch4 := ptOut.ZKInputs.Metadata.NewExitRootRaw.BigInt().String()

	// process Batch6:
	ptOut, err = txProcessor.ProcessTxs(nil, blocks[0].Rollup.Batches[5].L1UserTxs, nil, nil)
	require.NoError(t, err)
	assert.Equal(t,
		"17688031540912620894848983912708704736922099609001460827147265569563156468242",
		ptOut.ZKInputs.Metadata.NewExitRootRaw.BigInt().String())
	// Expect that the ExitRoot for the Batch6 will be equal than for the
	// Batch4, as the Batch4 & Batch6 have the same tx with Exit Amount=10,
	// and Batch4 has a 2nd tx with Exit Amount=0.
	assert.Equal(t, exitRootBatch4, ptOut.ZKInputs.Metadata.NewExitRootRaw.BigInt().String())

	// For the Batch8, as there is only 1 exit with Amount=0, the ExitRoot
	// should be 0.
	// process Batch8:
	ptOut, err = txProcessor.ProcessTxs(nil, blocks[0].Rollup.Batches[7].L1UserTxs, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "0", ptOut.ZKInputs.Metadata.NewExitRootRaw.BigInt().String())

	statedb.Close()
}

func TestUpdatedAccounts(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	statedb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
	assert.NoError(t, err)

	set := `
		Type: Blockchain
		AddToken(1)
		CreateAccountCoordinator(0) Coord // 256
		CreateAccountCoordinator(1) Coord // 257
		> batch // 1
		CreateAccountDeposit(0) A: 50 // 258
		CreateAccountDeposit(0) B: 60 // 259
		CreateAccountDeposit(1) A: 70 // 260
		CreateAccountDeposit(1) B: 80 // 261
		> batchL1 // 2
		> batchL1 // 3
		Transfer(0) A-B: 5 (126)
		> batch // 4
		Exit(1) B: 5 (126)
		> batch // 5
		> block
	`

	chainID := uint16(0)
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "Coord",
	}
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	tc.FillBlocksL1UserTxsBatchNum(blocks)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	require.Equal(t, 5, len(blocks[0].Rollup.Batches))

	config := Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	txProcessor := NewTxProcessor(statedb, config)

	sortedKeys := func(m map[common.Idx]*common.Account) []int {
		keys := make([]int, 0)
		for k := range m {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		return keys
	}

	for _, batch := range blocks[0].Rollup.Batches {
		l2Txs := common.L2TxsToPoolL2Txs(batch.L2Txs)
		ptOut, err := txProcessor.ProcessTxs(batch.Batch.FeeIdxsCoordinator, batch.L1UserTxs,
			batch.L1CoordinatorTxs, l2Txs)
		require.NoError(t, err)
		switch batch.Batch.BatchNum {
		case 1:
			assert.Equal(t, 2, len(ptOut.UpdatedAccounts))
			assert.Equal(t, []int{256, 257}, sortedKeys(ptOut.UpdatedAccounts))
		case 2:
			assert.Equal(t, 0, len(ptOut.UpdatedAccounts))
			assert.Equal(t, []int{}, sortedKeys(ptOut.UpdatedAccounts))
		case 3:
			assert.Equal(t, 4, len(ptOut.UpdatedAccounts))
			assert.Equal(t, []int{258, 259, 260, 261}, sortedKeys(ptOut.UpdatedAccounts))
		case 4:
			assert.Equal(t, 2+1, len(ptOut.UpdatedAccounts))
			assert.Equal(t, []int{256, 258, 259}, sortedKeys(ptOut.UpdatedAccounts))
		case 5:
			assert.Equal(t, 1+1, len(ptOut.UpdatedAccounts))
			assert.Equal(t, []int{257, 261}, sortedKeys(ptOut.UpdatedAccounts))
		}
		for idx, updAcc := range ptOut.UpdatedAccounts {
			acc, err := statedb.GetAccount(idx)
			require.NoError(t, err)
			// If acc.Balance is 0, set it to 0 with big.NewInt so
			// that the comparison succeeds.  Without this, the
			// comparison will not succeed because acc.Balance is
			// set from a slice, and thus the internal big.Int
			// buffer is not nil (big.Int.abs)
			if acc.Balance.BitLen() == 0 {
				acc.Balance = big.NewInt(0)
			}
			assert.Equal(t, acc, updAcc)
		}
	}

	statedb.Close()
}
