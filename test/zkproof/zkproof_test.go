package zkproof

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var proofServerURL string

const pollInterval = 200 * time.Millisecond

func TestMain(m *testing.M) {
	exitVal := 0
	proofServerURL = os.Getenv("PROOF_SERVER_URL")
	if proofServerURL != "" {
		exitVal = m.Run()
	}
	os.Exit(exitVal)
}

const MaxTx = 376
const NLevels = 32
const MaxL1Tx = 256
const MaxFeeTx = 64
const ChainID uint16 = 1

func TestZKInputs5(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(dir, 128, statedb.TypeBatchBuilder, NLevels)
	require.NoError(t, err)

	_, coordIdxs, l1UserTxs, l1CoordTxs, l2Txs := txsets.GenerateTxsZKInputs5(t, ChainID)

	config := txprocessor.Config{
		NLevels:  uint32(NLevels),
		MaxTx:    MaxTx,
		MaxL1Tx:  MaxL1Tx,
		MaxFeeTx: MaxFeeTx,
		ChainID:  ChainID,
	}
	tp := txprocessor.NewTxProcessor(sdb, config)

	// skip first batch to do the test with BatchNum=1
	_, err = tp.ProcessTxs(nil, nil, nil, nil)
	require.NoError(t, err)

	ptOut, err := tp.ProcessTxs(coordIdxs, l1UserTxs, l1CoordTxs, l2Txs)
	require.NoError(t, err)

	// Store zkinputs json for debugging purposes
	zkInputsJSON, err := json.Marshal(ptOut.ZKInputs)
	require.NoError(t, err)
	err = ioutil.WriteFile("/tmp/dbgZKInputs.json", zkInputsJSON, 0640) //nolint:gosec
	require.NoError(t, err)

	proofServerClient := prover.NewProofServerClient(proofServerURL, pollInterval)
	err = proofServerClient.WaitReady(context.Background())
	require.NoError(t, err)
	err = proofServerClient.CalculateProof(context.Background(), ptOut.ZKInputs)
	require.NoError(t, err)
	proof, pubInputs, err := proofServerClient.GetProof(context.Background())
	require.NoError(t, err)
	fmt.Printf("proof: %#v\n", proof)
	fmt.Printf("pubInputs: %#v\n", pubInputs)
}
