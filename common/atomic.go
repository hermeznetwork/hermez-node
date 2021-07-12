package common

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/tracerr"
)

// AtomicGroupIDLen is the length of a Hermez network atomic group
const AtomicGroupIDLen = 32

// AtomicGroupID is the identifier of a Hermez network atomic group
type AtomicGroupID [AtomicGroupIDLen]byte

// EmptyAtomicGroupID represents an empty Hermez network atomic group identifier
var EmptyAtomicGroupID = AtomicGroupID([32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

// CalculateAtomicGroupID calculates the atomic group ID given the identifiers of
// the transactions that conform the atomic group
func CalculateAtomicGroupID(txIDs []TxID) AtomicGroupID {
	txIDConcatenation := make([]byte, TxIDLen*len(txIDs))
	for i, id := range txIDs {
		idBytes := [TxIDLen]byte(id)
		copy(txIDConcatenation[i*TxIDLen:(i+1)*TxIDLen], idBytes[:])
	}
	h := ethCrypto.Keccak256Hash(txIDConcatenation).Bytes()
	var agid AtomicGroupID
	copy(agid[:], h)
	return agid
}

// Scan implements Scanner for database/sql.
func (agid *AtomicGroupID) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return tracerr.Wrap(fmt.Errorf("can't scan %T into AtomicGroupID", src))
	}
	if len(srcB) != AtomicGroupIDLen {
		return tracerr.Wrap(fmt.Errorf("can't scan []byte of len %d into AtomicGroupID, need %d",
			len(srcB), AtomicGroupIDLen))
	}
	copy(agid[:], srcB)
	return nil
}

// Value implements valuer for database/sql.
func (agid AtomicGroupID) Value() (driver.Value, error) {
	return agid[:], nil
}

// String returns a string hexadecimal representation of the AtomicGroupID
func (agid AtomicGroupID) String() string {
	return "0x" + hex.EncodeToString(agid[:])
}

// NewAtomicGroupIDFromString returns a string hexadecimal representation of the AtomicGroupID
func NewAtomicGroupIDFromString(idStr string) (AtomicGroupID, error) {
	agid := AtomicGroupID{}
	idStr = strings.TrimPrefix(idStr, "0x")
	decoded, err := hex.DecodeString(idStr)
	if err != nil {
		return AtomicGroupID{}, tracerr.Wrap(err)
	}
	if len(decoded) != AtomicGroupIDLen {
		return agid, tracerr.Wrap(errors.New("Invalid idStr"))
	}
	copy(agid[:], decoded)
	return agid, nil
}

// MarshalText marshals a AtomicGroupID
func (agid AtomicGroupID) MarshalText() ([]byte, error) {
	return []byte(agid.String()), nil
}

// UnmarshalText unmarshalls a AtomicGroupID
func (agid *AtomicGroupID) UnmarshalText(data []byte) error {
	idStr := string(data)
	id, err := NewAtomicGroupIDFromString(idStr)
	if err != nil {
		return tracerr.Wrap(err)
	}
	*agid = id
	return nil
}

// AtomicGroup represents a set of atomic transactions
type AtomicGroup struct {
	ID  AtomicGroupID `json:"atomicGroupId"`
	Txs []PoolL2Tx    `json:"transactions"`
}

// SetAtomicGroupID set the atomic group ID for an atomic group that already has Txs
func (ag *AtomicGroup) SetAtomicGroupID() {
	ids := []TxID{}
	for _, tx := range ag.Txs {
		ids = append(ids, tx.TxID)
	}
	ag.ID = CalculateAtomicGroupID(ids)
}

// IsAtomicGroupIDValid return false if the atomic group ID that is set
// doesn't match with the calculated
func (ag AtomicGroup) IsAtomicGroupIDValid() bool {
	ids := []TxID{}
	for _, tx := range ag.Txs {
		ids = append(ids, tx.TxID)
	}
	actualAGID := CalculateAtomicGroupID(ids)
	return actualAGID == ag.ID
}
