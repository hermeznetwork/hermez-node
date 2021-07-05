package api

const (
	// Public error messages (included in response objects)

	// ErrDuplicatedKey error message returned when trying to insert an item with duplicated key
	ErrDuplicatedKey = "Item already exists"
	// ErrSQLTimeout error message returned when timeout due to SQL connection
	ErrSQLTimeout = "The node is under heavy preasure, please try again later"
	// ErrExitAmount0 error message returned when receiving (and rejecting) a tr of type exit with amount 0
	ErrExitAmount0 = "Transaction rejected because an exit with amount 0 has no sense"

	// Internal error messages (used for logs or handling errors returned from internal comopnents)

	// errCtxTimeout error message received internally when context reaches timeout
	errCtxTimeout = "context deadline exceeded"

	// ErrInvalidSymbol error message returned when receiving (and rejecting) an invalid Symbol
	ErrInvalidSymbol = "Invalid Symbol"

	// ErrIsAtomic filter atomic transactions on POST /transactions-pool
	ErrIsAtomic = "Thies endpoint does not accept atomic transactions"
)
