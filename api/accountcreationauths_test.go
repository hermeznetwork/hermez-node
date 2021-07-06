package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

type testAuth struct {
	EthAddr   string    `json:"hezEthereumAddress" binding:"required"`
	BJJ       string    `json:"bjj" binding:"required"`
	Signature string    `json:"signature" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

func genTestAuths(auths []*common.AccountCreationAuth) []testAuth {
	testAuths := []testAuth{}
	for _, auth := range auths {
		testAuths = append(testAuths, testAuth{
			EthAddr:   common.EthAddrToHez(auth.EthAddr),
			BJJ:       common.BjjToString(auth.BJJ),
			Signature: "0x" + hex.EncodeToString(auth.Signature),
			Timestamp: auth.Timestamp,
		})
	}
	return testAuths
}

func TestAccountCreationAuth(t *testing.T) {
	// POST
	endpoint := apiURL + "account-creation-authorization"
	for _, auth := range tc.auths {
		jsonAuthBytes, err := json.Marshal(auth)
		assert.NoError(t, err)
		jsonAuthReader := bytes.NewReader(jsonAuthBytes)
		assert.NoError(
			t, doGoodReq(
				"POST",
				endpoint,
				jsonAuthReader, nil,
			),
		)
	}
	// GET
	endpoint += "/"
	for _, auth := range tc.auths {
		fetchedAuth := testAuth{}
		assert.NoError(
			t, doGoodReq(
				"GET",
				endpoint+auth.EthAddr,
				nil, &fetchedAuth,
			),
		)
		assertAuth(t, auth, fetchedAuth)
	}
	// POST
	// 400
	// Wrong addr
	badAuth := tc.auths[0]
	badAuth.EthAddr = common.EthAddrToHez(ethCommon.BigToAddress(big.NewInt(1)))
	jsonAuthBytes, err := json.Marshal(badAuth)
	assert.NoError(t, err)
	jsonAuthReader := bytes.NewReader(jsonAuthBytes)
	err = doBadReq("POST", endpoint, jsonAuthReader, 400)
	assert.NoError(t, err)
	// Wrong signature
	badAuth = tc.auths[0]
	badAuth.Signature = badAuth.Signature[:len(badAuth.Signature)-1]
	badAuth.Signature += "F"
	jsonAuthBytes, err = json.Marshal(badAuth)
	assert.NoError(t, err)
	jsonAuthReader = bytes.NewReader(jsonAuthBytes)
	err = doBadReq("POST", endpoint, jsonAuthReader, 400)
	assert.NoError(t, err)
	// GET
	// 400
	err = doBadReq("GET", endpoint+"hez:0xFooBar", nil, 400)
	assert.NoError(t, err)
	// 404
	err = doBadReq("GET", endpoint+"hez:0x0000000000000000000000000000000000000001", nil, 404)
	assert.NoError(t, err)
}

func assertAuth(t *testing.T, expected, actual testAuth) {
	// timestamp should be very close to now
	assert.Less(t, time.Now().UTC().Unix()-3, actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp
	assert.Equal(t, expected, actual)
}
