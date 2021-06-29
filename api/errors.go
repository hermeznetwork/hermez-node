package api

const (
	// Public error messages (included in response objects)

	// ErrDuplicatedKey error message returned when trying to insert an item with duplicated key
	ErrDuplicatedKey = "Item already exists"
	// ErrSQLTimeout error message returned when timeout due to SQL connection
	ErrSQLTimeout = "The node is under heavy preasure, please try again later"
	// ErrExitAmount0 error message returned when receiving (and rejecting) a tx of type exit with amount 0
	ErrExitAmount0 = "Transaction rejected because an exit with amount 0 has no sense"
	// ErrTxsNotAtomic error message returned when receiving (and rejecting) txs in the atomic endpoint with not all txs being atomic
	ErrTxsNotAtomic = "There is at least one transaction in the payload that could be forged without the others"
	// ErrSingleTxInAtomicEndpoint only one tx sent to the atomic-pool endpoint
	ErrSingleTxInAtomicEndpoint = "To use the atomic-pool endpoint at least two transactions are required"

	// Internal error messages (used for logs or handling errors returned from internal comopnents)

	// errCtxTimeout error message received internally when context reaches timeout
	errCtxTimeout = "context deadline exceeded"

	// ErrInvalidSymbol error message returned when receiving (and rejecting) an invalid Symbol
	ErrInvalidSymbol = "Invalid Symbol"

	// ErrIsAtomic filter atomic transactions on POST /transactions-pool
	ErrIsAtomic = "Thies endpoint does not accept atomic transactions"
	// ErrInvalidRqOffset error message returned when received an invalid request offset
	ErrInvalidRqOffset = "Invalid requestOffset. Valid values goes from 0 to 7"

	// ErrRqOffsetOutOfBounds error message returned when transaction tries to access another out the bounds of the array
	ErrRqOffsetOutOfBounds = "One of the transactions requested another one outside the bounds of the provided array"

	// ErrNotAtomicTxsInPostPoolTx error message returned when received an non-atomic transaction inside atomic pool
	ErrNotAtomicTxsInPostPoolTx = "Atomic transactions are only accepted in POST /atomic-pool"

	// ErrInvalidAtomicGroupID error message returned when received an invalid AtomicGroupID
	ErrInvalidAtomicGroupID = "Invalid atomicGroupId"
)
