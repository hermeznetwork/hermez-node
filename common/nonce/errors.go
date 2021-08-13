package nonce

import "errors"

// ErrNonceOverflow is used when a given nonce overflows the maximum capacity of the Nonce (2**40-1)
var ErrNonceOverflow = errors.New("Nonce overflow, max value: 2**40 -1")
