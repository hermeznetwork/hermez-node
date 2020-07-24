package txselector

import (
	"encoding/binary"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

func getAccountID(addr ethCommon.Address, tokenID uint32) [36]byte {
	var tokenIDBytes [4]byte
	binary.LittleEndian.PutUint32(tokenIDBytes[:], tokenID)
	accountIDBytes := append(addr[:], tokenIDBytes[:]...)
	var accountID [36]byte
	copy(accountID[:], accountIDBytes[:36])
	return accountID
}
