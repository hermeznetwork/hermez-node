package api

import (
	"encoding/base64"
	"math/big"
	"strconv"
	"strings"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

type queryParser struct {
	m map[string]string
}

func (qp *queryParser) Query(query string) string {
	if val, ok := qp.m[query]; ok {
		return val
	}
	return ""
}

func (qp *queryParser) Param(param string) string {
	if val, ok := qp.m[param]; ok {
		return val
	}
	return ""
}

func TestParseQueryUint(t *testing.T) {
	name := "foo"
	c := &queryParser{}
	c.m = make(map[string]string)
	var min uint = 1
	var max uint = 10
	var dflt *uint
	// Not uint
	c.m[name] = "-1"
	_, err := parseQueryUint(name, dflt, min, max, c)
	assert.Error(t, err)
	c.m[name] = "a"
	_, err = parseQueryUint(name, dflt, min, max, c)
	assert.Error(t, err)
	c.m[name] = "0.1"
	_, err = parseQueryUint(name, dflt, min, max, c)
	assert.Error(t, err)
	c.m[name] = "1.0"
	_, err = parseQueryUint(name, dflt, min, max, c)
	assert.Error(t, err)
	// Out of range
	c.m[name] = strconv.Itoa(int(min) - 1)
	_, err = parseQueryUint(name, dflt, min, max, c)
	assert.Error(t, err)
	c.m[name] = strconv.Itoa(int(max) + 1)
	_, err = parseQueryUint(name, dflt, min, max, c)
	assert.Error(t, err)
	// Default nil
	c.m[name] = ""
	res, err := parseQueryUint(name, dflt, min, max, c)
	assert.NoError(t, err)
	assert.Nil(t, res)
	// Default not nil
	dflt = new(uint)
	*dflt = uint(min)
	res, err = parseQueryUint(name, dflt, min, max, c)
	assert.NoError(t, err)
	assert.Equal(t, uint(min), *res)
	// Correct
	c.m[name] = strconv.Itoa(int(max))
	res, err = parseQueryUint(name, res, min, max, c)
	assert.NoError(t, err)
	assert.Equal(t, uint(max), *res)
}

func TestParseQueryBool(t *testing.T) {
	name := "foo"
	c := &queryParser{}
	c.m = make(map[string]string)
	var dflt *bool
	// Not bool
	c.m[name] = "x"
	_, err := parseQueryBool(name, dflt, c)
	assert.Error(t, err)
	c.m[name] = "False"
	_, err = parseQueryBool(name, dflt, c)
	assert.Error(t, err)
	c.m[name] = "0"
	_, err = parseQueryBool(name, dflt, c)
	assert.Error(t, err)
	c.m[name] = "1"
	_, err = parseQueryBool(name, dflt, c)
	assert.Error(t, err)
	// Default nil
	c.m[name] = ""
	res, err := parseQueryBool(name, dflt, c)
	assert.NoError(t, err)
	assert.Nil(t, res)
	// Default not nil
	dflt = new(bool)
	*dflt = true
	res, err = parseQueryBool(name, dflt, c)
	assert.NoError(t, err)
	assert.True(t, *res)
	// Correct
	c.m[name] = "false"
	res, err = parseQueryBool(name, dflt, c)
	assert.NoError(t, err)
	assert.False(t, *res)
	c.m[name] = "true"
	res, err = parseQueryBool(name, dflt, c)
	assert.NoError(t, err)
	assert.True(t, *res)
}

func TestParsePagination(t *testing.T) {
	c := &queryParser{}
	c.m = make(map[string]string)
	// fromItem out of range
	c.m["fromItem"] = "-1"
	_, _, _, err := parsePagination(c)
	assert.Error(t, err)
	c.m["fromItem"] = strconv.Itoa(maxUint32 + 1)
	_, _, _, err = parsePagination(c)
	assert.Error(t, err)
	c.m["fromItem"] = ""
	// Bad order
	c.m["order"] = "0"
	_, _, _, err = parsePagination(c)
	assert.Error(t, err)
	c.m["order"] = strconv.Itoa(int(maxLimit) + 1)
	_, _, _, err = parsePagination(c)
	assert.Error(t, err)
	// Default
	c.m["fromItem"] = ""
	c.m["order"] = ""
	c.m["limit"] = ""
	fromItem, order, limit, err := parsePagination(c)
	assert.NoError(t, err)
	assert.Nil(t, fromItem)
	assert.Equal(t, dfltOrder, order)
	assert.Equal(t, dfltLimit, *limit)
	// Correct
	c.m["fromItem"] = ""
	c.m["order"] = "ASC"
	c.m["limit"] = "25"
	fromItem, order, limit, err = parsePagination(c)
	assert.NoError(t, err)
	assert.Nil(t, fromItem)
	assert.Equal(t, "ASC", order)
	assert.Equal(t, 25, int(*limit))
	c.m["fromItem"] = "25"
	c.m["order"] = "DESC"
	c.m["limit"] = "50"
	fromItem, order, limit, err = parsePagination(c)
	assert.NoError(t, err)
	assert.Equal(t, 25, int(*fromItem))
	assert.Equal(t, "DESC", order)
	assert.Equal(t, 50, int(*limit))
}

func TestParseQueryHezEthAddr(t *testing.T) {
	name := "hezEthereumAddress"
	c := &queryParser{}
	c.m = make(map[string]string)
	ethAddr := ethCommon.BigToAddress(big.NewInt(int64(347683)))
	// Not HEZ Eth addr
	c.m[name] = "hez:0xf"
	_, err := parseQueryHezEthAddr(c)
	assert.Error(t, err)
	c.m[name] = ethAddr.String()
	_, err = parseQueryHezEthAddr(c)
	assert.Error(t, err)
	c.m[name] = "hez:0xXX942cfcd25ad4d90a62358b0dd84f33b39826XX"
	_, err = parseQueryHezEthAddr(c)
	assert.Error(t, err)
	// Default
	c.m[name] = ""
	res, err := parseQueryHezEthAddr(c)
	assert.NoError(t, err)
	assert.Nil(t, res)
	// Correct
	c.m[name] = "hez:" + ethAddr.String()
	res, err = parseQueryHezEthAddr(c)
	assert.NoError(t, err)
	assert.Equal(t, ethAddr, *res)
}

func TestParseQueryBJJ(t *testing.T) {
	name := "BJJ"
	c := &queryParser{}
	c.m = make(map[string]string)
	privK := babyjub.NewRandPrivKey()
	pubK := privK.Public()
	pkComp := [32]byte(pubK.Compress())
	// Not HEZ Eth addr
	c.m[name] = "hez:abcd"
	_, err := parseQueryBJJ(c)
	assert.Error(t, err)
	c.m[name] = pubK.String()
	_, err = parseQueryBJJ(c)
	assert.Error(t, err)
	// Wrong checksum
	bjjSum := append(pkComp[:], byte(1))
	c.m[name] = "hez:" + base64.RawStdEncoding.EncodeToString(bjjSum)
	_, err = parseQueryBJJ(c)
	assert.Error(t, err)
	// Default
	c.m[name] = ""
	res, err := parseQueryBJJ(c)
	assert.NoError(t, err)
	assert.Nil(t, res)
	// Correct
	c.m[name] = bjjToString(pubK.Compress())
	res, err = parseQueryBJJ(c)
	assert.NoError(t, err)
	assert.Equal(t, pubK.Compress(), *res)
}

func TestParseQueryTxType(t *testing.T) {
	name := "type"
	c := &queryParser{}
	c.m = make(map[string]string)
	// Incorrect values
	c.m[name] = "deposit"
	_, err := parseQueryTxType(c)
	assert.Error(t, err)
	c.m[name] = "1"
	_, err = parseQueryTxType(c)
	assert.Error(t, err)
	// Default
	c.m[name] = ""
	res, err := parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Nil(t, res)
	// Correct values
	c.m[name] = string(common.TxTypeExit)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeExit, *res)
	c.m[name] = string(common.TxTypeTransfer)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeTransfer, *res)
	c.m[name] = string(common.TxTypeDeposit)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeDeposit, *res)
	c.m[name] = string(common.TxTypeCreateAccountDeposit)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, *res)
	c.m[name] = string(common.TxTypeCreateAccountDepositTransfer)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeCreateAccountDepositTransfer, *res)
	c.m[name] = string(common.TxTypeDepositTransfer)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeDepositTransfer, *res)
	c.m[name] = string(common.TxTypeForceTransfer)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeForceTransfer, *res)
	c.m[name] = string(common.TxTypeForceExit)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeForceExit, *res)
	c.m[name] = string(common.TxTypeTransferToEthAddr)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeTransferToEthAddr, *res)
	c.m[name] = string(common.TxTypeTransferToBJJ)
	res, err = parseQueryTxType(c)
	assert.NoError(t, err)
	assert.Equal(t, common.TxTypeTransferToBJJ, *res)
}

func TestParseTokenFilters(t *testing.T) {
	ids := "ids"
	symbols := "symbols"
	name := "name"
	nameValue := "1"
	symbolsValue := "1,2,3"
	idsValue := "2,3,4"
	c := &queryParser{}
	c.m = make(map[string]string)
	// Incorrect values
	c.m[name] = nameValue
	c.m[ids] = idsValue
	c.m[symbols] = symbolsValue

	idsParse, symbolsParse, nameParse, err := parseTokenFilters(c)
	assert.NoError(t, err)

	// Correct values
	var tokenIds []common.TokenID = []common.TokenID{2, 3, 4}
	assert.Equal(t, tokenIds, idsParse)
	var symbolsArray []string = []string{"1", "2", "3"}
	assert.Equal(t, symbolsArray, symbolsParse)
	assert.Equal(t, nameValue, nameParse)
}
func TestParseEthAddr(t *testing.T) {
	name := "forgerAddr"
	c := &queryParser{}
	c.m = make(map[string]string)
	ethAddr := ethCommon.BigToAddress(big.NewInt(int64(123456)))
	// Default
	c.m[name] = ""
	res, err := parseQueryEthAddr(name, c)
	assert.NoError(t, err)
	assert.Nil(t, res)
	// Incorrect
	c.m[name] = "0x12345678"
	_, err = parseQueryEthAddr(name, c)
	assert.Error(t, err)
	// Correct
	c.m[name] = ethAddr.String()
	res, err = parseQueryEthAddr(name, c)
	assert.NoError(t, err)
	assert.Equal(t, ethAddr, *res)
}

func TestParseBidFilters(t *testing.T) {
	slotNum := "slotNum"
	bidderAddr := "bidderAddr"
	slotNumValue := "2"
	bidderAddrValue := "0xaa942cfcd25ad4d90a62358b0dd84f33b398262a"
	c := &queryParser{}
	c.m = make(map[string]string)
	// Incorrect values
	c.m[slotNum] = slotNumValue
	c.m[bidderAddr] = bidderAddrValue

	slotNumParse, bidderAddrParse, err := parseBidFilters(c)
	assert.NoError(t, err)

	// Correct values
	assert.Equal(t, strings.ToLower(bidderAddrParse.Hex()), bidderAddrValue)
	assert.Equal(t, slotNumValue, strconv.FormatUint(uint64(*slotNumParse), 10))
}
