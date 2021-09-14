package coordinatornetwork

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"time"

	"github.com/arnaubennassar/eth2libp2p"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-tcp-transport"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

func setupHost(ctx context.Context, ethPrivKey *ecdsa.PrivateKey, bootstrapPeers []multiaddr.Multiaddr) (host.Host, *dht.IpfsDHT, error) {
	// Set up the host identity options
	libp2pWallet, err := eth2libp2p.NewLibP2PIdentityFromEthPrivKey(ethPrivKey)
	if err != nil {
		return nil, nil, err
	}

	// Set up TCP transport and options
	transport := libp2p.Transport(tcp.NewTCPTransport)

	// Set up host listener address options
	muladdr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/" + common.CoordinatorsNetworkPort)
	if err != nil {
		return nil, nil, err
	}
	listen := libp2p.ListenAddrs(muladdr)

	// Set up the stream multiplexer and connection manager options
	// TODO: investigate and fine tune this. Got no clue on what it does, based on this example:
	// https://github.com/manishmeganathan/peerchat
	muxer := libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport)
	conn := libp2p.ConnectionManager(connmgr.NewConnManager(100, 400, time.Minute)) //nolint:gomnd

	// Setup NAT traversal and relay options
	nat := libp2p.NATPortMap()
	relay := libp2p.EnableAutoRelay()

	// Declare a KadDHT
	var coordnetDHT *dht.IpfsDHT
	// Setup a routing configuration with the KadDHT
	routing := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		coordnetDHT, err = setupCoordnetDHT(ctx, h, bootstrapPeers)
		return coordnetDHT, err
	})

	// Construct a new libP2P host with the created options
	opts := libp2p.ChainOptions(libp2pWallet.IdentityOption, listen, transport, muxer, conn, nat, routing, relay)
	libhost, err := libp2p.New(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	// Return the created host and the kademlia DHT
	return libhost, coordnetDHT, nil
}

func setupCoordnetDHT(ctx context.Context, self host.Host, bootstrapPeers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	// Create DHT server mode option
	dhtMode := dht.Mode(dht.ModeServer)
	bootstrapPeersInfo := []peer.AddrInfo{}
	for _, bootstrapPeer := range bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromP2pAddr(bootstrapPeer)
		if err != nil {
			log.Warn(err)
			continue
		}
		bootstrapPeersInfo = append(bootstrapPeersInfo, *peerInfo)
	}

	if len(bootstrapPeers) == 0 {
		log.Warn("Unable to set any bootstrap peer. Coordinator network will fail to stablish connection," +
			"unless someone connects to this coordinator directly (note that at some point there has to be a first peer in the network)")
	}

	// Start a Kademlia DHT on the host in server mode using coordinators as bootstrap peers
	coordnetDHT, err := dht.New(ctx, self, dhtMode, dht.BootstrapPeers(bootstrapPeersInfo...))
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
	cidvalue, err := generateCID(coordnet.discoveryServiceTag)
	if err != nil {
		return err
	}

	// Announce that this host can provide the service CID
	if err := coordnet.dht.Provide(coordnet.ctx, cidvalue, true); err != nil {
		return err
	}
	// Sleep to give time for the advertisement to propagate
	// TODO: fine tune delay value and/or enable users to configure it
	time.Sleep(time.Second * 5) //nolint:gomnd

	// Find the other providers for the service CID
	peerchan := coordnet.dht.FindProvidersAsync(coordnet.ctx, cidvalue, 0)
	// Connect to peers as they are discovered
	handlePeerDiscovery(coordnet.self, peerchan)
	return nil
}

func (coordnet CoordinatorNetwork) advertiseConnect() error {
	// Advertise the availability of the service on this node
	_, err := coordnet.discovery.Advertise(coordnet.ctx, coordnet.discoveryServiceTag)
	if err != nil {
		return err
	}
	// Sleep to give time for the advertisement to propagate
	// TODO: fine tune delay value and/or enable users to configure it
	time.Sleep(time.Second * 5) //nolint:gomnd

	// Find all peers advertising the same service
	peerchan, err := coordnet.discovery.FindPeers(coordnet.ctx, coordnet.discoveryServiceTag)
	// Handle any potential error
	if err != nil {
		return err
	}
	// Connect to peers as they are discovered
	handlePeerDiscovery(coordnet.self, peerchan)
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
	cidValue := cid.NewCidV1(12, mulhash) //nolint:gomnd
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
