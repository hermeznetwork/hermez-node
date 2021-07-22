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
	mrand "math/rand"
	"strconv"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

const (
	// port                = "3298"
	discoveryInterval   = time.Hour // TODO: move to config file
	discoveryServiceTag = "hermez-coordinator-network"
)

// CoordinatorNetwork it's a p2p communication layer that enables coordinators to exchange information
// in benefit of the network and them selfs. The main goal is to share L2 data (common.PoolL2Tx and common.AccountCreationAuth)
type CoordinatorNetwork struct {
	txsPool  pubSubTxsPool
	TxPoolCh chan *common.PoolL2Tx
}

// NewCoordinatorNetwork connects to coordinators network and return a CoordinatorNetwork
// to be able to receive and send information from and to other coordinators
// TODO: port should be constant, but this makes testing easier
func NewCoordinatorNetwork(port string) (CoordinatorNetwork, error) {
	ctx := context.Background()
	// create a new libp2p Host that listens on a random TCP port
	cfgOpts := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/" + port)
	randNum, err := strconv.Atoi(port)
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	// TODO: generate ID from coordinator's priv key
	r := mrand.New(mrand.NewSource(int64(randNum)))                       //nolint:gosec
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r) //nolint:gomnd
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	h, err := libp2p.New(ctx, cfgOpts, libp2p.Identity(priv))
	if err != nil {
		return CoordinatorNetwork{}, err
	}
	log.Debug("whoami: ", h.ID().Pretty())
	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return CoordinatorNetwork{}, err
	}

	// setup discovery
	err = setupDiscovery(ctx, h)
	if err != nil {
		return CoordinatorNetwork{}, err
	}

	// Join transactions pool pubsub
	txsPool, err := joinPubSubTxsPool(ctx, ps, h.ID())
	if err != nil {
		return CoordinatorNetwork{}, err
	}

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

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
// TODO: make discovery take advantage of the known coordinator URLs
func setupDiscovery(ctx context.Context, h host.Host) error {
	// setup mDNS discovery to find local peers
	disc, err := discovery.NewMdnsService(ctx, h, discoveryInterval, discoveryServiceTag)
	if err != nil {
		return err
	}

	n := discoveryNotifee{h: h}
	disc.RegisterNotifee(&n)
	return nil
}
