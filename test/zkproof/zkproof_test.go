package zkproof

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/iden3/go-iden3-crypto/babyjub"
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

func generateJsUsers(t *testing.T) []til.User {
	// same values than in the js test
	// skJsHex is equivalent to the 0000...000i js private key in commonjs
	skJsHex := []string{"7eb258e61862aae75c6c1d1f7efae5006ffc9e4d5596a6ff95f3df4ea209ea7f", "c005700f76f4b4cec710805c21595688648524df0a9d467afae537b7a7118819", "b373d14c67fb2a517bf4ac831c93341eec8e1b38dbc14e7d725b292a7cf84707", "2064b68d04a7aaae0ac3b36bf6f1850b380f1423be94a506c531940bd4a48b76"}
	addrHex := []string{"0x7e5f4552091a69125d5dfcb7b8c2659029395bdf", "0x2b5ad5c4795c026514f8317c7a215e218dccd6cf", "0x6813eb9362372eef6200f3b1dbc3f819671cba69", "0x1eff47bc3a10a45d4b230b5d10e37751fe6aa718"}
	var users []til.User
	for i := 0; i < len(skJsHex); i++ {
		skJs, err := hex.DecodeString(skJsHex[i])
		require.NoError(t, err)
		var sk babyjub.PrivateKey
		copy(sk[:], skJs)
		// bjj := sk.Public()
		user := til.User{
			Name: strconv.Itoa(i),
			BJJ:  &sk,
			Addr: ethCommon.HexToAddress(addrHex[i]),
		}
		users = append(users, user)
	}
	assert.Equal(t, "d746824f7d0ac5044a573f51b278acb56d823bec39551d1d7bf7378b68a1b021", users[0].BJJ.Public().String())
	assert.Equal(t, "4d05c307400c65795f02db96b1b81c60386fd53e947d9d3f749f3d99b1853909", users[1].BJJ.Public().String())
	assert.Equal(t, "38ffa002724562eb2a952a2503e206248962406cf16392ff32759b6f2a41fe11", users[2].BJJ.Public().String())
	assert.Equal(t, "c719e6401190be7fa7fbfcd3448fe2755233c01575341a3b09edadf5454f760b", users[3].BJJ.Public().String())

	return users
}

func signL2Tx(t *testing.T, chainID uint16, user til.User, l2Tx common.PoolL2Tx) common.PoolL2Tx {
	toSign, err := l2Tx.HashToSign(chainID)
	require.NoError(t, err)
	sig := user.BJJ.SignPoseidon(toSign)
	l2Tx.Signature = sig.Compress()
	return l2Tx
}

const MaxTx = 376
const NLevels = 32
const MaxL1Tx = 256
const MaxFeeTx = 64
const ChainID uint16 = 1

func TestTxProcessor(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := statedb.NewStateDB(dir, 128, statedb.TypeBatchBuilder, NLevels)
	require.NoError(t, err)

	// same values than in the js test
	users := generateJsUsers(t)

	l1Txs := []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
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
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       257,
			DepositAmount: big.NewInt(0),
			Amount:        big.NewInt(1000),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         1,
			Type:          common.TxTypeForceExit,
			UserOrigin:    true,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   257,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
		{
			FromIdx: 256,
			ToIdx:   1,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   1,
			Fee:     126,
			Type:    common.TxTypeExit,
		},
	}

	l2Txs[0] = signL2Tx(t, ChainID, users[0], l2Txs[0])
	l2Txs[1] = signL2Tx(t, ChainID, users[0], l2Txs[1])

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

	coordIdxs := []common.Idx{257}
	ptOut, err := tp.ProcessTxs(coordIdxs, l1Txs, nil, l2Txs)
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
