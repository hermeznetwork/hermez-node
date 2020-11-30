package common

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// AccountCreationAuth authorizations sent by users to the L2DB, to be used for account creations when necessary
type AccountCreationAuth struct {
	EthAddr   ethCommon.Address  `meddler:"eth_addr"`
	BJJ       *babyjub.PublicKey `meddler:"bjj"`
	Signature []byte             `meddler:"signature"`
	Timestamp time.Time          `meddler:"timestamp,utctime"`
}

// HashToSign builds the hash to be signed using BJJ pub key and the constant message
func (a *AccountCreationAuth) HashToSign() ([]byte, error) {
	// Calculate message to be signed
	const msg = "I authorize this babyjubjub key for hermez rollup account creation"
	comp, err := a.BJJ.Compress().MarshalText()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	// Hash message (msg || compressed-bjj)
	return ethCrypto.Keccak256Hash([]byte(msg), comp).Bytes(), nil
}

// VerifySignature ensures that the Signature is done with the specified EthAddr
func (a *AccountCreationAuth) VerifySignature() bool {
	// Calculate hash to be signed
	msg, err := a.HashToSign()
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
