package common

import (
	"encoding/hex"
	"testing"

	"github.com/iden3/go-merkletree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t,
		"2492816973395423007340226948038371729989170225696553239457870892535792679622",
		pk.X.String())
	assert.Equal(t,
		"15238403086306505038849621710779816852318505119327426213168494964113886299863",
		pk.Y.String())
}

func TestRmEndingZeroes(t *testing.T) {
	s0, err :=
		merkletree.NewHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	s1, err :=
		merkletree.NewHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000001")
	require.NoError(t, err)
	s2, err :=
		merkletree.NewHashFromHex("0x0000000000000000000000000000000000000000000000000000000000000002")
	require.NoError(t, err)

	// expect cropped last zeroes
	circomSiblings := []*merkletree.Hash{s0, s1, s0, s1, s1, s1, s2, s0, s0, s0, s0}
	siblings := RmEndingZeroes(circomSiblings)
	expected := []*merkletree.Hash{s0, s1, s0, s1, s1, s1, s2}
	assert.Equal(t, expected, siblings)

	// expect empty array when input is an empty array
	siblings = RmEndingZeroes([]*merkletree.Hash{})
	assert.Equal(t, []*merkletree.Hash{}, siblings)
	// expect nil when input is nil
	siblings = RmEndingZeroes(nil)
	assert.Nil(t, siblings)

	// cases when inputs are [x], [x,0], [0,x]
	circomSiblings = []*merkletree.Hash{s1}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, []*merkletree.Hash{s1}, siblings)
	circomSiblings = []*merkletree.Hash{s1, s0}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, []*merkletree.Hash{s1}, siblings)
	circomSiblings = []*merkletree.Hash{s0, s1}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, []*merkletree.Hash{s0, s1}, siblings)

	// expect empty array when input is all zeroes
	circomSiblings = []*merkletree.Hash{s0}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, []*merkletree.Hash{}, siblings)
	circomSiblings = []*merkletree.Hash{s0, s0}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, []*merkletree.Hash{}, siblings)
	circomSiblings = []*merkletree.Hash{s0, s0, s0, s0, s0}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, []*merkletree.Hash{}, siblings)

	// expect input equal to output when last element!=0
	circomSiblings = []*merkletree.Hash{s0, s1, s0, s1, s1, s1, s2, s0, s0, s0, s0, s2}
	siblings = RmEndingZeroes(circomSiblings)
	assert.Equal(t, circomSiblings, siblings)
}
