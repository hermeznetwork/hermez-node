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
	// ErrRqTxIDNotProvided error message returned when receiving (and rejecting) a tx that has malformed RqTxID
	ErrRqTxIDNotProvided = "Transaction requestId is not set or is invalid"

	// Internal error messages (used for logs or handling errors returned from internal comopnents)

	// errCtxTimeout error message received internally when context reaches timeout
	errCtxTimeout = "context deadline exceeded"
)
