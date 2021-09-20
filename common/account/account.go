package account

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	common2 "github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-iden3-crypto/utils"
	"math/big"
)

const (
	// NLeafElems is the number of elements for a leaf
	NLeafElems = 4

	// maxBalanceBytes is the maximum bytes that can use the
	// Account.Balance *big.Int
	maxBalanceBytes = 24

	// UserThreshold determines the threshold from the User Idxs can be
	UserThreshold = 256
)

var (
	// FFAddr is used to check if an ethereum address is 0xff..ff
	FFAddr = common.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")
	// EmptyAddr is used to check if an ethereum address is 0
	EmptyAddr = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

// Account is a struct that gives information of the holdings of an address and
// a specific token. Is the data structure that generates the Value stored in
// the leaf of the MerkleTree
type Account struct {
	Idx      Idx                   `meddler:"idx"`
	TokenID  common2.TokenID       `meddler:"token_id"`
	BatchNum common2.BatchNum      `meddler:"batch_num"`
	BJJ      babyjub.PublicKeyComp `meddler:"bjj"`
	EthAddr  common.Address        `meddler:"eth_addr"`
	Nonce    nonce.Nonce           `meddler:"-"` // max of 40 bits used
	Balance  *big.Int              `meddler:"-"` // max of 192 bits used
}

func (a *Account) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Idx: %v, ", a.Idx)
	fmt.Fprintf(buf, "BJJ: %s..., ", a.BJJ.String()[:10])
	fmt.Fprintf(buf, "EthAddr: %s..., ", a.EthAddr.String()[:10])
	fmt.Fprintf(buf, "TokenID: %v, ", a.TokenID)
	fmt.Fprintf(buf, "Nonce: %d, ", a.Nonce)
	fmt.Fprintf(buf, "Balance: %s, ", a.Balance.String())
	fmt.Fprintf(buf, "BatchNum: %v, ", a.BatchNum)
	return buf.String()
}

// Bytes returns the bytes representing the Account, in a way that each BigInt
// is represented by 32 bytes, in spite of the BigInt could be represented in
// less bytes (due a small big.Int), so in this way each BigInt is always 32
// bytes and can be automatically parsed from a byte array.
func (a *Account) Bytes() ([32 * NLeafElems]byte, error) {
	var b [32 * NLeafElems]byte

	if a.Nonce > nonce.MaxNonceValue {
		return b, tracerr.Wrap(fmt.Errorf("%s Nonce", common2.ErrNumOverflow))
	}
	if len(a.Balance.Bytes()) > maxBalanceBytes {
		return b, tracerr.Wrap(fmt.Errorf("%s Balance", common2.ErrNumOverflow))
	}

	nonceBytes, err := a.Nonce.Bytes()
	if err != nil {
		return b, tracerr.Wrap(err)
	}

	copy(b[28:32], a.TokenID.Bytes())
	copy(b[23:28], nonceBytes[:])

	pkSign, pkY := babyjub.UnpackSignY(a.BJJ)
	if pkSign {
		b[22] = 1
	}
	balanceBytes := a.Balance.Bytes()
	copy(b[64-len(balanceBytes):64], balanceBytes)
	ayBytes := pkY.Bytes()
	copy(b[96-len(ayBytes):96], ayBytes)
	copy(b[108:128], a.EthAddr.Bytes())

	return b, nil
}

// BigInts returns the [5]*big.Int, where each *big.Int is inside the Finite Field
func (a *Account) BigInts() ([NLeafElems]*big.Int, error) {
	e := [NLeafElems]*big.Int{}

	b, err := a.Bytes()
	if err != nil {
		return e, tracerr.Wrap(err)
	}

	e[0] = new(big.Int).SetBytes(b[0:32])
	e[1] = new(big.Int).SetBytes(b[32:64])
	e[2] = new(big.Int).SetBytes(b[64:96])
	e[3] = new(big.Int).SetBytes(b[96:128])

	return e, nil
}

// HashValue returns the value of the Account, which is the Poseidon hash of its
// *big.Int representation
func (a *Account) HashValue() (*big.Int, error) {
	bi, err := a.BigInts()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return poseidon.Hash(bi[:])
}

// AccountFromBigInts returns a Account from a [5]*big.Int
func AccountFromBigInts(e [NLeafElems]*big.Int) (*Account, error) {
	if !utils.CheckBigIntArrayInField(e[:]) {
		return nil, tracerr.Wrap(common2.ErrNotInFF)
	}
	e0B := e[0].Bytes()
	e1B := e[1].Bytes()
	e2B := e[2].Bytes()
	e3B := e[3].Bytes()
	var b [32 * NLeafElems]byte
	copy(b[32-len(e0B):32], e0B)
	copy(b[64-len(e1B):64], e1B)
	copy(b[96-len(e2B):96], e2B)
	copy(b[128-len(e3B):128], e3B)

	return AccountFromBytes(b)
}

// AccountFromBytes returns a Account from a byte array
func AccountFromBytes(b [32 * NLeafElems]byte) (*Account, error) {
	tokenID, err := common2.TokenIDFromBytes(b[28:32])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var nonceBytes5 [5]byte
	copy(nonceBytes5[:], b[23:28])
	nonce := nonce.FromBytes(nonceBytes5)
	sign := b[22] == 1

	balance := new(big.Int).SetBytes(b[40:64])
	// Balance is max of 192 bits (24 bytes)
	if !bytes.Equal(b[32:40], []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		return nil, tracerr.Wrap(fmt.Errorf("%s Balance", common2.ErrNumOverflow))
	}
	ay := new(big.Int).SetBytes(b[64:96])
	publicKeyComp := babyjub.PackSignY(sign, ay)
	ethAddr := common.BytesToAddress(b[108:128])

	if !utils.CheckBigIntInField(balance) {
		return nil, tracerr.Wrap(common2.ErrNotInFF)
	}
	if !utils.CheckBigIntInField(ay) {
		return nil, tracerr.Wrap(common2.ErrNotInFF)
	}

	a := Account{
		TokenID: common2.TokenID(tokenID),
		Nonce:   nonce,
		Balance: balance,
		BJJ:     publicKeyComp,
		EthAddr: ethAddr,
	}
	return &a, nil
}

// AccountUpdate represents an account balance and/or nonce update after a
// processed batch
type AccountUpdate struct {
	EthBlockNum int64            `meddler:"eth_block_num"`
	BatchNum    common2.BatchNum `meddler:"batch_num"`
	Idx         Idx              `meddler:"idx"`
	Nonce       nonce.Nonce      `meddler:"nonce"`
	Balance     *big.Int         `meddler:"balance,bigint"`
}
