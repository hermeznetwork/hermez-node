package common

import (
	"errors"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/multiformats/go-multiaddr"
)

// Coordinator represents a Hermez network coordinator who wins an auction for an specific slot
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this
// may change a lot.
type Coordinator struct {
	// Bidder is the address of the bidder
	Bidder ethCommon.Address `meddler:"bidder_addr"`
	// Forger is the address of the forger
	Forger ethCommon.Address `meddler:"forger_addr"`
	// EthBlockNum is the block in which the coordinator was registered
	EthBlockNum int64 `meddler:"eth_block_num"`
	// URL of the coordinators API
	URL string `meddler:"url"`
}

// CoordinatorsNetworkPort is the port used by coordinators for libp2p
const CoordinatorsNetworkPort = "3598"

// P2PAddr returns a multi address that allows to connect with the Coordinator using libp2p2
func (coord Coordinator) P2PAddr() (multiaddr.Multiaddr, error) {
	/*
		addr must be one of the following formats:
		- /dns/<URL>/tcp/3598/p2p/<libp2p ID>
		- /ip4/<IPv4>/tcp/3598/p2p/<libp2p ID>

		TODO:
		- parse the coordinator URL and decide to use ip4 or dns
		- Use API to get libp2p ID OR find a way to derivate it from Ethereum public key
	*/
	return nil, errors.New("not implemented yet")
}
