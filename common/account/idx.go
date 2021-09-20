package account

import (
	"encoding/binary"
	"fmt"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/tracerr"
	"math/big"
	"strconv"
)

const (
	// IdxBytesLen idx bytes
	IdxBytesLen = 6

	// maxIdxValue is the maximum value that Idx can have (48 bits:
	// maxIdxValue=2**48-1)
	maxIdxValue = 0xffffffffffff

	// IdxUserThreshold is a Idx type value that determines the threshold
	// from the User Idxs can be
	IdxUserThreshold = Idx(UserThreshold)
)

// Idx represents the account Index in the MerkleTree
type Idx uint64

// String returns a string representation of the Idx
func (idx Idx) String() string {
	return strconv.Itoa(int(idx))
}

// Bytes returns a byte array representing the Idx
func (idx Idx) Bytes() ([6]byte, error) {
	if idx > maxIdxValue {
		return [6]byte{}, tracerr.Wrap(common.ErrIdxOverflow)
	}
	var idxBytes [8]byte
	binary.BigEndian.PutUint64(idxBytes[:], uint64(idx))
	var b [6]byte
	copy(b[:], idxBytes[2:])
	return b, nil
}

// BigInt returns a *big.Int representing the Idx
func (idx Idx) BigInt() *big.Int {
	return big.NewInt(int64(idx))
}

// IdxFromBytes returns Idx from a byte array
func IdxFromBytes(b []byte) (Idx, error) {
	if len(b) != IdxBytesLen {
		return 0, tracerr.Wrap(fmt.Errorf("can not parse Idx, bytes len %d, expected %d",
			len(b), IdxBytesLen))
	}
	var idxBytes [8]byte
	copy(idxBytes[2:], b[:])
	idx := binary.BigEndian.Uint64(idxBytes[:])
	return Idx(idx), nil
}

// IdxFromBigInt converts a *big.Int to Idx type
func IdxFromBigInt(b *big.Int) (Idx, error) {
	if b.Int64() > maxIdxValue {
		return 0, tracerr.Wrap(common.ErrNumOverflow)
	}
	return Idx(uint64(b.Int64())), nil
}

// IdxNonce is a pair of Idx and Nonce representing an account
type IdxNonce struct {
	Idx   Idx         `db:"idx"`
	Nonce nonce.Nonce `db:"nonce"`
}
