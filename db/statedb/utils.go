package statedb

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// TODO
func (s *StateDB) getIdxByEthAddr(addr ethCommon.Address) common.Idx {
	return common.Idx(0)
}

// TODO
func (s *StateDB) getIdxByBJJ(pk *babyjub.PublicKey) common.Idx {
	return common.Idx(0)
}
