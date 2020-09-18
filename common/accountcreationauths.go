package common

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// AccountCreationAuth authorizations sent by users to the L2DB, to be used for account creations when necessary
type AccountCreationAuth struct {
	EthAddr   ethCommon.Address  `meddler:"eth_addr"`
	BJJ       *babyjub.PublicKey `meddler:"bjj"`
	Signature []byte             `meddler:"signature"`
	Timestamp time.Time          `meddler:"timestamp,utctime"`
}
