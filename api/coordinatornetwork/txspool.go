package coordinatornetwork

import (
	"context"
	"encoding/json"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	txsPoolTopicName = "hermez-coordinator-network-txs-pool"
)

// pubSubTxsPool represents a subscription to a single PubSub topic. Messages
// can be published to the topic with pubSubTxsPool.Publish, and received
// messages are pushed to the Messages channel.
type pubSubTxsPool struct {
	ctx     context.Context
	ps      *pubsub.PubSub
	topic   *pubsub.Topic
	sub     *pubsub.Subscription
	self    peer.ID
	handler func(common.PoolL2Tx) error
}

// joinPubSubTxsPool tries to subscribe to the PubSub topic for the room name, returning
// a pubSubTxsPool on success.
func joinPubSubTxsPool(
	ctx context.Context,
	ps *pubsub.PubSub,
	selfID peer.ID,
	handler func(common.PoolL2Tx) error,
) (pubSubTxsPool, error) {
	// join the pubsub topic
	topic, err := ps.Join(txsPoolTopicName)
	if err != nil {
		return pubSubTxsPool{}, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return pubSubTxsPool{}, err
	}

	psTxsPool := pubSubTxsPool{
		ctx:     ctx,
		ps:      ps,
		topic:   topic,
		sub:     sub,
		self:    selfID,
		handler: handler,
	}

	// start reading messages from the subscription in a loop
	go psTxsPool.readLoop()
	return psTxsPool, nil
}

// publish sends a PoolL2Tx to the pubsub topic.
func (psTxsPool *pubSubTxsPool) publish(tx common.PoolL2Tx) error {
	msgBytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	log.Debug("publishing tx")
	return psTxsPool.topic.Publish(psTxsPool.ctx, msgBytes)
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (psTxsPool *pubSubTxsPool) readLoop() {
	handler := func(tx common.PoolL2Tx) {
		if err := psTxsPool.handler(tx); err != nil {
			log.Warn(err.Error())
		}
	}
	for {
		msg, err := psTxsPool.sub.Next(psTxsPool.ctx)
		if err != nil {
			return
		}
		// Only forward messages delivered by others
		if msg.ReceivedFrom == psTxsPool.self {
			continue
		}
		log.Debug("Received tx from ", msg.ReceivedFrom.Pretty())
		tx := &common.PoolL2Tx{}
		err = json.Unmarshal(msg.Data, tx)
		if err != nil {
			log.Warn(err)
			continue
		}
		log.Infof("Received tx from %s: %s", msg.ReceivedFrom.Pretty(), tx.TxID.String())
		// Handle tx
		go handler(*tx)
	}
}
