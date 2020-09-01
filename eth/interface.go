package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
)

// ClientInterface is the eth Client interface used by hermez-node modules to
// interact with Ethereum Blockchain and smart contracts.
type ClientInterface interface {
	CurrentBlock() (*big.Int, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	BlockByNumber(context.Context, *big.Int) (*common.Block, error)
	ForgeCall(*common.CallDataForge) ([]byte, error)
}
