package common

import (
	"time"

	eth "github.com/ethereum/go-ethereum/common"
)

// Block represents of an Ethereum block
type Block struct {
	EthBlockNum uint64    `meddler:"eth_block_num"`
	Timestamp   time.Time `meddler:"timestamp"`
	Hash        eth.Hash  `meddler:"hash"`
}
