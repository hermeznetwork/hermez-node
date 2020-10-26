package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenAuths(t *testing.T) {
	const nAuths = 5
	auths := GenAuths(nAuths)
	for _, auth := range auths {
		assert.True(t, auth.VerifySignature())
	}
}
