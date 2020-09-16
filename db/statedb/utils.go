package statedb

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

// GetIdxByEthAddr returns the smallest Idx in the StateDB for the given
// Ethereum Address. Will return 0 in case that Idx is not found in the
// StateDB.
func (s *StateDB) GetIdxByEthAddr(addr ethCommon.Address) common.Idx {
	// TODO
	return common.Idx(0)
}

// GetIdxByBJJ returns the smallest Idx in the StateDB for the given BabyJubJub
// PublicKey. Will return 0 in case that Idx is not found in the StateDB.
func (s *StateDB) GetIdxByBJJ(pk *babyjub.PublicKey) common.Idx {
	// TODO
	return common.Idx(0)
}

// GetIdxByEthAddrBJJ returns the smallest Idx in the StateDB for the given
// Ethereum Address AND the given BabyJubJub PublicKey. If `addr` is the zero
// address, it's ignored in the query.  If `pk` is nil, it's ignored in the
// query.  Will return 0 in case that Idx is not found in the StateDB.
func (s *StateDB) GetIdxByEthAddrBJJ(addr ethCommon.Address, pk *babyjub.PublicKey) common.Idx {
	// TODO
	return common.Idx(0)
}

func siblingsToZKInputFormat(s []*merkletree.Hash) []*big.Int {
	b := make([]*big.Int, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = s[i].BigInt()
	}
	return b
}

// BJJCompressedTo256BigInts returns a [256]*big.Int array with the bit
// representation of the babyjub.PublicKeyComp
func BJJCompressedTo256BigInts(pkComp babyjub.PublicKeyComp) [256]*big.Int {
	var r [256]*big.Int
	b := pkComp[:]

	for i := 0; i < 256; i++ {
		if b[i/8]&(1<<(i%8)) == 0 {
			r[i] = big.NewInt(0)
		} else {
			r[i] = big.NewInt(1)
		}
	}

	return r
}
