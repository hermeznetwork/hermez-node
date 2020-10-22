package apitypes

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
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
	// decode base64 src
	var decoded []byte
	var err error
	if srcStr, ok := src.(string); ok {
		// src is a string
		decoded, err = base64.StdEncoding.DecodeString(srcStr)
	} else if srcBytes, ok := src.([]byte); ok {
		// src is []byte
		decoded, err = base64.StdEncoding.DecodeString(string(srcBytes))
	} else {
		// unexpected src
		return fmt.Errorf("can't scan %T into apitypes.BigIntStr", src)
	}
	if err != nil {
		return err
	}
	// decoded bytes to *big.Int
	bigInt := &big.Int{}
	bigInt = bigInt.SetBytes(decoded)
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
	bigInt := &big.Int{}
	bigInt, ok := bigInt.SetString(string(b), 10)
	if !ok || bigInt == nil {
		return nil, errors.New("invalid representation of a *big.Int")
	}
	// *big.Int to base64
	return base64.StdEncoding.EncodeToString(bigInt.Bytes()), nil
}

type CollectedFees map[common.TokenID]BigIntStr

func (c *CollectedFees) UnmarshalJSON(text []byte) error {
	fmt.Println(string(text))
	bigIntMap := make(map[common.TokenID]*big.Int)
	if err := json.Unmarshal(text, &bigIntMap); err != nil {
		return err
	}
	bStrMap := make(map[common.TokenID]BigIntStr)
	for k, v := range bigIntMap {
		bStr := NewBigIntStr(v)
		bStrMap[k] = *bStr
	}
	*c = CollectedFees(bStrMap)
	return nil
	// fmt.Println(string(text))
	// *b = BigIntStr(string(text))
	// return nil
	// bigInt := &big.Int{}
	// if err := bigInt.UnmarshalText(text); err != nil {
	// 	return err
	// }
	// bigIntStr := NewBigIntStr(bigInt)
	// if bigIntStr == nil {
	// 	return nil
	// }
	// *b = *bigIntStr
	// return nil
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
		return err
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
		return nil, err
	}
	return ethAddr.Value()
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

// ToBJJ returns a *babyjub.PublicKey created from HezBJJ
func (b HezBJJ) ToBJJ() (*babyjub.PublicKey, error) {
	const decodedLen = 33
	const encodedLen = 44
	formatErr := errors.New("invalid BJJ format. Must follow this regex: ^hez:[A-Za-z0-9_-]{44}$")
	encoded := strings.TrimPrefix(string(b), "hez:")
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
		return nil, errors.New("checksum verification failed")
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	return bjjComp.Decompress()
}

// Scan implements Scanner for database/sql
func (b *HezBJJ) Scan(src interface{}) error {
	bjj := &babyjub.PublicKey{}
	if err := bjj.Scan(src); err != nil {
		return err
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
		return nil, err
	}
	return bjj.Value()
}
