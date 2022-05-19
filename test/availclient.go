package test

import (
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
)

type AvailClient struct{}

func NewAvailClient() *AvailClient {
	return &AvailClient{}
}

func (cl *AvailClient) GetLastBlock() (*common.BlockAvail, error) {
	return nil, nil
}

func (cl *AvailClient) GetBlockByNumber(num uint64) (*common.BlockAvail, error) {
	return nil, nil
}

func (cl *AvailClient) SendTxs(stateRoot *big.Int, l1UserTxs, l1CoordTxs []common.L1Tx, l2Txs []common.L2Tx) error {
	return nil
}
