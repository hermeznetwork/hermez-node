package coordinatornetwork

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

func setupHost(ctx context.Context, config *CoordinatorNetworkAdvancedConfig) (host.Host, *dht.DHT, error) {
	// Set up the host identity options
	// Create ID by generating private key
	// TODO: generate ID from coordinator's priv key
	rand.Seed(time.Now().UnixNano())
	r := rand.New(rand.NewSource(int64(rand.Int())))                        //nolint:gosec
	prvkey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r) //nolint:gomnd
	if err != nil {
		return nil, nil, err
	}
	identity := libp2p.Identity(prvkey)

	// Trace log
	log.Info("Generated P2P Identity Configuration")

	// Set up TLS secured TCP transport and options
	transport := libp2p.Transport(tcp.NewTCPTransport)

	// Trace log
	log.Info("Generated P2P Security and Transport Configurations.")

	// Set up host listener address options
	libp2pPort := common.CoordinatorsNetworkPort
	if config != nil && config.port != "" {
		libp2pPort = config.port
	}
	muladdr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/" + libp2pPort)
	if err != nil {
		return nil, nil, err
	}
	listen := libp2p.ListenAddrs(muladdr)

	// Trace log
	log.Info("Generated P2P Address Listener Configuration.")

	// Set up the stream multiplexer and connection manager options
	muxer := libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport)
	conn := libp2p.ConnectionManager(connmgr.NewConnManager(100, 400, time.Minute))

	// Trace log
	log.Info("Generated P2P Stream Multiplexer, Connection Manager Configurations.")

	// Setup NAT traversal and relay options
	nat := libp2p.NATPortMap()
	relay := libp2p.EnableAutoRelay()

	// Trace log
	log.Info("Generated P2P NAT Traversal and Relay Configurations.")

	// Declare a KadDHT
	var coordnetDHT *dht.DHT
	// Setup a routing configuration with the KadDHT
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		coordnetDHT, err = setupCoordnetDHT(ctx, h)
		return coordnetDHT, err
	})

	// Trace log
	log.Info("Generated P2P Routing Configurations.")

	opts := libp2p.ChainOptions(identity, listen, transport, muxer, conn, nat, routing, relay)

	// Construct a new libP2P host with the created options
	libhost, err := libp2p.New(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	// Return the created host and the kademlia DHT
	return libhost, coordnetDHT, nil
}

func setupCoordnetDHT(ctx context.Context, self host.Host) (*dht.DHT, error) {
	// Create DHT server mode option
	// dhtMode := dht.LanDHTOption()
	dhtMode := dht.WanDHTOption()

	// Start a Kademlia DHT on the host in server mode
	coordnetDHT, err := dht.New(ctx, self, dhtMode)
	// Handle any potential error
	if err != nil {
		return nil, err
	}

	// Return the KadDHT
	return coordnetDHT, nil
}

// A function that bootstraps a given Kademlia DHT to satisfy the IPFS router
// interface and connects to all the bootstrap peers provided by libp2p
func bootstrapDHT(ctx context.Context, self host.Host, coordnetDHT *dht.DHT) error {
	// Bootstrap the DHT to satisfy the IPFS Router interface
	if err := coordnetDHT.Bootstrap(ctx); err != nil {
		return err
	}

	// Trace log
	logrus.Info("Set the Kademlia DHT into Bootstrap Mode.")

	// TODO: this should be built using coordinator info
	// Connect to other coordinators to bootstrap the network
	_addr := os.Getenv("ADDR")
	addr, err := multiaddr.NewMultiaddr(_addr)
	if err != nil {
		return err
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Warn(err)
	} else {
		if err := self.Connect(ctx, *peerInfo); err != nil {
			log.Infof("Failed to connect to coordinator %s: %s", peerInfo.String(), err)
		} else {
			log.Infof("Connection established with coordinator: %s", peerInfo.String())
		}
	}
	log.Info("Generated DHT Configuration.")

	// Log the number of bootstrap peers connected
	return nil
}

func setupPubSub(ctx context.Context, nodehost host.Host, routingdiscovery *discovery.RoutingDiscovery) *pubsub.PubSub {
	// Create a new PubSub service which uses a GossipSub router
	pubsubHandler, err := pubsub.NewGossipSub(ctx, nodehost, pubsub.WithDiscovery(routingdiscovery))
	// Handle any potential error
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"type":  "GossipSub",
		}).Fatalln("PubSub Handler Creation Failed!")
	}

	// Return the PubSub handler
	return pubsubHandler
}
