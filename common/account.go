package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
)

const (
	// NLeafElems is the number of elements for a leaf
	NLeafElems = 4
	// maxNonceValue is the maximum value that the Account.Nonce can have (40 bits: maxNonceValue=2**40-1)
	maxNonceValue = 0xffffffffff
	// maxBalanceBytes is the maximum bytes that can use the Account.Balance *big.Int
	maxBalanceBytes = 24

	// IdxBytesLen idx bytes
	IdxBytesLen = 6
	// maxIdxValue is the maximum value that Idx can have (48 bits: maxIdxValue=2**48-1)
	maxIdxValue = 0xffffffffffff

	// UserThreshold determines the threshold from the User Idxs can be
	UserThreshold = 256
	// IdxUserThreshold is a Idx type value that determines the threshold
	// from the User Idxs can be
	IdxUserThreshold = Idx(UserThreshold)
)

var (
	// FFAddr is used to check if an ethereum address is 0xff..ff
	FFAddr = ethCommon.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")
	// EmptyAddr is used to check if an ethereum address is 0
	EmptyAddr = ethCommon.HexToAddress("0x0000000000000000000000000000000000000000")
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
		return [6]byte{}, ErrIdxOverflow
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
		return 0, fmt.Errorf("can not parse Idx, bytes len %d, expected %d", len(b), IdxBytesLen)
	}
	var idxBytes [8]byte
	copy(idxBytes[2:], b[:])
	idx := binary.BigEndian.Uint64(idxBytes[:])
	return Idx(idx), nil
}

// IdxFromBigInt converts a *big.Int to Idx type
func IdxFromBigInt(b *big.Int) (Idx, error) {
	if b.Int64() > maxIdxValue {
		return 0, ErrNumOverflow
	}
	return Idx(uint64(b.Int64())), nil
}

// Nonce represents the nonce value in a uint64, which has the method Bytes that returns a byte array of length 5 (40 bits).
type Nonce uint64

// Bytes returns a byte array of length 5 representing the Nonce
func (n Nonce) Bytes() ([5]byte, error) {
	if n > maxNonceValue {
		return [5]byte{}, ErrNonceOverflow
	}
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], uint64(n))
	var b [5]byte
	copy(b[:], nonceBytes[3:])
	return b, nil
}

// BigInt returns the *big.Int representation of the Nonce value
func (n Nonce) BigInt() *big.Int {
	return big.NewInt(int64(n))
}

// NonceFromBytes returns Nonce from a [5]byte
func NonceFromBytes(b [5]byte) Nonce {
	var nonceBytes [8]byte
	copy(nonceBytes[3:], b[:])
	nonce := binary.BigEndian.Uint64(nonceBytes[:])
	return Nonce(nonce)
}

// Account is a struct that gives information of the holdings of an address and a specific token. Is the data structure that generates the Value stored in the leaf of the MerkleTree
type Account struct {
	Idx       Idx                `meddler:"idx"`
	TokenID   TokenID            `meddler:"token_id"`
	BatchNum  BatchNum           `meddler:"batch_num"`
	PublicKey *babyjub.PublicKey `meddler:"bjj"`
	EthAddr   ethCommon.Address  `meddler:"eth_addr"`
	Nonce     Nonce              `meddler:"-"` // max of 40 bits used
	Balance   *big.Int           `meddler:"-"` // max of 192 bits used
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

// Bytes returns the bytes representing the Account, in a way that each BigInt
// is represented by 32 bytes, in spite of the BigInt could be represented in
// less bytes (due a small big.Int), so in this way each BigInt is always 32
// bytes and can be automatically parsed from a byte array.
func (a *Account) Bytes() ([32 * NLeafElems]byte, error) {
	var b [32 * NLeafElems]byte

	if a.Nonce > maxNonceValue {
		return b, fmt.Errorf("%s Nonce", ErrNumOverflow)
	}
	if len(a.Balance.Bytes()) > maxBalanceBytes {
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
	copy(b[32:64], SwapEndianness(a.Balance.Bytes()))
	copy(b[64:96], SwapEndianness(a.PublicKey.Y.Bytes()))
	copy(b[96:116], a.EthAddr.Bytes())

	return b, nil
}

// BigInts returns the [5]*big.Int, where each *big.Int is inside the Finite Field
func (a *Account) BigInts() ([NLeafElems]*big.Int, error) {
	e := [NLeafElems]*big.Int{}

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
func AccountFromBigInts(e [NLeafElems]*big.Int) (*Account, error) {
	if !cryptoUtils.CheckBigIntArrayInField(e[:]) {
		return nil, ErrNotInFF
	}
	var b [32 * NLeafElems]byte
	copy(b[0:32], SwapEndianness(e[0].Bytes())) // SwapEndianness, as big.Int uses BigEndian
	copy(b[32:64], SwapEndianness(e[1].Bytes()))
	copy(b[64:96], SwapEndianness(e[2].Bytes()))
	copy(b[96:128], SwapEndianness(e[3].Bytes()))

	return AccountFromBytes(b)
}

// AccountFromBytes returns a Account from a byte array
func AccountFromBytes(b [32 * NLeafElems]byte) (*Account, error) {
	tokenID, err := TokenIDFromBytes(b[0:4])
	if err != nil {
		return nil, err
	}
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
