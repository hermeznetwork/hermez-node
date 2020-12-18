package common

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBJJFromStringWithChecksum(t *testing.T) {
	s := "21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7"
	pkComp, err := BJJFromStringWithChecksum(s)
	assert.NoError(t, err)
	sBytes, err := hex.DecodeString(s)
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(SwapEndianness(sBytes)), pkComp.String())

	pk, err := pkComp.Decompress()
	assert.NoError(t, err)

	// expected values computed with js implementation
	assert.Equal(t, "2492816973395423007340226948038371729989170225696553239457870892535792679622", pk.X.String())
	assert.Equal(t, "15238403086306505038849621710779816852318505119327426213168494964113886299863", pk.Y.String())
}
