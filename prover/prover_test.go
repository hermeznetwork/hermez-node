package prover

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var apiURL = "http://localhost:3000/api"

const pollInterval = 1 * time.Second

var proofServerClient *ProofServerClient

func TestMain(m *testing.M) {
	exitVal := 0
	if os.Getenv("INTEGRATION") != "" {
		_apiURL := os.Getenv("PROOF_SERVER_URL")
		if _apiURL != "" {
			apiURL = _apiURL
		}
		proofServerClient = NewProofServerClient(apiURL, pollInterval)
		err := proofServerClient.WaitReady(context.Background())
		if err != nil {
			panic(err)
		}
		exitVal = m.Run()
	}
	os.Exit(exitVal)
}

func TestApiServer(t *testing.T) {
	t.Run("testAPIStatus", testAPIStatus)
	t.Run("testCalculateProof", testCalculateProof)
	time.Sleep(time.Second / 4)
	err := proofServerClient.WaitReady(context.Background())
	require.NoError(t, err)
	t.Run("testGetProof", testGetProof)
	t.Run("testCancel", testCancel)
}

func testAPIStatus(t *testing.T) {
	status, err := proofServerClient.apiStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, true, status.Status.IsReady())
}

func testCalculateProof(t *testing.T) {
	zkInputs := common.NewZKInputs(0, 100, 24, 512, 32, big.NewInt(1))
	err := proofServerClient.CalculateProof(context.Background(), zkInputs)
	require.NoError(t, err)
}

func testGetProof(t *testing.T) {
	proof, pubInputs, err := proofServerClient.GetProof(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, proof.PiA)
	assert.NotEqual(t, [2]*big.Int{}, proof.PiA)
	assert.NotNil(t, proof.PiB)
	assert.NotEqual(t, [3][2]*big.Int{}, proof.PiB)
	assert.NotNil(t, proof.PiC)
	assert.NotEqual(t, [2]*big.Int{}, proof.PiC)
	assert.NotNil(t, proof.Protocol)
	assert.NotEqual(t, 0, len(pubInputs))
}

func testCancel(t *testing.T) {
	zkInputs := common.NewZKInputs(0, 100, 24, 512, 32, big.NewInt(1))
	err := proofServerClient.CalculateProof(context.Background(), zkInputs)
	require.NoError(t, err)
	time.Sleep(time.Second / 4)
	err = proofServerClient.Cancel(context.Background())
	require.NoError(t, err)
	status, err := proofServerClient.apiStatus(context.Background())
	require.NoError(t, err)
	for status.Status == StatusCodeBusy {
		time.Sleep(proofServerClient.pollInterval)
		status, err = proofServerClient.apiStatus(context.Background())
		require.NoError(t, err)
	}
	assert.Equal(t, StatusCodeAborted, status.Status)
}
