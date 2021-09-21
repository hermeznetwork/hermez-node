package apitypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshall(t *testing.T) {
	var l2tx TxL2
	_, err := l2tx.MarshalJSON()
	assert.NoError(t, err)
}

func TestUnmarshall(t *testing.T) {
	var l2tx, dest TxL2
	bytes, err := l2tx.MarshalJSON()
	assert.NoError(t, err)
	err = dest.UnmarshalJSON(bytes)
	assert.NoError(t, err)
}
