package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type healthResponse struct {
	HistoryDB struct {
		LastMigration string `json:"last_migration"`
		Status        string `json:"status"`
		Version       string `json:"version"`
	} `json:"historyDB"`
	L2DB struct {
		LastMigration string `json:"last_migration"`
		Status        string `json:"status"`
		Version       string `json:"version"`
	} `json:"l2DB"`
	Status                   string    `json:"status"`
	Timestamp                time.Time `json:"timestamp"`
	Version                  string    `json:"version"`
	CoordinatorForgerBalance string    `json:"coordinatorForgerBalance"`
}

func TestHealth(t *testing.T) {
	endpoint := apiURL + "health"
	var healthResponseTest healthResponse
	err := doGoodReq("GET", endpoint, nil, &healthResponseTest)

	assert.NoError(t, err)
	assert.NotNil(t, healthResponseTest.L2DB.LastMigration)
	assert.Equal(t, healthResponseTest.L2DB.Status, "UP")
	assert.NotNil(t, healthResponseTest.HistoryDB.LastMigration)
	assert.Equal(t, healthResponseTest.HistoryDB.Status, "UP")
	assert.Equal(t, healthResponseTest.Version, "test")
	assert.Equal(t, healthResponseTest.Status, "UP")
}
