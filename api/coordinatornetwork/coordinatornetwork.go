/*
Package coordinatornetwork implements a comunication layer among coordinators
in order to share information such as transactions in the pool and create account authorizations.

To do so the pubsub gossip protocol is used.
This code is currently eavily based on this example: https://github.com/libp2p/go-libp2p/blob/master/examples/pubsub
*/
package coordinatornetwork

import (
	"context"
	"fmt"
	"math/rand"
	mrand "math/rand"
	"os"
	"sync"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

const (
	discoveryInterval   = time.Hour                             // TODO: move to config file
	discoveryServiceTag = "coordnet/hermez-coordinator-network" // TODO: should include ChainID
	RendezvousString    = "coordnet/hermez-coordinator-network-meeting-point"
)

type CoordinatorNetworkAdvancedConfig struct {
	SetupCustomDiscovery func(context.Context, host.Host) error
	// for testing purpose
	port string
}

// CoordinatorNetwork it's a p2p communication layer that enables coordinators to exchange information
// in benefit of the network and them selfs. The main goal is to share L2 data (common.PoolL2Tx and common.AccountCreationAuth)
type CoordinatorNetwork struct {
	txsPool  pubSubTxsPool
	TxPoolCh chan *common.PoolL2Tx
}

// NewCoordinatorNetwork connects to coordinators network and return a CoordinatorNetwork
// to be able to receive and send information from and to other coordinators.
// For default config set config to nil
// TODO: port should be constant, but this makes testing easier
func NewCoordinatorNetwork(registeredCoords []common.Coordinator, config *CoordinatorNetworkAdvancedConfig) (CoordinatorNetwork, error) {
	ctx := context.Background()
	libp2pPort := common.CoordinatorsNetworkPort
	if config != nil && config.port != "" {
		libp2pPort = config.port
	}
	// Create a new libp2p Host
	cfgOpts := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/" + libp2pPort)

	// Create ID by generating private key
	// TODO: generate ID from coordinator's priv key
	rand.Seed(time.Now().UnixNano())
	r := mrand.New(mrand.NewSource(int64(rand.Int())))                    //nolint:gosec
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r) //nolint:gomnd
	if err != nil {
		return CoordinatorNetwork{}, err
	}

	// Create libp2p host instance
	h, err := libp2p.New(ctx, cfgOpts, libp2p.Identity(priv))
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	log.Debug("libp2p ID: ", h.ID().Pretty())
	log.Debug("libp2p Addr: ", h.Addrs())

	// Create pubsub instance
	var ps *pubsub.PubSub
	if config != nil && config.SetupCustomDiscovery != nil {
		// Custom discovery
		if err := config.SetupCustomDiscovery(ctx, h); err != nil {
			return CoordinatorNetwork{}, err
		}
		// Create a new PubSub service using the GossipSub router without discovery option
		ps, err = pubsub.NewGossipSub(ctx, h)
		if err != nil {
			return CoordinatorNetwork{}, err
		}
	} else {
		// setup default discovery
		discoveryOption, err := setupDiscovery(ctx, h, registeredCoords)
		if err != nil && err.Error() == "Unable to connect to any peer" {
			log.Error(err.Error(), ". This is expected if this is the first coordinator to join the network")
		} else if err != nil {
			return CoordinatorNetwork{}, err
		}
		// Create a new PubSub service using the GossipSub router with discovery option
		ps, err = pubsub.NewGossipSub(ctx, h, discoveryOption)
		if err != nil {
			return CoordinatorNetwork{}, err
		}
	}

	// Join transactions pool pubsub network
	txsPool, err := joinPubSubTxsPool(ctx, ps, h.ID())
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	log.Info("Joined to tx pool pubsub network")
	// TODO: add support for atomic txs and account creation auths

	return CoordinatorNetwork{
		txsPool:  txsPool,
		TxPoolCh: txsPool.Txs,
	}, nil
}

// PublishTx send a L2 transaction to the coordinators network
func (cn CoordinatorNetwork) PublishTx(tx common.PoolL2Tx) error {
	return cn.txsPool.publish(tx)
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// handlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("%s discovered new peer %s\n", n.h.ID().Pretty(), pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates a local KDHT that tries to add peers based on the already registered
// coordinators on the smart contract, then it will set a mechanism to discover more peers over time.
// Note: code based on: https://gist.github.com/popsUlfr/3cab45cc6203e10d11942f16f82c65c1
func setupDiscovery(ctx context.Context, h host.Host, registeredCoordinators []common.Coordinator) (pubsub.Option, error) {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	// dht.New
	dht, err := dht.New(ctx, h, dht.WanDHTOption())
	if err != nil {
		return nil, err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	// TODO: add more advanced configuration support
	log.Info("Bootstrapping the DHT to handle peer discovery")
	if err = dht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	connectedPeers := 0
	for i := range registeredCoordinators {
		log.Debug(i)
		// addr, err := coord.P2PAddr()
		// if err != nil {
		// 	log.Warn(err)
		// 	continue
		// }
		_addr := os.Getenv("ADDR")
		addr, err := multiaddr.NewMultiaddr(_addr)
		if err != nil {
			log.Warn(err)
			continue
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Warn(err)
			continue
		}
		log.Debug("Connecting to ", peerInfo.String())
		wg.Add(1)
		go func() {
			if err := h.Connect(ctx, *peerInfo); err != nil {
				log.Infof("Failed to connect to coordinator %s: %s", peerInfo.String(), err)
			} else {
				log.Infof("Connection established with coordinator: %s", peerInfo.String())
				connectedPeers++
			}
			wg.Done()
		}()
	}
	wg.Wait()
	disc := discovery.NewRoutingDiscovery(dht)
	discoveryOpt := pubsub.WithDiscovery(disc)
	if connectedPeers == 0 {
		log.Warn("Unable to connect to any peer")
	}

	// TODO: advertise myself and find more peers beyond network bootstrap
	// TODO: investigate about potential performance issues if too many peers, good default values for rescans, ...
	pesto := h.Peerstore()
	ids := pesto.Peers()
	go func() {
		for {
			newIds := pesto.Peers()
			if len(newIds) != len(ids) {
				log.Infof("Connected peers on coordinators network: %d", len(newIds))
				ids = newIds
			}
			time.Sleep(10 * time.Second)
			_, err := disc.Advertise(ctx, RendezvousString)
			if err != nil {
				log.Debug("Advertise failed retrying... ", err)
			}
			// see limit option
			peerAddrsChan, err := disc.FindPeers(ctx, RendezvousString)
			if err != nil {
				log.Error(err)
				return
			}
			select {
			case peerAddr := <-peerAddrsChan:
				// most the peers are empty!
				if (len(peerAddr.Addrs) == 0 && peerAddr.ID == "") || peerAddr.ID == h.ID() {
					continue
				}
				log.Debug("New peer found at the rendezvous point")
				err := h.Connect(ctx, peerAddr)
				if err != nil {
					log.Error("Routing connect error:", err, peerAddr)
				} else {
					log.Error("Routing connect:", peerAddr)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return discoveryOpt, nil
}
