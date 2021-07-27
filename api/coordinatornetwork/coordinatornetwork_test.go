package coordinatornetwork

import (
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/stretchr/testify/require"
)

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

	coordnet, err := NewCoordinatorNetwork(registeredCoordinators)
	require.NoError(t, err)

	// find other peers
	go func() {
		for {
			if err := coordnet.FindMorePeers(); err != nil {
				log.Warn(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()

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
		log.Info("peers on the pubsub: ")
		peers := coordnet.txsPool.topic.ListPeers()
		for _, v := range peers {
			log.Info(v.Pretty())
		}
		require.NoError(t, coordnet.PublishTx(*txToPublish))
		log.Infof("Tx %s published to the network", txToPublish.TxID.String())
		return
	}
	log.Warn("Entering endless loop, until ^C is received")
	receivedTx := <-coordnet.TxPoolCh
	log.Info("Tx received: ", receivedTx.TxID)
}
