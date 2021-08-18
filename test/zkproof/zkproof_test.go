package zkproof

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/stretchr/testify/require"
)

var proofServerURL string

const pollInterval = 200 * time.Millisecond

func TestMain(m *testing.M) {
	exitVal := 0
	proofServerURL = os.Getenv("PROOF_SERVER_URL")
	exitVal = m.Run()
	for _, dir := range deleteme {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
	os.Exit(exitVal)
}

const MaxTx = 352
const NLevels = 32
const MaxL1Tx = 256
const MaxFeeTx = 64
const ChainID uint16 = 1

var txprocConfig = txprocessor.Config{
	NLevels:  uint32(NLevels),
	MaxTx:    MaxTx,
	MaxL1Tx:  MaxL1Tx,
	MaxFeeTx: MaxFeeTx,
	ChainID:  ChainID,
}

func initStateDB(t *testing.T, typ statedb.TypeStateDB) *statedb.StateDB {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128, Type: typ, NLevels: NLevels})
	require.NoError(t, err)
	return sdb
}

func sendProofAndCheckResp(t *testing.T, zki *common.ZKInputs) {
	if proofServerURL == "" {
		log.Debug("No PROOF_SERVER_URL defined, not using ProofServer")
		return
	}

	log.Infof("sending proof to %s", proofServerURL)
	// Store zkinputs json for debugging purposes
	zkInputsJSON, err := json.Marshal(zki)
	require.NoError(t, err)
	err = ioutil.WriteFile("/tmp/dbgZKInputs.json", zkInputsJSON, 0640) //nolint:gosec
	require.NoError(t, err)

	proofServerClient := prover.NewProofServerClient(proofServerURL, pollInterval)
	err = proofServerClient.WaitReady(context.Background())
	require.NoError(t, err)
	err = proofServerClient.CalculateProof(context.Background(), zki)
	require.NoError(t, err)
	proof, pubInputs, err := proofServerClient.GetProof(context.Background())
	require.NoError(t, err)
	fmt.Printf("proof: %#v\n", proof)
	fmt.Printf("pubInputs: %#v\n", pubInputs)
}

func TestZKInputsEmpty(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)

	coordIdxs := []common.Idx{}
	l1UserTxs := []common.L1Tx{}
	l1CoordTxs := []common.L1Tx{}
	l2Txs := []common.PoolL2Tx{}
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs) // test empty batch ZKInputs

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs = txsets.GenerateTxsZKInputs0(t, ChainID)

	_, err = tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	coordIdxs = []common.Idx{}
	l1UserTxs = []common.L1Tx{}
	l1CoordTxs = []common.L1Tx{}
	l2Txs = []common.PoolL2Tx{}
	ptOut, err = tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)
	sendProofAndCheckResp(t, ptOut.ZKInputs) // test empty batch ZKInputs after a non-empty batch

	sdb.Close()
}

func TestZKInputs0(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs0(t, ChainID)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}
func TestZKInputs1(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs1(t, ChainID)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}
func TestZKInputs2(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs2(t, ChainID)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}
func TestZKInputs3(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs3(t, ChainID)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}
func TestZKInputs4(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs4(t, ChainID)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}

func TestZKInputs5(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs5(t, ChainID)

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}

func TestZKInputs6(t *testing.T) {
	sdb := initStateDB(t, statedb.TypeBatchBuilder)

	tc := til.NewContext(ChainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(txsets.SetBlockchainMinimumFlow0)
	require.NoError(t, err)

	// restart nonces of TilContext, as will be set by generating directly
	// the PoolL2Txs for each specific batch with tc.GeneratePoolL2Txs
	tc.RestartNonces()

	tp := txprocessor.NewTxProcessor(sdb, txprocConfig)
	// batch1
	ptOut, err := tp.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch2
	l1UserTxs := []common.L1Tx{}
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch3
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[2].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch4
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[3].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[3].L2Txs)
	ptOut, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[3].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch5
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[4].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[4].L2Txs)
	ptOut, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[4].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch6
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[5].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[5].L2Txs)
	ptOut, err = tp.ProcessTxs(nil, l1UserTxs, blocks[0].Rollup.Batches[5].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch7
	// simulate the PoolL2Txs of the batch7
	batchPoolL2 := `
	Type: PoolL2
	PoolTransferToEthAddr(1) A-B: 200 (126)
	PoolTransferToEthAddr(0) B-C: 100 (126)`
	poolL2Txs, err := tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{261, 262}
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[6].Batch.ForgeL1TxsNum])
	l2Txs = poolL2Txs
	ptOut, err = tp.ProcessTxs(coordIdxs, l1UserTxs,
		blocks[0].Rollup.Batches[6].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch8
	// simulate the PoolL2Txs of the batch8
	batchPoolL2 = `
	Type: PoolL2
	PoolTransfer(0) A-B: 100 (126)
	PoolTransfer(0) C-A: 50 (126)
	PoolTransfer(1) B-C: 100 (126)
	PoolExit(0) A: 100 (126)`
	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)

	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[7].Batch.ForgeL1TxsNum])
	l2Txs = poolL2Txs
	ptOut, err = tp.ProcessTxs(coordIdxs, l1UserTxs,
		blocks[0].Rollup.Batches[7].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch9
	// simulate the PoolL2Txs of the batch9
	batchPoolL2 = `
	Type: PoolL2
	PoolTransfer(0) D-A: 300 (126)
	PoolTransfer(0) B-D: 100 (126)`
	poolL2Txs, err = tc.GeneratePoolL2Txs(batchPoolL2)
	require.NoError(t, err)

	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[0].Batch.ForgeL1TxsNum])
	l2Txs = poolL2Txs
	coordIdxs = []common.Idx{262}
	ptOut, err = tp.ProcessTxs(coordIdxs, l1UserTxs,
		blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	// batch10
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l2Txs = []common.PoolL2Tx{}
	coordIdxs = []common.Idx{}
	ptOut, err = tp.ProcessTxs(coordIdxs, l1UserTxs,
		blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.NoError(t, err)

	sendProofAndCheckResp(t, ptOut.ZKInputs)

	sdb.Close()
}
