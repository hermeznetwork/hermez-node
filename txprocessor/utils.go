package txprocessor

import (
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

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

// formatAccumulatedFees returns an array of [nFeeAccounts]*big.Int containing
// the balance of each FeeAccount, taken from the 'collectedFees' map, in the
// order of the 'orderTokenIDs'
func formatAccumulatedFees(collectedFees map[common.TokenID]*big.Int, orderTokenIDs []*big.Int) []*big.Int {
	accFeeOut := make([]*big.Int, len(orderTokenIDs))
	for i := 0; i < len(orderTokenIDs); i++ {
		tokenID := common.TokenIDFromBigInt(orderTokenIDs[i])
		if _, ok := collectedFees[tokenID]; ok {
			accFeeOut[i] = new(big.Int).Set(collectedFees[tokenID])
		} else {
			accFeeOut[i] = big.NewInt(0)
		}
	}
	return accFeeOut
}
