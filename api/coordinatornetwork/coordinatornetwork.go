/*
Package coordinatornetwork implements a comunication layer among coordinators
in order to share information such as transactions in the pool and create account authorizations.

To do so the pubsub gossip protocol is used.
This code is currently eavily based on this example: https://github.com/libp2p/go-libp2p/blob/master/examples/pubsub
*/
package coordinatornetwork

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	// dht "github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multihash"
	"github.com/sirupsen/logrus"
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
	self, coordnetDHT, err := setupHost(ctx)
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	log.Debug("libp2p ID: ", self.ID().Pretty())
	// Debug log
	log.Debug("Created the P2P Host and the Kademlia DHT.")

	// Bootstrap the Kad DHT
	if err := bootstrapDHT(ctx, self, coordnetDHT); err != nil {
		return CoordinatorNetwork{}, err
	}
	// Debug log
	log.Debug("Bootstrapped the Kademlia DHT and Connected to Bootstrap Peers")

	// Create a peer discovery service using the Kad DHT
	routingDiscovery := discovery.NewRoutingDiscovery(coordnetDHT)
	// Debug log
	log.Debug("Created the Peer Discovery Service.")

	// Create a PubSub handler with the routing discovery
	pubsubHandler := setupPubSub(ctx, self, routingDiscovery)

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

func (coordnet CoordinatorNetwork) announceConnect() error {
	// Generate the Service CID
	cidvalue, err := generateCID(discoveryServiceTag)
	if err != nil {
		return err
	}
	logrus.Debug("Generated the Service CID.")

	// Announce that this host can provide the service CID
	if err := coordnet.dht.Provide(coordnet.ctx, cidvalue, true); err != nil {
		return err
	}
	// Debug log
	log.Debug("Announced the PeerChat Service.")
	// Sleep to give time for the advertisment to propogate
	time.Sleep(time.Second * 5)

	// Find the other providers for the service CID
	peerchan := coordnet.dht.FindProvidersAsync(coordnet.ctx, cidvalue, 0)
	// Trace log
	log.Debug("Discovered PeerChat Service Peers.")

	// Connect to peers as they are discovered
	go handlePeerDiscovery(coordnet.self, peerchan)
	// Debug log
	log.Debug("Started Peer Connection Handler.")
	return nil
}

func (coordnet CoordinatorNetwork) advertiseConnect() error {
	// Advertise the availabilty of the service on this node
	ttl, err := coordnet.discovery.Advertise(coordnet.ctx, discoveryServiceTag)
	if err != nil {
		return err
	}
	// Debug log
	logrus.Debugln("Advertised the PeerChat Service.")
	// Sleep to give time for the advertisment to propogate
	time.Sleep(time.Second * 5)
	// Debug log
	log.Debugf("Service Time-to-Live is %s", ttl)

	// Find all peers advertising the same service
	peerchan, err := coordnet.discovery.FindPeers(coordnet.ctx, discoveryServiceTag)
	// Handle any potential error
	if err != nil {
		return err
	}
	// Trace log
	logrus.Debug("Discovered PeerChat Service Peers.")

	// Connect to peers as they are discovered
	go handlePeerDiscovery(coordnet.self, peerchan)
	// Trace log
	logrus.Info("Started Peer Connection Handler.")
	return nil
}

func generateCID(namestring string) (cid.Cid, error) {
	// Hash the service content ID with SHA256
	hash := sha256.Sum256([]byte(namestring))
	// Append the hash with the hashing codec ID for SHA2-256 (0x12),
	// the digest size (0x20) and the hash of the service content ID
	finalhash := append([]byte{0x12, 0x20}, hash[:]...)
	// Encode the fullhash to Base58
	b58string := base58.Encode(finalhash)

	// Generate a Multihash from the base58 string
	mulhash, err := multihash.FromB58String(string(b58string))
	if err != nil {
		return cid.Cid{}, err
	}

	// Generate a CID from the Multihash
	cidValue := cid.NewCidV1(12, mulhash)
	// Return the CID
	return cidValue, nil
}

func handlePeerDiscovery(self host.Host, peerchan <-chan peer.AddrInfo) {
	// Iterate over the peer channel
	for peer := range peerchan {
		// Ignore if the discovered peer is the host itself
		if peer.ID == self.ID() {
			continue
		}
		log.Debug("New peer found")
		// Connect to the peer
		if err := self.Connect(context.Background(), peer); err != nil {
			log.Warn("Error connecting to discovered peer: ", err)
		} else {
			log.Info("Connected to new peer. ", peer.ID)
		}
	}
}
