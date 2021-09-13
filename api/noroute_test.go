package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoRouteVersionNotProvided(t *testing.T) {
	endpoint := apiIP + apiPort + "/"
	// not using doGoodReq, bcs internally
	// there is a method FindRoute that checks route and returns error
	resp, err := doSimpleReq("GET", endpoint)
	assert.NoError(t, err)
	assert.Equal(t,
		"{\"error\":\"Version not provided, please provide a valid version in the path such as v1\"}",
		resp)
}

func TestNoRoute(t *testing.T) {
	endpoint := apiURL
	// not using doGoodReq, bcs internally
	// there is a method FindRoute that checks route and returns error
	resp, err := doSimpleReq("GET", endpoint)
	assert.NoError(t, err)
	assert.Equal(t,
		"{\"error\":\"404 page not found\"}",
		resp)
}
