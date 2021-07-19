package test

import (
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenAuths(t *testing.T) {
	chainID := uint16(0)
	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")

	const nAuths = 5
	auths := GenAuths(nAuths, chainID, hermezContractAddr)
	for _, auth := range auths {
		isValid, err := auth.VerifySignature(chainID, hermezContractAddr)
		require.NoError(t, err)
		assert.True(t, isValid)
	}
}
