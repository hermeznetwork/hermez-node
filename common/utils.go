package common

import (
	"encoding/hex"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
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
func BJJFromStringWithChecksum(s string) (*babyjub.PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	pkBytes := SwapEndianness(b)
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkBytes[:])
	return pkComp.Decompress()
}
