package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type querier interface {
	Query(string) string
}

func parsePagination(c querier) (*uint, *bool, *uint, error) {
	// Offset
	offset := new(uint)
	*offset = 0
	offset, err := parseQueryUint("offset", offset, 0, maxUint32, c)
	if err != nil {
		return nil, nil, nil, err
	}
	// Last
	last := new(bool)
	*last = dfltLast
	last, err = parseQueryBool("last", last, c)
	if err != nil {
		return nil, nil, nil, err
	}
	if *last && (offset != nil && *offset > 0) {
		return nil, nil, nil, errors.New(
			"last and offset are incompatible, provide only one of them",
		)
	}
	// Limit
	limit := new(uint)
	*limit = dfltLimit
	limit, err = parseQueryUint("limit", limit, 1, maxLimit, c)
	if err != nil {
		return nil, nil, nil, err
	}
	return offset, last, limit, nil
}

func parseQueryUint(name string, dflt *uint, min, max uint, c querier) (*uint, error) { //nolint:SA4009 res may be not overwriten
	str := c.Query(name)
	if str != "" {
		resInt, err := strconv.Atoi(str)
		if err != nil || resInt < 0 || resInt < int(min) || resInt > int(max) {
			return nil, fmt.Errorf(
				"Inavlid %s. Must be an integer within the range [%d, %d]",
				name, min, max)
		}
		res := uint(resInt)
		return &res, nil
	}
	return dflt, nil
}

func parseQueryBool(name string, dflt *bool, c querier) (*bool, error) { //nolint:SA4009 res may be not overwriten
	str := c.Query(name)
	if str == "" {
		return dflt, nil
	}
	if str == "true" {
		res := new(bool)
		*res = true
		return res, nil
	}
	if str == "false" {
		res := new(bool)
		*res = false
		return res, nil
	}
	return nil, fmt.Errorf("Inavlid %s. Must be eithe true or false", name)
}

func parseQueryHezEthAddr(c querier) (*ethCommon.Address, error) {
	const name = "hermezEthereumAddress"
	addrStr := c.Query(name)
	if addrStr == "" {
		return nil, nil
	}
	splitted := strings.Split(addrStr, "hez:")
	if len(splitted) != 2 || len(splitted[1]) != 42 {
		return nil, fmt.Errorf(
			"Invalid %s, must follow this regex: ^hez:0x[a-fA-F0-9]{40}$", name)
	}
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(splitted[1]))
	return &addr, err
}

func parseQueryBJJ(c querier) (*babyjub.PublicKey, error) {
	const name = "BJJ"
	const decodedLen = 33
	bjjStr := c.Query(name)
	if bjjStr == "" {
		return nil, nil
	}
	splitted := strings.Split(bjjStr, "hez:")
	if len(splitted) != 2 || len(splitted[1]) != 44 {
		return nil, fmt.Errorf(
			"Invalid %s, must follow this regex: ^hez:[A-Za-z0-9+/=]{44}$",
			name)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(splitted[1])
	if err != nil {
		return nil, fmt.Errorf(
			"Invalid %s, error decoding base64 string: %s",
			name, err.Error())
	}
	if len(decoded) != decodedLen {
		return nil, fmt.Errorf(
			"invalid %s, error decoding base64 string: unexpected byte array length",
			name)
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return nil, fmt.Errorf("invalid %s, checksum failed",
			name)
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	bjj, err := bjjComp.Decompress()
	if err != nil {
		return nil, fmt.Errorf(
			"invalid %s, error decompressing public key: %s",
			name, err.Error())
	}
	return bjj, nil
}

func parseQueryTxType(c querier) (*common.TxType, error) {
	const name = "type"
	typeStr := c.Query(name)
	if typeStr == "" {
		return nil, nil
	}
	switch common.TxType(typeStr) {
	case common.TxTypeExit:
		ret := common.TxTypeExit
		return &ret, nil
	case common.TxTypeTransfer:
		ret := common.TxTypeTransfer
		return &ret, nil
	case common.TxTypeDeposit:
		ret := common.TxTypeDeposit
		return &ret, nil
	case common.TxTypeCreateAccountDeposit:
		ret := common.TxTypeCreateAccountDeposit
		return &ret, nil
	case common.TxTypeCreateAccountDepositTransfer:
		ret := common.TxTypeCreateAccountDepositTransfer
		return &ret, nil
	case common.TxTypeDepositTransfer:
		ret := common.TxTypeDepositTransfer
		return &ret, nil
	case common.TxTypeForceTransfer:
		ret := common.TxTypeForceTransfer
		return &ret, nil
	case common.TxTypeForceExit:
		ret := common.TxTypeForceExit
		return &ret, nil
	case common.TxTypeTransferToEthAddr:
		ret := common.TxTypeTransferToEthAddr
		return &ret, nil
	case common.TxTypeTransferToBJJ:
		ret := common.TxTypeTransferToBJJ
		return &ret, nil
	}
	return nil, fmt.Errorf(
		"invalid %s, %s is not a valid option. Check the valid options in the docmentation",
		name, typeStr,
	)
}

func parseIdx(c querier) (*uint, error) {
	const name = "accountIndex"
	addrStr := c.Query(name)
	if addrStr == "" {
		return nil, nil
	}
	splitted := strings.Split(addrStr, ":")
	const expectedLen = 3
	if len(splitted) != expectedLen {
		return nil, fmt.Errorf(
			"invalid %s, must follow this: hez:<tokenSymbol>:index", name)
	}
	idxInt, err := strconv.Atoi(splitted[2])
	idx := uint(idxInt)
	return &idx, err
}
