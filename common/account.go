package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
)

const NLEAFELEMS = 4

// Account is a struct that gives information of the holdings of an address and a specific token. Is the data structure that generates the Value stored in the leaf of the MerkleTree
type Account struct {
	TokenID   TokenID
	Nonce     Nonce    // max of 40 bits used
	Balance   *big.Int // max of 192 bits used
	PublicKey *babyjub.PublicKey
	EthAddr   ethCommon.Address
}

func (a *Account) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "PublicKey: %s..., ", a.PublicKey.String()[:10])
	fmt.Fprintf(buf, "EthAddr: %s..., ", a.EthAddr.String()[:10])
	fmt.Fprintf(buf, "TokenID: %v, ", a.TokenID)
	fmt.Fprintf(buf, "Nonce: %d, ", a.Nonce)
	fmt.Fprintf(buf, "Balance: %s, ", a.Balance.String())
	return buf.String()
}

// Bytes returns the bytes representing the Account, in a way that each BigInt is represented by 32 bytes, in spite of the BigInt could be represented in less bytes (due a small big.Int), so in this way each BigInt is always 32 bytes and can be automatically parsed from a byte array.
func (a *Account) Bytes() ([32 * NLEAFELEMS]byte, error) {
	var b [32 * NLEAFELEMS]byte

	if a.Nonce > 0xffffffffff {
		return b, fmt.Errorf("%s Nonce", ErrNumOverflow)
	}
	if len(a.Balance.Bytes()) > 24 {
		return b, fmt.Errorf("%s Balance", ErrNumOverflow)
	}

	nonceBytes, err := a.Nonce.Bytes()
	if err != nil {
		return b, err
	}

	copy(b[0:4], a.TokenID.Bytes())
	copy(b[4:9], nonceBytes[:])
	if babyjub.PointCoordSign(a.PublicKey.X) {
		b[10] = 1
	}
	copy(b[32:64], SwapEndianness(a.Balance.Bytes())) // SwapEndianness, as big.Int uses BigEndian
	copy(b[64:96], SwapEndianness(a.PublicKey.Y.Bytes()))
	copy(b[96:116], a.EthAddr.Bytes())

	return b, nil
}

// BigInts returns the [5]*big.Int, where each *big.Int is inside the Finite Field
func (a *Account) BigInts() ([NLEAFELEMS]*big.Int, error) {
	e := [NLEAFELEMS]*big.Int{}

	b, err := a.Bytes()
	if err != nil {
		return e, err
	}

	e[0] = new(big.Int).SetBytes(SwapEndianness(b[0:32]))
	e[1] = new(big.Int).SetBytes(SwapEndianness(b[32:64]))
	e[2] = new(big.Int).SetBytes(SwapEndianness(b[64:96]))
	e[3] = new(big.Int).SetBytes(SwapEndianness(b[96:128]))

	return e, nil
}

// HashValue returns the value of the Account, which is the Poseidon hash of its *big.Int representation
func (a *Account) HashValue() (*big.Int, error) {
	b0 := big.NewInt(0)
	toHash := []*big.Int{b0, b0, b0, b0, b0, b0}
	lBI, err := a.BigInts()
	if err != nil {
		return nil, err
	}
	copy(toHash[:], lBI[:])

	v, err := poseidon.Hash(toHash)
	return v, err
}

// AccountFromBigInts returns a Account from a [5]*big.Int
func AccountFromBigInts(e [NLEAFELEMS]*big.Int) (*Account, error) {
	if !cryptoUtils.CheckBigIntArrayInField(e[:]) {
		return nil, ErrNotInFF
	}
	var b [32 * NLEAFELEMS]byte
	copy(b[0:32], SwapEndianness(e[0].Bytes())) // SwapEndianness, as big.Int uses BigEndian
	copy(b[32:64], SwapEndianness(e[1].Bytes()))
	copy(b[64:96], SwapEndianness(e[2].Bytes()))
	copy(b[96:128], SwapEndianness(e[3].Bytes()))

	return AccountFromBytes(b)
}

// AccountFromBytes returns a Account from a byte array
func AccountFromBytes(b [32 * NLEAFELEMS]byte) (*Account, error) {
	tokenID := binary.LittleEndian.Uint32(b[0:4])
	var nonceBytes5 [5]byte
	copy(nonceBytes5[:], b[4:9])
	nonce := NonceFromBytes(nonceBytes5)
	sign := b[10] == 1
	balance := new(big.Int).SetBytes(SwapEndianness(b[32:56])) // b[32:56], as Balance is 192 bits (24 bytes)
	if !bytes.Equal(b[56:64], []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		return nil, fmt.Errorf("%s Balance", ErrNumOverflow)
	}
	ay := new(big.Int).SetBytes(SwapEndianness(b[64:96]))
	pkPoint, err := babyjub.PointFromSignAndY(sign, ay)
	if err != nil {
		return nil, err
	}
	publicKey := babyjub.PublicKey(*pkPoint)
	ethAddr := ethCommon.BytesToAddress(b[96:116])

	if !cryptoUtils.CheckBigIntInField(balance) {
		return nil, ErrNotInFF
	}
	if !cryptoUtils.CheckBigIntInField(ay) {
		return nil, ErrNotInFF
	}

	a := Account{
		TokenID:   TokenID(tokenID),
		Nonce:     nonce,
		Balance:   balance,
		PublicKey: &publicKey,
		EthAddr:   ethAddr,
	}
	return &a, nil
}
