package coordinatornetwork

import (
	"context"
	"crypto/sha256"
	"math/rand"
	"os"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/mr-tron/base58"

	// dht "github.com/libp2p/go-libp2p-kad-dht/dual"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

func setupHost(ctx context.Context, registeredCoordinators []common.Coordinator) (host.Host, *dht.IpfsDHT, error) {
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
	muladdr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/" + common.CoordinatorsNetworkPort)
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
	var coordnetDHT *dht.IpfsDHT
	// Setup a routing configuration with the KadDHT
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		coordnetDHT, err = setupCoordnetDHT(ctx, h, registeredCoordinators)
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

func setupCoordnetDHT(ctx context.Context, self host.Host, registeredCoordinators []common.Coordinator) (*dht.IpfsDHT, error) {
	// Create DHT server mode option
	dhtMode := dht.Mode(dht.ModeServer)

	bootstrapPeers := []peer.AddrInfo{}
	for i := 0; i < len(registeredCoordinators); i++ {
		coordAddr, err := registeredCoordinators[i].P2PAddr()
		if err != nil {
			log.Warn(err)
			continue
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(coordAddr)
		if err != nil {
			log.Warn(err)
			continue
		} else {
			bootstrapPeers = append(bootstrapPeers, *peerInfo)
		}
	}
	// Create the DHT bootstrap peers option
	// TODO: using this for testing as part of the PoC, find a better way for testing
	_addr := os.Getenv("ADDR")
	if _addr != "" {
		addr, err := multiaddr.NewMultiaddr(_addr)
		if err != nil {
			return nil, err
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Warn(err)
		} else {
			bootstrapPeers = append(bootstrapPeers, *peerInfo)
		}
	}

	if len(bootstrapPeers) == 0 {
		log.Error("Unable to set any bootstrap peer. Coordinator network will fail to stablish connection," +
			"unless someone connects to this coordinator directly (note that at some point there has to be a first peer in the network)")
	}

	// Start a Kademlia DHT on the host in server mode using coordinators as bootstrap peers
	coordnetDHT, err := dht.New(ctx, self, dhtMode, dht.BootstrapPeers(bootstrapPeers...))
	if err != nil {
		return nil, err
	}
	if err := coordnetDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}
	return coordnetDHT, nil
}

func setupPubSub(ctx context.Context, nodehost host.Host, routingdiscovery *discovery.RoutingDiscovery) (*pubsub.PubSub, error) {
	// Create a new PubSub service which uses a GossipSub router
	return pubsub.NewGossipSub(ctx, nodehost, pubsub.WithDiscovery(routingdiscovery))
}

func (coordnet CoordinatorNetwork) announceConnect() error {
	// Generate the Service CID
	cidvalue, err := generateCID(discoveryServiceTag)
	if err != nil {
		return err
	}

	// Announce that this host can provide the service CID
	if err := coordnet.dht.Provide(coordnet.ctx, cidvalue, true); err != nil {
		return err
	}
	// Sleep to give time for the advertisement to propagate
	time.Sleep(time.Second * 5)

	// Find the other providers for the service CID
	peerchan := coordnet.dht.FindProvidersAsync(coordnet.ctx, cidvalue, 0)
	// Connect to peers as they are discovered
	go handlePeerDiscovery(coordnet.self, peerchan)
	return nil
}

func (coordnet CoordinatorNetwork) advertiseConnect() error {
	// Advertise the availabilty of the service on this node
	_, err := coordnet.discovery.Advertise(coordnet.ctx, discoveryServiceTag)
	if err != nil {
		return err
	}
	// Sleep to give time for the advertisement to propagate
	time.Sleep(time.Second * 5)

	// Find all peers advertising the same service
	peerchan, err := coordnet.discovery.FindPeers(coordnet.ctx, discoveryServiceTag)
	// Handle any potential error
	if err != nil {
		return err
	}
	// Connect to peers as they are discovered
	go handlePeerDiscovery(coordnet.self, peerchan)
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
		// Connect to the peer
		if err := self.Connect(context.Background(), peer); err != nil {
			log.Debug(err)
		}
	}
}
