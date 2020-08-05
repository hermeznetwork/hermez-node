package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/poseidon"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
)

const NLEAFELEMS = 4

// Leaf is the data structure stored in the Leaf of the MerkleTree
type Leaf struct {
	TokenID TokenID
	Nonce   uint64   // max of 40 bits used
	Balance *big.Int // max of 192 bits used
	Sign    bool
	Ay      *big.Int
	EthAddr eth.Address
}

// Bytes returns the bytes representing the Leaf, in a way that each BigInt is represented by 32 bytes, in spite of the BigInt could be represented in less bytes (due a small big.Int), so in this way each BigInt is always 32 bytes and can be automatically parsed from a byte array.
func (l *Leaf) Bytes() ([32 * NLEAFELEMS]byte, error) {
	var b [32 * NLEAFELEMS]byte

	if l.Nonce >= uint64(math.Pow(2, 40)) {
		return b, fmt.Errorf("%s Nonce", ErrNumOverflow)
	}
	if len(l.Balance.Bytes()) > 24 {
		return b, fmt.Errorf("%s Balance", ErrNumOverflow)
	}

	var tokenIDBytes [4]byte
	binary.LittleEndian.PutUint32(tokenIDBytes[:], uint32(l.TokenID))
	var nonceBytes [8]byte
	binary.LittleEndian.PutUint64(nonceBytes[:], l.Nonce)

	copy(b[0:4], tokenIDBytes[:])
	copy(b[4:9], nonceBytes[:])
	if l.Sign {
		b[10] = 1
	}
	copy(b[32:64], SwapEndianness(l.Balance.Bytes())) // SwapEndianness, as big.Int uses BigEndian
	copy(b[64:96], SwapEndianness(l.Ay.Bytes()))
	copy(b[96:116], l.EthAddr.Bytes())

	return b, nil
}

// BigInts returns the [5]*big.Int, where each *big.Int is inside the Finite Field
func (l *Leaf) BigInts() ([NLEAFELEMS]*big.Int, error) {
	e := [NLEAFELEMS]*big.Int{}

	b, err := l.Bytes()
	if err != nil {
		return e, err
	}

	e[0] = new(big.Int).SetBytes(SwapEndianness(b[0:32]))
	e[1] = new(big.Int).SetBytes(SwapEndianness(b[32:64]))
	e[2] = new(big.Int).SetBytes(SwapEndianness(b[64:96]))
	e[3] = new(big.Int).SetBytes(SwapEndianness(b[96:128]))

	return e, nil
}

// HashValue returns the value of the Leaf, which is the Poseidon hash of its *big.Int representation
func (l *Leaf) HashValue() (*big.Int, error) {
	toHash := [poseidon.T]*big.Int{}
	lBI, err := l.BigInts()
	if err != nil {
		return nil, err
	}
	copy(toHash[:], lBI[:])

	v, err := poseidon.Hash(toHash)
	return v, err
}

// LeafFromBigInts returns a Leaf from a [5]*big.Int
func LeafFromBigInts(e [NLEAFELEMS]*big.Int) (*Leaf, error) {
	if !cryptoUtils.CheckBigIntArrayInField(e[:]) {
		return nil, ErrNotInFF
	}
	var b [32 * NLEAFELEMS]byte
	copy(b[0:32], SwapEndianness(e[0].Bytes())) // SwapEndianness, as big.Int uses BigEndian
	copy(b[32:64], SwapEndianness(e[1].Bytes()))
	copy(b[64:96], SwapEndianness(e[2].Bytes()))
	copy(b[96:128], SwapEndianness(e[3].Bytes()))

	return LeafFromBytes(b)
}

// LeafFromBytes returns a Leaf from a byte array
func LeafFromBytes(b [32 * NLEAFELEMS]byte) (*Leaf, error) {
	tokenID := binary.LittleEndian.Uint32(b[0:4])
	var nonceBytes [8]byte
	copy(nonceBytes[:], b[4:9])
	nonce := binary.LittleEndian.Uint64(nonceBytes[:])
	if nonce >= uint64(math.Pow(2, 40)) {
		return nil, fmt.Errorf("%s Nonce", ErrNumOverflow)
	}
	sign := b[10] == 1
	balance := new(big.Int).SetBytes(SwapEndianness(b[32:56])) // b[32:56], as Balance is 192 bits (24 bytes)
	if !bytes.Equal(b[56:64], []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		return nil, fmt.Errorf("%s Balance", ErrNumOverflow)
	}
	ay := new(big.Int).SetBytes(SwapEndianness(b[64:96]))
	ethAddr := eth.BytesToAddress(b[96:116])

	if !cryptoUtils.CheckBigIntInField(balance) {
		return nil, ErrNotInFF
	}
	if !cryptoUtils.CheckBigIntInField(ay) {
		return nil, ErrNotInFF
	}

	l := Leaf{
		TokenID: TokenID(tokenID),
		Nonce:   nonce,
		Balance: balance,
		Sign:    sign,
		Ay:      ay,
		EthAddr: ethAddr,
	}
	return &l, nil
}
