package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type foo struct {
	V int
}

func TestSliceToSlicePtrs(t *testing.T) {
	n := 16
	a := make([]foo, n)
	for i := 0; i < n; i++ {
		a[i] = foo{V: i}
	}
	b := SliceToSlicePtrs(a).([]*foo)
	for i := 0; i < len(a); i++ {
		assert.Equal(t, a[i], *b[i])
	}
}

func TestSlicePtrsToSlice(t *testing.T) {
	n := 16
	a := make([]*foo, n)
	for i := 0; i < n; i++ {
		a[i] = &foo{V: i}
	}
	b := SlicePtrsToSlice(a).([]foo)
	for i := 0; i < len(a); i++ {
		assert.Equal(t, *a[i], b[i])
	}
}
