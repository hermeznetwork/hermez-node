package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/ztrue/tracerr"
)

const exitIdx = "hez:EXIT:1"

// Query parsers

type querier interface {
	Query(string) string
}

func parsePagination(c querier) (fromItem *uint, order string, limit *uint, err error) {
	// FromItem
	fromItem, err = parseQueryUint("fromItem", nil, 0, maxUint32, c)
	if err != nil {
		return nil, "", nil, tracerr.Wrap(err)
	}
	// Order
	order = dfltOrder
	const orderName = "order"
	orderStr := c.Query(orderName)
	if orderStr != "" && !(orderStr == historydb.OrderAsc || historydb.OrderDesc == orderStr) {
		return nil, "", nil, tracerr.Wrap(errors.New(
			"order must have the value " + historydb.OrderAsc + " or " + historydb.OrderDesc,
		))
	}
	if orderStr == historydb.OrderAsc {
		order = historydb.OrderAsc
	} else if orderStr == historydb.OrderDesc {
		order = historydb.OrderDesc
	}
	// Limit
	limit = new(uint)
	*limit = dfltLimit
	limit, err = parseQueryUint("limit", limit, 1, maxLimit, c)
	if err != nil {
		return nil, "", nil, tracerr.Wrap(err)
	}
	return fromItem, order, limit, nil
}

// nolint reason: res may be not overwriten
func parseQueryUint(name string, dflt *uint, min, max uint, c querier) (*uint, error) { //nolint:SA4009
	str := c.Query(name)
	return stringToUint(str, name, dflt, min, max)
}

// nolint reason: res may be not overwriten
func parseQueryInt64(name string, dflt *int64, min, max int64, c querier) (*int64, error) { //nolint:SA4009
	str := c.Query(name)
	return stringToInt64(str, name, dflt, min, max)
}

// nolint reason: res may be not overwriten
func parseQueryBool(name string, dflt *bool, c querier) (*bool, error) { //nolint:SA4009
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
	return nil, tracerr.Wrap(fmt.Errorf("Invalid %s. Must be eithe true or false", name))
}

func parseQueryHezEthAddr(c querier) (*ethCommon.Address, error) {
	const name = "hermezEthereumAddress"
	addrStr := c.Query(name)
	return hezStringToEthAddr(addrStr, name)
}

func parseQueryBJJ(c querier) (*babyjub.PublicKey, error) {
	const name = "BJJ"
	bjjStr := c.Query(name)
	if bjjStr == "" {
		return nil, nil
	}
	return hezStringToBJJ(bjjStr, name)
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
	return nil, tracerr.Wrap(fmt.Errorf(
		"invalid %s, %s is not a valid option. Check the valid options in the docmentation",
		name, typeStr,
	))
}

func parseIdx(c querier) (*common.Idx, error) {
	const name = "accountIndex"
	idxStr := c.Query(name)
	return stringToIdx(idxStr, name)
}

func parseExitFilters(c querier) (*common.TokenID, *ethCommon.Address, *babyjub.PublicKey, *common.Idx, error) {
	// TokenID
	tid, err := parseQueryUint("tokenId", nil, 0, maxUint32, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	var tokenID *common.TokenID
	if tid != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*tid)
	}
	// Hez Eth addr
	addr, err := parseQueryHezEthAddr(c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	// BJJ
	bjj, err := parseQueryBJJ(c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	if addr != nil && bjj != nil {
		return nil, nil, nil, nil, tracerr.Wrap(errors.New("bjj and hermezEthereumAddress params are incompatible"))
	}
	// Idx
	idx, err := parseIdx(c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	if idx != nil && (addr != nil || bjj != nil || tokenID != nil) {
		return nil, nil, nil, nil, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hermezEthereumAddress and tokenId"))
	}
	return tokenID, addr, bjj, idx, nil
}

func parseTokenFilters(c querier) ([]common.TokenID, []string, string, error) {
	idsStr := c.Query("ids")
	symbolsStr := c.Query("symbols")
	nameStr := c.Query("name")
	var tokensIDs []common.TokenID
	if idsStr != "" {
		ids := strings.Split(idsStr, ",")

		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return nil, nil, "", tracerr.Wrap(err)
			}
			tokenID := common.TokenID(idUint)
			tokensIDs = append(tokensIDs, tokenID)
		}
	}
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
	}
	return tokensIDs, symbols, nameStr, nil
}

func parseBidFilters(c querier) (*int64, *ethCommon.Address, error) {
	slotNum, err := parseQueryInt64("slotNum", nil, 0, maxInt64, c)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	bidderAddr, err := parseQueryEthAddr("bidderAddr", c)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	return slotNum, bidderAddr, nil
}

func parseSlotFilters(c querier) (*int64, *int64, *ethCommon.Address, *bool, error) {
	minSlotNum, err := parseQueryInt64("minSlotNum", nil, 0, maxInt64, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	maxSlotNum, err := parseQueryInt64("maxSlotNum", nil, 0, maxInt64, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	wonByEthereumAddress, err := parseQueryEthAddr("wonByEthereumAddress", c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	finishedAuction, err := parseQueryBool("finishedAuction", nil, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	return minSlotNum, maxSlotNum, wonByEthereumAddress, finishedAuction, nil
}

func parseAccountFilters(c querier) ([]common.TokenID, *ethCommon.Address, *babyjub.PublicKey, error) {
	// TokenID
	idsStr := c.Query("tokenIds")
	var tokenIDs []common.TokenID
	if idsStr != "" {
		ids := strings.Split(idsStr, ",")

		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return nil, nil, nil, tracerr.Wrap(err)
			}
			tokenID := common.TokenID(idUint)
			tokenIDs = append(tokenIDs, tokenID)
		}
	}
	// Hez Eth addr
	addr, err := parseQueryHezEthAddr(c)
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	// BJJ
	bjj, err := parseQueryBJJ(c)
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if addr != nil && bjj != nil {
		return nil, nil, nil, tracerr.Wrap(errors.New("bjj and hermezEthereumAddress params are incompatible"))
	}

	return tokenIDs, addr, bjj, nil
}

// Param parsers

type paramer interface {
	Param(string) string
}

func parseParamTxID(c paramer) (common.TxID, error) {
	const name = "id"
	txIDStr := c.Param(name)
	if txIDStr == "" {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("%s is required", name))
	}
	txID, err := common.NewTxIDFromString(txIDStr)
	if err != nil {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("invalid %s", name))
	}
	return txID, nil
}

func parseParamIdx(c paramer) (*common.Idx, error) {
	const name = "accountIndex"
	idxStr := c.Param(name)
	return stringToIdx(idxStr, name)
}

// nolint reason: res may be not overwriten
func parseParamUint(name string, dflt *uint, min, max uint, c paramer) (*uint, error) { //nolint:SA4009
	str := c.Param(name)
	return stringToUint(str, name, dflt, min, max)
}

// nolint reason: res may be not overwriten
func parseParamInt64(name string, dflt *int64, min, max int64, c paramer) (*int64, error) { //nolint:SA4009
	str := c.Param(name)
	return stringToInt64(str, name, dflt, min, max)
}

func stringToIdx(idxStr, name string) (*common.Idx, error) {
	if idxStr == "" {
		return nil, nil
	}
	splitted := strings.Split(idxStr, ":")
	const expectedLen = 3
	if len(splitted) != expectedLen || splitted[0] != "hez" {
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, must follow this: hez:<tokenSymbol>:index", name))
	}
	// TODO: check that the tokenSymbol match the token related to the account index
	idxInt, err := strconv.Atoi(splitted[2])
	idx := common.Idx(idxInt)
	return &idx, tracerr.Wrap(err)
}

func stringToUint(uintStr, name string, dflt *uint, min, max uint) (*uint, error) {
	if uintStr != "" {
		resInt, err := strconv.Atoi(uintStr)
		if err != nil || resInt < 0 || resInt < int(min) || resInt > int(max) {
			return nil, tracerr.Wrap(fmt.Errorf(
				"Invalid %s. Must be an integer within the range [%d, %d]",
				name, min, max))
		}
		res := uint(resInt)
		return &res, nil
	}
	return dflt, nil
}

func stringToInt64(uintStr, name string, dflt *int64, min, max int64) (*int64, error) {
	if uintStr != "" {
		resInt, err := strconv.Atoi(uintStr)
		if err != nil || resInt < 0 || resInt < int(min) || resInt > int(max) {
			return nil, tracerr.Wrap(fmt.Errorf(
				"Invalid %s. Must be an integer within the range [%d, %d]",
				name, min, max))
		}
		res := int64(resInt)
		return &res, nil
	}
	return dflt, nil
}

func hezStringToEthAddr(addrStr, name string) (*ethCommon.Address, error) {
	if addrStr == "" {
		return nil, nil
	}
	splitted := strings.Split(addrStr, "hez:")
	if len(splitted) != 2 || len(splitted[1]) != 42 {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, must follow this regex: ^hez:0x[a-fA-F0-9]{40}$", name))
	}
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(splitted[1]))
	return &addr, tracerr.Wrap(err)
}

func hezStringToBJJ(bjjStr, name string) (*babyjub.PublicKey, error) {
	const decodedLen = 33
	splitted := strings.Split(bjjStr, "hez:")
	if len(splitted) != 2 || len(splitted[1]) != 44 {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, must follow this regex: ^hez:[A-Za-z0-9+/=]{44}$",
			name))
	}
	decoded, err := base64.RawURLEncoding.DecodeString(splitted[1])
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, error decoding base64 string: %s",
			name, err.Error()))
	}
	if len(decoded) != decodedLen {
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, error decoding base64 string: unexpected byte array length",
			name))
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return nil, tracerr.Wrap(fmt.Errorf("invalid %s, checksum failed",
			name))
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	bjj, err := bjjComp.Decompress()
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, error decompressing public key: %s",
			name, err.Error()))
	}
	return bjj, nil
}

func parseQueryEthAddr(name string, c querier) (*ethCommon.Address, error) {
	addrStr := c.Query(name)
	if addrStr == "" {
		return nil, nil
	}
	return parseEthAddr(addrStr)
}

func parseParamEthAddr(name string, c paramer) (*ethCommon.Address, error) {
	addrStr := c.Param(name)
	if addrStr == "" {
		return nil, nil
	}
	return parseEthAddr(addrStr)
}

func parseEthAddr(ethAddrStr string) (*ethCommon.Address, error) {
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(ethAddrStr))
	return &addr, tracerr.Wrap(err)
}

func parseParamHezEthAddr(c paramer) (*ethCommon.Address, error) {
	const name = "hermezEthereumAddress"
	addrStr := c.Param(name)
	return hezStringToEthAddr(addrStr, name)
}

type errorMsg struct {
	Message string
}

func bjjToString(bjj *babyjub.PublicKey) string {
	pkComp := [32]byte(bjj.Compress())
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

func ethAddrToHez(addr ethCommon.Address) string {
	return "hez:" + addr.String()
}

func idxToHez(idx common.Idx, tokenSymbol string) string {
	if idx == 1 {
		return exitIdx
	}
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}
