/*
Package coordinatornetwork implements a comunication layer among coordinators
in order to share information such as transactions in the pool and create account authorizations.

To do so the pubsub gossip protocol is used.
This code is currently eavily based on this example: https://github.com/libp2p/go-libp2p/blob/master/examples/pubsub
*/
package coordinatornetwork

import (
	"context"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p-core/host"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	// dht "github.com/libp2p/go-libp2p-kad-dht/dual"
)

const (
	discoveryServiceTag = "coordnet/hermez-coordinator-network" // TODO: should include ChainID
	RendezvousString    = "coordnet/hermez-coordinator-network-meeting-point"
)

// CoordinatorNetwork it's a p2p communication layer that enables coordinators to exchange information
// in benefit of the network and them selfs. The main goal is to share L2 data (common.PoolL2Tx and common.AccountCreationAuth)
type CoordinatorNetwork struct {
	self host.Host
	dht  *dht.IpfsDHT
	// dht       *dht.DHT
	ctx       context.Context
	discovery *discovery.RoutingDiscovery
	txsPool   pubSubTxsPool
	TxPoolCh  chan *common.PoolL2Tx
}

// NewCoordinatorNetwork connects to coordinators network and return a CoordinatorNetwork
// to be able to receive and send information from and to other coordinators.
// For default config set config to nil
// TODO: port should be constant, but this makes testing easier
func NewCoordinatorNetwork(registeredCoords []common.Coordinator) (CoordinatorNetwork, error) {
	// Setup a background context
	ctx := context.Background()

	// Setup a P2P Host Node
	self, coordnetDHT, err := setupHost(ctx, registeredCoords)
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	log.Debug("libp2p ID: ", self.ID().Pretty())

	// Create a peer discovery service using the Kad DHT
	routingDiscovery := discovery.NewRoutingDiscovery(coordnetDHT)
	// Debug log
	log.Debug("Created the Peer Discovery Service.")

	// Create a PubSub handler with the routing discovery
	pubsubHandler, err := setupPubSub(ctx, self, routingDiscovery)
	if err != nil {
		return CoordinatorNetwork{}, err
	}

	// Join transactions pool pubsub network
	txsPool, err := joinPubSubTxsPool(ctx, pubsubHandler, self.ID())
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	log.Info("Joined to tx pool pubsub network")
	// TODO: add support for atomic txs and account creation auths

	return CoordinatorNetwork{
		self:      self,
		dht:       coordnetDHT,
		ctx:       ctx,
		discovery: routingDiscovery,
		txsPool:   txsPool,
		TxPoolCh:  txsPool.Txs,
	}, nil
}

// PublishTx send a L2 transaction to the coordinators network
func (coordnet CoordinatorNetwork) PublishTx(tx common.PoolL2Tx) error {
	return coordnet.txsPool.publish(tx)
}

func (coordnet CoordinatorNetwork) FindMorePeers() error {
	if err := coordnet.advertiseConnect(); err != nil {
		return err
	}
	if err := coordnet.announceConnect(); err != nil {
		return err
	}
	return nil
}
