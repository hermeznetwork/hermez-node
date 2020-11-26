package apitypes

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/ztrue/tracerr"
)

// BigIntStr is used to scan/value *big.Int directly into strings from/to sql DBs.
// It assumes that *big.Int are inserted/fetched to/from the DB using the BigIntMeddler meddler
// defined at github.com/hermeznetwork/hermez-node/db
type BigIntStr string

// NewBigIntStr creates a *BigIntStr from a *big.Int.
// If the provided bigInt is nil the returned *BigIntStr will also be nil
func NewBigIntStr(bigInt *big.Int) *BigIntStr {
	if bigInt == nil {
		return nil
	}
	bigIntStr := BigIntStr(bigInt.String())
	return &bigIntStr
}

// Scan implements Scanner for database/sql
func (b *BigIntStr) Scan(src interface{}) error {
	srcBytes, ok := src.([]byte)
	if !ok {
		return tracerr.Wrap(fmt.Errorf("can't scan %T into apitypes.BigIntStr", src))
	}
	// bytes to *big.Int
	bigInt := new(big.Int).SetBytes(srcBytes)
	// *big.Int to BigIntStr
	bigIntStr := NewBigIntStr(bigInt)
	if bigIntStr == nil {
		return nil
	}
	*b = *bigIntStr
	return nil
}

// Value implements valuer for database/sql
func (b BigIntStr) Value() (driver.Value, error) {
	// string to *big.Int
	bigInt, ok := new(big.Int).SetString(string(b), 10)
	if !ok || bigInt == nil {
		return nil, tracerr.Wrap(errors.New("invalid representation of a *big.Int"))
	}
	// *big.Int to bytes
	return bigInt.Bytes(), nil
}

// StrBigInt is used to unmarshal BigIntStr directly into an alias of big.Int
type StrBigInt big.Int

// UnmarshalText unmarshals a StrBigInt
func (s *StrBigInt) UnmarshalText(text []byte) error {
	bi, ok := (*big.Int)(s).SetString(string(text), 10)
	if !ok {
		return tracerr.Wrap(fmt.Errorf("could not unmarshal %s into a StrBigInt", text))
	}
	*s = StrBigInt(*bi)
	return nil
}

// CollectedFees is used to retrieve common.batch.CollectedFee from the DB
type CollectedFees map[common.TokenID]BigIntStr

// UnmarshalJSON unmarshals a json representation of map[common.TokenID]*big.Int
func (c *CollectedFees) UnmarshalJSON(text []byte) error {
	bigIntMap := make(map[common.TokenID]*big.Int)
	if err := json.Unmarshal(text, &bigIntMap); err != nil {
		return tracerr.Wrap(err)
	}
	*c = CollectedFees(make(map[common.TokenID]BigIntStr))
	for k, v := range bigIntMap {
		bStr := NewBigIntStr(v)
		(CollectedFees(*c)[k]) = *bStr
	}
	// *c = CollectedFees(bStrMap)
	return nil
}

// HezEthAddr is used to scan/value Ethereum Address directly into strings that follow the Ethereum address hez fotmat (^hez:0x[a-fA-F0-9]{40}$) from/to sql DBs.
// It assumes that Ethereum Address are inserted/fetched to/from the DB using the default Scan/Value interface
type HezEthAddr string

// NewHezEthAddr creates a HezEthAddr from an Ethereum addr
func NewHezEthAddr(addr ethCommon.Address) HezEthAddr {
	return HezEthAddr("hez:" + addr.String())
}

// ToEthAddr returns an Ethereum Address created from HezEthAddr
func (a HezEthAddr) ToEthAddr() (ethCommon.Address, error) {
	addrStr := strings.TrimPrefix(string(a), "hez:")
	var addr ethCommon.Address
	return addr, addr.UnmarshalText([]byte(addrStr))
}

// Scan implements Scanner for database/sql
func (a *HezEthAddr) Scan(src interface{}) error {
	ethAddr := &ethCommon.Address{}
	if err := ethAddr.Scan(src); err != nil {
		return tracerr.Wrap(err)
	}
	if ethAddr == nil {
		return nil
	}
	*a = NewHezEthAddr(*ethAddr)
	return nil
}

// Value implements valuer for database/sql
func (a HezEthAddr) Value() (driver.Value, error) {
	ethAddr, err := a.ToEthAddr()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return ethAddr.Value()
}

// StrHezEthAddr is used to unmarshal HezEthAddr directly into an alias of ethCommon.Address
type StrHezEthAddr ethCommon.Address

// UnmarshalText unmarshals a StrHezEthAddr
func (s *StrHezEthAddr) UnmarshalText(text []byte) error {
	withoutHez := strings.TrimPrefix(string(text), "hez:")
	var addr ethCommon.Address
	if err := addr.UnmarshalText([]byte(withoutHez)); err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrHezEthAddr(addr)
	return nil
}

// HezBJJ is used to scan/value *babyjub.PublicKey directly into strings that follow the BJJ public key hez fotmat (^hez:[A-Za-z0-9_-]{44}$) from/to sql DBs.
// It assumes that *babyjub.PublicKey are inserted/fetched to/from the DB using the default Scan/Value interface
type HezBJJ string

// NewHezBJJ creates a HezBJJ from a *babyjub.PublicKey.
// Calling this method with a nil bjj causes panic
func NewHezBJJ(bjj *babyjub.PublicKey) HezBJJ {
	pkComp := [32]byte(bjj.Compress())
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return HezBJJ("hez:" + base64.RawURLEncoding.EncodeToString(bjjSum))
}

func hezStrToBJJ(s string) (*babyjub.PublicKey, error) {
	const decodedLen = 33
	const encodedLen = 44
	formatErr := errors.New("invalid BJJ format. Must follow this regex: ^hez:[A-Za-z0-9_-]{44}$")
	encoded := strings.TrimPrefix(s, "hez:")
	if len(encoded) != encodedLen {
		return nil, formatErr
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, formatErr
	}
	if len(decoded) != decodedLen {
		return nil, formatErr
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return nil, tracerr.Wrap(errors.New("checksum verification failed"))
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	return bjjComp.Decompress()
}

// ToBJJ returns a *babyjub.PublicKey created from HezBJJ
func (b HezBJJ) ToBJJ() (*babyjub.PublicKey, error) {
	return hezStrToBJJ(string(b))
}

// Scan implements Scanner for database/sql
func (b *HezBJJ) Scan(src interface{}) error {
	bjj := &babyjub.PublicKey{}
	if err := bjj.Scan(src); err != nil {
		return tracerr.Wrap(err)
	}
	if bjj == nil {
		return nil
	}
	*b = NewHezBJJ(bjj)
	return nil
}

// Value implements valuer for database/sql
func (b HezBJJ) Value() (driver.Value, error) {
	bjj, err := b.ToBJJ()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return bjj.Value()
}

// StrHezBJJ is used to unmarshal HezBJJ directly into an alias of babyjub.PublicKey
type StrHezBJJ babyjub.PublicKey

// UnmarshalText unmarshals a StrHezBJJ
func (s *StrHezBJJ) UnmarshalText(text []byte) error {
	bjj, err := hezStrToBJJ(string(text))
	if err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrHezBJJ(*bjj)
	return nil
}

// HezIdx is used to value common.Idx directly into strings that follow the Idx key hez fotmat (hez:tokenSymbol:idx) to sql DBs.
// Note that this can only be used to insert to DB since there is no way to automaticaly read from the DB since it needs the tokenSymbol
type HezIdx string

// StrHezIdx is used to unmarshal HezIdx directly into an alias of common.Idx
type StrHezIdx common.Idx

// UnmarshalText unmarshals a StrHezIdx
func (s *StrHezIdx) UnmarshalText(text []byte) error {
	withoutHez := strings.TrimPrefix(string(text), "hez:")
	splitted := strings.Split(withoutHez, ":")
	const expectedLen = 2
	if len(splitted) != expectedLen {
		return tracerr.Wrap(fmt.Errorf("can not unmarshal %s into StrHezIdx", text))
	}
	idxInt, err := strconv.Atoi(splitted[1])
	if err != nil {
		return tracerr.Wrap(err)
	}
	*s = StrHezIdx(common.Idx(idxInt))
	return nil
}

// EthSignature is used to scan/value []byte representing an Ethereum signature directly into strings from/to sql DBs.
type EthSignature string

// NewEthSignature creates a *EthSignature from []byte
// If the provided signature is nil the returned *EthSignature will also be nil
func NewEthSignature(signature []byte) *EthSignature {
	if signature == nil {
		return nil
	}
	ethSignature := EthSignature("0x" + hex.EncodeToString(signature))
	return &ethSignature
}

// Scan implements Scanner for database/sql
func (e *EthSignature) Scan(src interface{}) error {
	if srcStr, ok := src.(string); ok {
		// src is a string
		*e = *(NewEthSignature([]byte(srcStr)))
		return nil
	} else if srcBytes, ok := src.([]byte); ok {
		// src is []byte
		*e = *(NewEthSignature(srcBytes))
		return nil
	} else {
		// unexpected src
		return tracerr.Wrap(fmt.Errorf("can't scan %T into apitypes.EthSignature", src))
	}
}

// Value implements valuer for database/sql
func (e EthSignature) Value() (driver.Value, error) {
	without0x := strings.TrimPrefix(string(e), "0x")
	return hex.DecodeString(without0x)
}

// UnmarshalText unmarshals a StrEthSignature
func (e *EthSignature) UnmarshalText(text []byte) error {
	without0x := strings.TrimPrefix(string(text), "0x")
	signature, err := hex.DecodeString(without0x)
	if err != nil {
		return tracerr.Wrap(err)
	}
	*e = EthSignature([]byte(signature))
	return nil
}
