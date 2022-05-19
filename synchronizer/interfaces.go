package synchronizer

import (
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
)

type availClient interface {
	GetLastBlock() (*common.BlockAvail, error)
	GetBlockByNumber(num uint64) (*common.BlockAvail, error)
	SendTxs(stateRoot *big.Int, l1UserTxs, l1CoordTxs []common.L1Tx, l2Txs []common.L2Tx) error
}
