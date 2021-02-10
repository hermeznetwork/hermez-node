package txprocessor

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkBalance(t *testing.T, tc *til.Context, sdb *statedb.StateDB, username string, tokenid int, expected string) {
	idx := tc.Users[username].Accounts[common.TokenID(tokenid)].Idx
	acc, err := sdb.GetAccount(idx)
	require.NoError(t, err)
	assert.Equal(t, expected, acc.Balance.String())
}

func checkBalanceByIdx(t *testing.T, sdb *statedb.StateDB, idx common.Idx, expected string) {
	acc, err := sdb.GetAccount(idx)
	require.NoError(t, err)
	assert.Equal(t, expected, acc.Balance.String())
}

func TestComputeEffectiveAmounts(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
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
	tp := NewTxProcessor(sdb, config)
	_, err = tp.ProcessTxs(nil, blocks[0].Rollup.L1UserTxs, nil, nil)
	require.NoError(t, err)

	tx := common.L1Tx{
		FromIdx:       256,
		ToIdx:         257,
		Amount:        big.NewInt(10),
		DepositAmount: big.NewInt(0),
		FromEthAddr:   tc.Users["A"].Addr,
		UserOrigin:    true,
	}
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
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
	tp.computeEffectiveAmounts(&tx)
	assert.Equal(t, big.NewInt(8), tx.EffectiveDepositAmount)
	assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)
}

func TestProcessTxsBalances(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
	assert.NoError(t, err)

	chainID := uint16(0)
	// generate test transactions from test.SetBlockchain0 code
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
	tp := NewTxProcessor(sdb, config)

	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	_, err = tp.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)
	assert.Equal(t, "0", tp.s.MT.Root().BigInt().String())

	log.Debug("block:0 batch:2")
	l1UserTxs := []common.L1Tx{}
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, "0", tp.s.MT.Root().BigInt().String())

	log.Debug("block:0 batch:3")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[2].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	_, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "A", 0, "500")
	assert.Equal(t, "13644148972047617726265275926674266298636745191961029124811988256139761111521", tp.s.MT.Root().BigInt().String())

	log.Debug("block:0 batch:4")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[3].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[3].L2Txs)
	_, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[3].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "A", 0, "500")
	checkBalance(t, tc, sdb, "A", 1, "500")
	assert.Equal(t, "12433441613247342495680642890662773367605896324555599297255745922589338651261", tp.s.MT.Root().BigInt().String())

	log.Debug("block:0 batch:5")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[4].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[4].L2Txs)
	_, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[4].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "A", 0, "500")
	checkBalance(t, tc, sdb, "A", 1, "500")
	assert.Equal(t, "12433441613247342495680642890662773367605896324555599297255745922589338651261", tp.s.MT.Root().BigInt().String())

	log.Debug("block:0 batch:6")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[5].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[5].L2Txs)
	_, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[5].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "A", 0, "600")
	checkBalance(t, tc, sdb, "A", 1, "500")
	checkBalance(t, tc, sdb, "B", 0, "400")
	assert.Equal(t, "4191361650490017591061467288209836928064232431729236465872209988325272262963", tp.s.MT.Root().BigInt().String())

	coordIdxs := []common.Idx{261, 262}
	log.Debug("block:0 batch:7")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[6].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[6].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[6].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "10")
	checkBalance(t, tc, sdb, "Coord", 1, "20")
	checkBalance(t, tc, sdb, "A", 0, "600")
	checkBalance(t, tc, sdb, "A", 1, "280")
	checkBalance(t, tc, sdb, "B", 0, "290")
	checkBalance(t, tc, sdb, "B", 1, "200")
	checkBalance(t, tc, sdb, "C", 0, "100")
	checkBalance(t, tc, sdb, "D", 0, "800")
	assert.Equal(t, "7614010373759339299470010949167613050707822522530721724565424494781010548240", tp.s.MT.Root().BigInt().String())

	log.Debug("block:0 batch:8")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[7].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[7].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[7].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "35")
	checkBalance(t, tc, sdb, "Coord", 1, "30")
	checkBalance(t, tc, sdb, "A", 0, "430")
	checkBalance(t, tc, sdb, "A", 1, "280")
	checkBalance(t, tc, sdb, "B", 0, "390")
	checkBalance(t, tc, sdb, "B", 1, "90")
	checkBalance(t, tc, sdb, "C", 0, "45")
	checkBalance(t, tc, sdb, "C", 1, "100")
	checkBalance(t, tc, sdb, "D", 0, "800")
	assert.Equal(t, "21231789250434471575486264439945776732824482207853465397552873521865656677689", tp.s.MT.Root().BigInt().String())

	coordIdxs = []common.Idx{262}
	log.Debug("block:1 batch:1")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[0].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "75")
	checkBalance(t, tc, sdb, "Coord", 1, "30")
	checkBalance(t, tc, sdb, "A", 0, "730")
	checkBalance(t, tc, sdb, "A", 1, "280")
	checkBalance(t, tc, sdb, "B", 0, "380")
	checkBalance(t, tc, sdb, "B", 1, "90")
	checkBalance(t, tc, sdb, "C", 0, "845")
	checkBalance(t, tc, sdb, "C", 1, "100")
	checkBalance(t, tc, sdb, "D", 0, "470")
	assert.Equal(t, "11289313644810782435120113035387729451095637380468777086895109386127538554246", tp.s.MT.Root().BigInt().String())

	coordIdxs = []common.Idx{}
	log.Debug("block:1 batch:2")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, "10342681351319338354912862547249967104198317571995055517008223832276478908482", tp.s.MT.Root().BigInt().String())

	// use Set of PoolL2 txs
	poolL2Txs, err := tc.GeneratePoolL2Txs(txsets.SetPoolL2MinimumFlow1)
	assert.NoError(t, err)

	_, err = tp.ProcessTxs(coordIdxs, []common.L1Tx{}, []common.L1Tx{}, poolL2Txs)
	require.NoError(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "75")
	checkBalance(t, tc, sdb, "Coord", 1, "30")
	checkBalance(t, tc, sdb, "A", 0, "510")
	checkBalance(t, tc, sdb, "A", 1, "170")
	checkBalance(t, tc, sdb, "B", 0, "480")
	checkBalance(t, tc, sdb, "B", 1, "190")
	checkBalance(t, tc, sdb, "C", 0, "845")
	checkBalance(t, tc, sdb, "C", 1, "100")
	checkBalance(t, tc, sdb, "D", 0, "360")
	checkBalance(t, tc, sdb, "F", 0, "100")
}

func TestProcessTxsSynchronizer(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
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
	assert.Equal(t, 62, len(blocks[1].Rollup.Batches[0].L2Txs))
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
	tp := NewTxProcessor(sdb, config)

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	ptOut, err := tp.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)
	assert.Equal(t, 4, len(ptOut.CreatedAccounts))
	assert.Equal(t, 0, len(ptOut.CollectedFees))

	log.Debug("block:0 batch:2")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = tp.ProcessTxs(coordIdxs, blocks[0].Rollup.L1UserTxs,
		blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 31, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err := sdb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:3")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = tp.ProcessTxs(coordIdxs, nil, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "2", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "1", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = sdb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	// before processing expect l2Txs[0:2].Nonce==0
	assert.Equal(t, common.Nonce(0), l2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(0), l2Txs[1].Nonce)
	assert.Equal(t, common.Nonce(0), l2Txs[2].Nonce)

	ptOut, err = tp.ProcessTxs(coordIdxs, nil, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	// after processing expect l2Txs[0:2].Nonce!=0 and has expected value
	assert.Equal(t, common.Nonce(5), l2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(6), l2Txs[1].Nonce)
	assert.Equal(t, common.Nonce(7), l2Txs[2].Nonce)

	assert.Equal(t, 4, len(ptOut.ExitInfos)) // the 'ForceExit(1)' is not computed yet, as the batch is without L1UserTxs
	assert.Equal(t, 1, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "1", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = sdb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	log.Debug("block:1 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	ptOut, err = tp.ProcessTxs(coordIdxs, blocks[1].Rollup.L1UserTxs,
		blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	assert.Equal(t, 1, len(ptOut.ExitInfos)) // 1, as previous batch was without L1UserTxs, and has pending the 'ForceExit(1) A: 5', and the 2 exit transactions get grouped under 1 ExitInfo
	assert.Equal(t, 1, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = sdb.GetAccount(idxA1)
	assert.NoError(t, err)
	assert.Equal(t, "77", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = sdb.GetAccount(idxB0)
	require.NoError(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = sdb.GetAccount(common.Idx(256))
	require.NoError(t, err)
	assert.Equal(t, "2", acc.Balance.String())
}

func TestProcessTxsBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
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
	tp := NewTxProcessor(sdb, config)

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	ptOut, err := tp.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)
	// expect 0 at CreatedAccount, as is only computed when StateDB.Type==TypeSynchronizer
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))

	log.Debug("block:0 batch:2")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = tp.ProcessTxs(coordIdxs, blocks[0].Rollup.L1UserTxs, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err := sdb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:3")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = tp.ProcessTxs(coordIdxs, nil, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err = sdb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, nil, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	acc, err = sdb.GetAccount(idxA1)
	require.NoError(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	log.Debug("block:1 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, blocks[1].Rollup.L1UserTxs, blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)
	acc, err = sdb.GetAccount(idxA1)
	assert.NoError(t, err)
	assert.Equal(t, "77", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = sdb.GetAccount(idxB0)
	require.NoError(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = sdb.GetAccount(common.Idx(256))
	require.NoError(t, err)
	assert.Equal(t, common.TokenID(0), acc.TokenID)
	assert.Equal(t, "2", acc.Balance.String())
	acc, err = sdb.GetAccount(common.Idx(257))
	require.NoError(t, err)
	assert.Equal(t, common.TokenID(1), acc.TokenID)
	assert.Equal(t, "2", acc.Balance.String())

	assert.Equal(t, "18894163991492573893706613133132363559300580460789469708968288074813925659539", sdb.MT.Root().BigInt().String())
}

func TestProcessTxsRootTestVectors(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: 32})
	assert.NoError(t, err)

	// same values than in the js test
	bjj0, err := common.BJJFromStringWithChecksum("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7")
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
	tp := NewTxProcessor(sdb, config)
	_, err = tp.ProcessTxs(nil, l1Txs, nil, l2Txs)
	require.NoError(t, err)
	assert.Equal(t, "9827704113668630072730115158977131501210702363656902211840117643154933433410", sdb.MT.Root().BigInt().String())
}

func TestCreateAccountDepositMaxValue(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	nLevels := 16
	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
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
	tp := NewTxProcessor(sdb, config)

	_, err = tp.ProcessTxs(nil, l1Txs, nil, nil)
	require.NoError(t, err)

	// check balances
	acc, err := sdb.GetAccount(common.Idx(256))
	require.NoError(t, err)
	assert.Equal(t, daMaxBI, acc.Balance)
	acc, err = sdb.GetAccount(common.Idx(257))
	require.NoError(t, err)
	assert.Equal(t, daMax1BI, acc.Balance)
}

func initTestMultipleCoordIdxForTokenID(t *testing.T) (*TxProcessor, *til.Context, []common.BlockData) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
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
	tp := NewTxProcessor(sdb, config)
	// batch1
	_, err = tp.ProcessTxs(nil, nil, nil, nil) // to simulate the first batch from the Til set
	require.NoError(t, err)

	return tp, tc, blocks
}

func TestMultipleCoordIdxForTokenID(t *testing.T) {
	// Check that ProcessTxs always uses the first occurrence of the
	// CoordIdx for each TokenID

	coordIdxs := []common.Idx{257, 257, 257}
	tp, tc, blocks := initTestMultipleCoordIdxForTokenID(t)
	l1UserTxs := til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l1CoordTxs := blocks[0].Rollup.Batches[1].L1CoordinatorTxs
	l1CoordTxs = append(l1CoordTxs, l1CoordTxs[0]) // duplicate the CoordAccount for TokenID=0
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	checkBalanceByIdx(t, tp.s, 256, "90")  // A
	checkBalanceByIdx(t, tp.s, 257, "10")  // Coord0
	checkBalanceByIdx(t, tp.s, 258, "100") // B
	checkBalanceByIdx(t, tp.s, 259, "0")   // Coord0

	// reset StateDB values
	coordIdxs = []common.Idx{259, 257}
	tp, tc, blocks = initTestMultipleCoordIdxForTokenID(t)
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l1CoordTxs = blocks[0].Rollup.Batches[1].L1CoordinatorTxs
	l1CoordTxs = append(l1CoordTxs, l1CoordTxs[0]) // duplicate the CoordAccount for TokenID=0
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	checkBalanceByIdx(t, tp.s, 256, "90")  // A
	checkBalanceByIdx(t, tp.s, 257, "0")   // Coord0
	checkBalanceByIdx(t, tp.s, 258, "100") // B
	checkBalanceByIdx(t, tp.s, 259, "10")  // Coord0

	// reset StateDB values
	coordIdxs = []common.Idx{257, 259}
	tp, tc, blocks = initTestMultipleCoordIdxForTokenID(t)
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l1CoordTxs = blocks[0].Rollup.Batches[1].L1CoordinatorTxs
	l1CoordTxs = append(l1CoordTxs, l1CoordTxs[0]) // duplicate the CoordAccount for TokenID=0
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	checkBalanceByIdx(t, tp.s, 256, "90")  // A
	checkBalanceByIdx(t, tp.s, 257, "10")  // Coord0
	checkBalanceByIdx(t, tp.s, 258, "100") // B
	checkBalanceByIdx(t, tp.s, 259, "0")   // Coord0
}

func TestTwoExits(t *testing.T) {
	// In the first part we generate a batch with two force exits for the
	// same account of 20 each.  The txprocessor output should be a single
	// exitInfo with balance of 40.
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
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
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
		ChainID:  chainID,
	}
	tp := NewTxProcessor(sdb, config)
	ptOuts := []*ProcessTxOutput{}
	for _, block := range blocks {
		for _, batch := range block.Rollup.Batches {
			ptOut, err := tp.ProcessTxs(nil, batch.L1UserTxs, nil, nil)
			require.NoError(t, err)
			ptOuts = append(ptOuts, ptOut)
		}
	}

	assert.Equal(t, 1, len(ptOuts[3].ExitInfos))
	assert.Equal(t, big.NewInt(40), ptOuts[3].ExitInfos[0].Balance)
	acc, err := sdb.GetAccount(256)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(60), acc.Balance)

	// In the second part we start a fresh statedb and generate a batch
	// with one force exit for the same account as before.  The txprocessor
	// output should be a single exitInfo with balance of 40, and the exit
	// merkle tree proof should be equal to the previous one.

	dir2, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir2))

	sdb2, err := statedb.NewStateDB(statedb.Config{Path: dir2, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 32})
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

	tp = NewTxProcessor(sdb2, config)
	ptOuts2 := []*ProcessTxOutput{}
	for _, block := range blocks {
		for _, batch := range block.Rollup.Batches {
			ptOut, err := tp.ProcessTxs(nil, batch.L1UserTxs, nil, nil)
			require.NoError(t, err)
			ptOuts2 = append(ptOuts2, ptOut)
		}
	}

	assert.Equal(t, ptOuts[3].ExitInfos[0].MerkleProof, ptOuts2[3].ExitInfos[0].MerkleProof)
}
