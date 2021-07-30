package coordinatornetwork

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/require"
)

func TestPubSubFakeServer(t *testing.T) {
	// NOTE THAT THIS IS INTENDED TO RUN MANUALLY USING DIFFERENT MACHINES
	// Fake server
	if os.Getenv("FAKE_COORDNET") != "yes" {
		return
	}
	coordnet, err := NewCoordinatorNetwork([]common.Coordinator{})
	require.NoError(t, err)

	// find other peers
	go func() {
		for {
			if err := coordnet.FindMorePeers(); err != nil {
				log.Warn(err)
			}
			time.Sleep(10 * time.Second)
			log.Infof("%d peers connected to the pubsub", len(coordnet.txsPool.topic.ListPeers()))
		}
	}()

	// Receive or send
	if os.Getenv("PUBLISH") == "yes" {
		// Wait until some peers have been found
		peers := []peer.ID{}
		for len(peers) == 0 {
			peers = coordnet.txsPool.topic.ListPeers()
		}
		log.Info("peers on the pubsub: ")
		for _, peer := range peers {
			log.Info(peer.Pretty())
		}
		time.Sleep(60 * time.Second)
		// Send tx
		txToPublish, err := common.NewPoolL2Tx(&common.PoolL2Tx{
			FromIdx:     666,
			ToIdx:       555,
			Amount:      big.NewInt(555555),
			TokenID:     1,
			TokenSymbol: "HEZ",
		})
		require.NoError(t, err)
		require.NoError(t, coordnet.PublishTx(*txToPublish))
		log.Infof("Tx %s published to the network", txToPublish.TxID.String())
		time.Sleep(10 * time.Second)
		return
	}
	log.Warn("Entering endless loop, until a tx is received or ^C is received")
	receivedTx := <-coordnet.TxPoolCh
	log.Info("Tx received: ", receivedTx.TxID)
}
