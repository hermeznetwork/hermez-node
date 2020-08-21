package common

import "errors"

// ErrNotInFF is used when the *big.Int does not fit inside the Finite Field
var ErrNotInFF = errors.New("BigInt not inside the Finite Field")

// ErrNumOverflow is used when a given value overflows the maximum capacity of the parameter
var ErrNumOverflow = errors.New("Value overflows the type")

// ErrNonceOverflow is used when a given nonce overflows the maximum capacity of the Nonce (2**40-1)
var ErrNonceOverflow = errors.New("Nonce overflow, max value: 2**40 -1")

// ErrBatchQueueEmpty is used when the coordinator.BatchQueue.Pop() is called and has no elements
var ErrBatchQueueEmpty = errors.New("BatchQueue empty")
