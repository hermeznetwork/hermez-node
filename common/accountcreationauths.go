package common

import (
	"encoding/hex"
	"fmt"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethMath "github.com/ethereum/go-ethereum/common/math"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	ethSigner "github.com/ethereum/go-ethereum/signer/core"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// AccountCreationAuthMsg is the message that is signed to authorize a
	// Hermez account creation
	AccountCreationAuthMsg = "Account creation"
	// EIP712Version is the used version of the EIP-712
	EIP712Version = "1"
	// EIP712Provider defines the Provider for the EIP-712
	EIP712Provider = "Hermez Network"
)

var (
	// EmptyEthSignature is an ethereum signature of all zeroes
	EmptyEthSignature = make([]byte, 65)
)

// AccountCreationAuth authorizations sent by users to the L2DB, to be used for
// account creations when necessary
type AccountCreationAuth struct {
	EthAddr   ethCommon.Address     `meddler:"eth_addr"`
	BJJ       babyjub.PublicKeyComp `meddler:"bjj"`
	Signature []byte                `meddler:"signature"`
	Timestamp time.Time             `meddler:"timestamp,utctime"`
}

// toHash returns a byte array to be hashed from the AccountCreationAuth, which
// follows the EIP-712 encoding
func (a *AccountCreationAuth) toHash(chainID uint16,
	hermezContractAddr ethCommon.Address) ([]byte, error) {
	chainIDFormatted := ethMath.NewHexOrDecimal256(int64(chainID))

	signerData := ethSigner.TypedData{
		Types: ethSigner.Types{
			"EIP712Domain": []ethSigner.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Authorise": []ethSigner.Type{
				{Name: "Provider", Type: "string"},
				{Name: "Authorisation", Type: "string"},
				{Name: "BJJKey", Type: "bytes32"},
			},
		},
		PrimaryType: "Authorise",
		Domain: ethSigner.TypedDataDomain{
			Name:              EIP712Provider,
			Version:           EIP712Version,
			ChainId:           chainIDFormatted,
			VerifyingContract: hermezContractAddr.Hex(),
		},
		Message: ethSigner.TypedDataMessage{
			"Provider":      EIP712Provider,
			"Authorisation": AccountCreationAuthMsg,
			"BJJKey":        SwapEndianness(a.BJJ[:]),
		},
	}

	domainSeparator, err := signerData.HashStruct("EIP712Domain", signerData.Domain.Map())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	typedDataHash, err := signerData.HashStruct(signerData.PrimaryType, signerData.Message)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	rawData := []byte{0x19, 0x01} // "\x19\x01"
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, typedDataHash...)
	return rawData, nil
}

// HashToSign returns the hash to be signed by the Ethereum address to authorize
// the account creation, which follows the EIP-712 encoding
func (a *AccountCreationAuth) HashToSign(chainID uint16,
	hermezContractAddr ethCommon.Address) ([]byte, error) {
	b, err := a.toHash(chainID, hermezContractAddr)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return ethCrypto.Keccak256(b), nil
}

// Sign signs the account creation authorization message using the provided
// `signHash` function, and stores the signature in `a.Signature`.  `signHash`
// should do an ethereum signature using the account corresponding to
// `a.EthAddr`.  The `signHash` function is used to make signing flexible: in
// tests we sign directly using the private key, outside tests we sign using
// the keystore (which never exposes the private key). Sign follows the EIP-712
// encoding.
func (a *AccountCreationAuth) Sign(signHash func(hash []byte) ([]byte, error),
	chainID uint16, hermezContractAddr ethCommon.Address) error {
	hash, err := a.HashToSign(chainID, hermezContractAddr)
	if err != nil {
		return tracerr.Wrap(err)
	}
	sig, err := signHash(hash)
	if err != nil {
		return tracerr.Wrap(err)
	}
	sig[64] += 27
	a.Signature = sig
	a.Timestamp = time.Now()
	return nil
}

// VerifySignature ensures that the Signature is done with the EthAddr, for the
// chainID and hermezContractAddress passed by parameter. VerifySignature
// follows the EIP-712 encoding.
func (a *AccountCreationAuth) VerifySignature(chainID uint16,
	hermezContractAddr ethCommon.Address) (bool, error) {
	// Calculate hash to be signed
	hash, err := a.HashToSign(chainID, hermezContractAddr)
	if err != nil {
		signatureV := hex.EncodeToString(a.Signature)
		return false, fmt.Errorf("error calculating hash to be signed: %s. "+
			"ChainId: %d. HermezContractAddress: %s. EthereumAddress: %s. Bjj: %s. "+
			"Signature: %s. Timestamp: %s", err.Error(), chainID, hermezContractAddr.String(),
			a.EthAddr.String(), BjjToString(a.BJJ), signatureV, a.Timestamp.String())
	}

	var sig [65]byte
	copy(sig[:], a.Signature[:])
	sig[64] -= 27

	// Get public key from Signature
	pubKBytes, err := ethCrypto.Ecrecover(hash, sig[:])
	if err != nil {
		signatureV := hex.EncodeToString(a.Signature)
		return false, fmt.Errorf("error getting public key from Signature: %s. "+
			"ChainId: %d. HermezContractAddress: %s. EthereumAddress: %s. Bjj: %s. "+
			"Signature: %s. Timestamp: %s", err.Error(), chainID, hermezContractAddr.String(),
			a.EthAddr.String(), BjjToString(a.BJJ), signatureV, a.Timestamp.String())
	}
	pubK, err := ethCrypto.UnmarshalPubkey(pubKBytes)
	if err != nil {
		signatureV := hex.EncodeToString(a.Signature)
		return false, fmt.Errorf("error unmarshalling public key: %s. ChainId: %d. "+
			"HermezContractAddress: %s. EthereumAddress: %s. Bjj: %s. Signature: %s. "+
			"Timestamp: %s", err.Error(), chainID, hermezContractAddr.String(),
			a.EthAddr.String(), BjjToString(a.BJJ), signatureV, a.Timestamp.String())
	}
	// Get addr from pubK
	addr := ethCrypto.PubkeyToAddress(*pubK)
	if addr != a.EthAddr {
		signatureV := hex.EncodeToString(a.Signature)
		return false, fmt.Errorf("error: Ethereum address doesn't match with the one used in the signature. "+
			"ChainId: %d. HermezContractAddress: %s. EthereumAddress: %s. Bjj: %s. Signature: %s. Timestamp: %s",
			chainID, hermezContractAddr.String(), a.EthAddr.String(), BjjToString(a.BJJ), signatureV, a.Timestamp.String())
	}
	return true, nil
}
