package eth

import "github.com/hermeznetwork/hermez-node/common"

type EthClient struct {
}

func NewEthClient() *EthClient {
	// TODO
	return &EthClient{}
}
func (ec *EthClient) ForgeCall(callData *common.CallDataForge) ([]byte, error) {
	// TODO this depends on the smart contracts, once are ready this will be updated
	return nil, nil
}
