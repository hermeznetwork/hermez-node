package common

import (
	"bytes"
	"encoding/hex"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

// SwapEndianness swaps the order of the bytes in the slice.
func SwapEndianness(b []byte) []byte {
	o := make([]byte, len(b))
	for i := range b {
		o[len(b)-1-i] = b[i]
	}
	return o
}

// EthAddrToBigInt returns a *big.Int from a given ethereum common.Address.
func EthAddrToBigInt(a ethCommon.Address) *big.Int {
	return new(big.Int).SetBytes(a.Bytes())
}

// BJJFromStringWithChecksum parses a hex string in Hermez format (which has
// the Hermez checksum at the last byte, and is encoded in BigEndian) and
// returns the corresponding *babyjub.PublicKey. This method is not part of the
// spec, is used for importing javascript test vectors data.
func BJJFromStringWithChecksum(s string) (babyjub.PublicKeyComp, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return EmptyBJJComp, tracerr.Wrap(err)
	}
	pkBytes := SwapEndianness(b)
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkBytes[:])
	return pkComp, nil
}

// CopyBigInt returns a copy of the big int
func CopyBigInt(a *big.Int) *big.Int {
	return new(big.Int).SetBytes(a.Bytes())
}

// RmEndingZeroes is used to convert the Siblings from a CircomProof into
// Siblings of a merkletree Proof compatible with the js version. This method
// should be used only if it exist an already generated CircomProof compatible
// with circom circuits and a CircomProof compatible with SmartContracts is
// needed. If the proof is not generated yet, this method should not be needed
// and should be used mt.GenerateSCVerifierProof to directly generate the
// CircomProof for the SmartContracts.
func RmEndingZeroes(siblings []*merkletree.Hash) []*merkletree.Hash {
	pos := 0
	for i := len(siblings) - 1; i >= 0; i-- {
		if !bytes.Equal(siblings[i].Bytes(), merkletree.HashZero.Bytes()) {
			pos = i + 1
			break
		}
	}
	return siblings[:pos]
}

// TokensToUSD is a helper function to calculate the USD value of a certain
// amount of tokens considering the normalized token price (which is the price
// commonly reported by exhanges)
func TokensToUSD(amount *big.Int, decimals uint64, valueUSD float64) float64 {
	amountF := new(big.Float).SetInt(amount)
	// Divide by 10^decimals to normalize the amount
	baseF := new(big.Float).SetInt(new(big.Int).Exp(
		big.NewInt(10), big.NewInt(int64(decimals)), nil)) //nolint:gomnd
	amountF.Mul(amountF, big.NewFloat(valueUSD))
	amountF.Quo(amountF, baseF)
	amountUSD, _ := amountF.Float64()
	return amountUSD
}
