package common

import (
	"encoding/binary"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// AccountCreationAuthMsg is the message that is signed to authorize a Hermez
// account creation
const AccountCreationAuthMsg = "I authorize this babyjubjub key for hermez rollup account creation"

// EthMsgPrefix is the prefix for message signing at the Ethereum ecosystem
const EthMsgPrefix = "\x19Ethereum Signed Message:\n"

// AccountCreationAuth authorizations sent by users to the L2DB, to be used for
// account creations when necessary
type AccountCreationAuth struct {
	EthAddr   ethCommon.Address     `meddler:"eth_addr"`
	BJJ       babyjub.PublicKeyComp `meddler:"bjj"`
	Signature []byte                `meddler:"signature"`
	Timestamp time.Time             `meddler:"timestamp,utctime"`
}

func (a *AccountCreationAuth) toHash(chainID uint16,
	hermezContractAddr ethCommon.Address) []byte {
	var chainIDBytes [2]byte
	binary.BigEndian.PutUint16(chainIDBytes[:], chainID)
	// [EthPrefix | AccountCreationAuthMsg | compressedBJJ | chainID | hermezContractAddr]
	var b []byte
	b = append(b, []byte(AccountCreationAuthMsg)...)
	b = append(b, SwapEndianness(a.BJJ[:])...) // for js implementation compatibility
	b = append(b, chainIDBytes[:]...)
	b = append(b, hermezContractAddr[:]...)

	ethPrefix := EthMsgPrefix + strconv.Itoa(len(b))
	return append([]byte(ethPrefix), b...)
}

// HashToSign returns the hash to be signed by the Etherum address to authorize
// the account creation
func (a *AccountCreationAuth) HashToSign(chainID uint16,
	hermezContractAddr ethCommon.Address) ([]byte, error) {
	b := a.toHash(chainID, hermezContractAddr)
	return ethCrypto.Keccak256Hash(b).Bytes(), nil
}

// Sign signs the account creation authorization message using the provided
// `signHash` function, and stores the signaure in `a.Signature`.  `signHash`
// should do an ethereum signature using the account corresponding to
// `a.EthAddr`.  The `signHash` function is used to make signig flexible: in
// tests we sign directly using the private key, outside tests we sign using
// the keystore (which never exposes the private key).
func (a *AccountCreationAuth) Sign(signHash func(hash []byte) ([]byte, error),
	chainID uint16, hermezContractAddr ethCommon.Address) error {
	hash, err := a.HashToSign(chainID, hermezContractAddr)
	if err != nil {
		return err
	}
	sig, err := signHash(hash)
	if err != nil {
		return err
	}
	sig[64] += 27
	a.Signature = sig
	a.Timestamp = time.Now()
	return nil
}

// VerifySignature ensures that the Signature is done with the EthAddr, for the
// chainID and hermezContractAddress passed by parameter
func (a *AccountCreationAuth) VerifySignature(chainID uint16,
	hermezContractAddr ethCommon.Address) bool {
	// Calculate hash to be signed
	hash, err := a.HashToSign(chainID, hermezContractAddr)
	if err != nil {
		return false
	}

	var sig [65]byte
	copy(sig[:], a.Signature[:])
	sig[64] -= 27

	// Get public key from Signature
	pubKBytes, err := ethCrypto.Ecrecover(hash, sig[:])
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
