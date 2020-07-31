package common

import (
	"time"

	eth "github.com/ethereum/go-ethereum/common"
)

// Block represents of an Ethereum block
type Block struct {
	EthBlockNum uint64
	Timestamp   time.Time
	Hash        eth.Hash
	PrevHash    eth.Hash
}
