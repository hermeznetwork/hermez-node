package common

import (
	"encoding/binary"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// AccountCreationAuthMsg is the message that is signed to authorize an account
// creation
const AccountCreationAuthMsg = "I authorize this babyjubjub key for hermez rollup account creation"

// AccountCreationAuth authorizations sent by users to the L2DB, to be used for
// account creations when necessary
type AccountCreationAuth struct {
	EthAddr   ethCommon.Address     `meddler:"eth_addr"`
	BJJ       babyjub.PublicKeyComp `meddler:"bjj"`
	Signature []byte                `meddler:"signature"`
	Timestamp time.Time             `meddler:"timestamp,utctime"`
}

// HashToSign builds the hash to be signed using BJJ pub key and the constant message
func (a *AccountCreationAuth) HashToSign(chainID uint16,
	hermezContractAddr ethCommon.Address) ([]byte, error) {
	// Calculate message to be signed
	var chainIDBytes [2]byte
	binary.BigEndian.PutUint16(chainIDBytes[:], chainID)
	// to hash: [AccountCreationAuthMsg | compressedBJJ | chainID | hermezContractAddr]
	return ethCrypto.Keccak256Hash([]byte(AccountCreationAuthMsg), a.BJJ[:], chainIDBytes[:],
		hermezContractAddr[:]).Bytes(), nil
}

// VerifySignature ensures that the Signature is done with the specified EthAddr
func (a *AccountCreationAuth) VerifySignature(chainID uint16,
	hermezContractAddr ethCommon.Address) bool {
	// Calculate hash to be signed
	msg, err := a.HashToSign(chainID, hermezContractAddr)
	if err != nil {
		return false
	}
	// Get public key from Signature
	pubKBytes, err := ethCrypto.Ecrecover(msg, a.Signature)
	if err != nil {
		return false
	}
	pubK, err := ethCrypto.UnmarshalPubkey(pubKBytes)
	if err != nil {
		return false
	}
	// Get addr from pubK
	addr := ethCrypto.PubkeyToAddress(*pubK)
	return addr == a.EthAddr
}
