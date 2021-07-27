package coordinatornetwork

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p-core/host"
	localDiscovery "github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSubTxsPoolLocal(t *testing.T) {
	net1, err := NewCoordinatorNetwork([]common.Coordinator{}, &CoordinatorNetworkAdvancedConfig{
		port:                 "1234",
		SetupCustomDiscovery: setupDiscoveryLocal,
	})
	require.NoError(t, err)
	net2, err := NewCoordinatorNetwork([]common.Coordinator{}, &CoordinatorNetworkAdvancedConfig{
		port:                 "4321",
		SetupCustomDiscovery: setupDiscoveryLocal,
	})
	require.NoError(t, err)

	txToSend := common.PoolL2Tx{
		FromIdx:     2344,
		ToIdx:       4324,
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

func TestPubSubFakeServer(t *testing.T) {
	// Fake server
	if os.Getenv("FAKE_COORDNET") != "yes" {
		return
	}
	peerList := os.Getenv("PEER_LIST")
	if peerList == "" {
		panic("Expecting ENV PEER_LIST, containing a coma separated list of URLs")
	}
	peers := strings.Split(peerList, ",")
	registeredCoordinators := []common.Coordinator{}
	for i := 0; i < len(peers); i++ {
		log.Info(peers[i])
		registeredCoordinators = append(registeredCoordinators, common.Coordinator{URL: peers[i]})
	}

	coordnet, err := NewCoordinatorNetwork(registeredCoordinators, nil)
	require.NoError(t, err)

	// Receive or send
	if os.Getenv("PUBLISH") == "yes" {
		txToPublish, err := common.NewPoolL2Tx(&common.PoolL2Tx{
			FromIdx:     666,
			ToIdx:       555,
			Amount:      big.NewInt(555555),
			TokenID:     1,
			TokenSymbol: "HEZ",
		})
		require.NoError(t, err)
		time.Sleep(30 * time.Second)
		require.NoError(t, coordnet.PublishTx(*txToPublish))
		log.Infof("Tx %s published to the network", txToPublish.TxID.String())
		return
	}
	log.Warn("Entering endless loop, until ^C is received")
	receivedTx := <-coordnet.TxPoolCh
	log.Info("Tx received: ", receivedTx.TxID)
}

// setupDiscoveryLocal is used for local testing purposes
// it creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them
func setupDiscoveryLocal(ctx context.Context, h host.Host) error {
	// setup mDNS discovery to find local peers
	disc, err := localDiscovery.NewMdnsService(ctx, h, discoveryInterval, discoveryServiceTag)
	if err != nil {
		return err
	}

	n := discoveryNotifee{h: h}
	disc.RegisterNotifee(&n)
	return nil
}
