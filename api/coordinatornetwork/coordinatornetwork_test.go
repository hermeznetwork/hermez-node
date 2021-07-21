package coordinatornetwork

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSubTxsPool(t *testing.T) {
	net1, err := NewCoordinatorNetwork("4321")
	require.NoError(t, err)
	net2, err := NewCoordinatorNetwork("1234")
	require.NoError(t, err)

	txToSend := common.PoolL2Tx{
		FromIdx:     234,
		ToIdx:       432,
		TokenID:     4,
		TokenSymbol: "FOO",
		Amount:      big.NewInt(7),
	}
	// TODO: better way to way until libp2p is ready
	time.Sleep(10 * time.Second)
	require.NoError(t, net2.PublishTx(txToSend))
	receivedTx := <-net1.TxPoolCh
	// TODO: Cleaner test, this marshaling/unmarshaling it's ugly
	expectedTxBytes, err := json.Marshal(txToSend)
	require.NoError(t, err)
	expectedTx := common.PoolL2Tx{}
	require.NoError(t, json.Unmarshal(expectedTxBytes, &expectedTx))
	assert.Equal(t, expectedTx, *receivedTx)
}
