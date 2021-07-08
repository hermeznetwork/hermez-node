package api

type apiErrorCode uint
type apiErrorType string

const (
	// Public error messages (included in response objects)

	// ErrParamValidationFailedCode code for param validation failed error
	ErrParamValidationFailedCode apiErrorCode = 1
	// ErrParamValidationFailedType type for param validation failed error
	ErrParamValidationFailedType apiErrorType = "ErrParamValidationFailed"

	// ErrDuplicatedKey error message returned when trying to insert an item with duplicated key
	ErrDuplicatedKey = "Item already exists"
	// ErrDuplicatedKeyCode code for duplicated key error
	ErrDuplicatedKeyCode apiErrorCode = 2
	// ErrDuplicatedKeyType type for duplicated key error
	ErrDuplicatedKeyType apiErrorType = "ErrDuplicatedKey"

	// ErrSQLTimeout error message returned when timeout due to SQL connection
	ErrSQLTimeout = "The node is under heavy preasure, please try again later"
	// ErrSQLTimeoutCode code for sql timeout error
	ErrSQLTimeoutCode apiErrorCode = 3
	// ErrSQLTimeoutType type for sql timeout type
	ErrSQLTimeoutType apiErrorType = "ErrSQLTimeout"

	// ErrSQLNoRowsCode code for no rows error
	ErrSQLNoRowsCode apiErrorCode = 4
	// ErrSQLNoRowsType type for now rows error
	ErrSQLNoRowsType apiErrorType = "ErrSQLNoRows"

	// ErrExitAmount0 error message returned when receiving (and rejecting) a tr of type exit with amount 0
	ErrExitAmount0 = "Transaction rejected because an exit with amount 0 has no sense"
	// ErrExitAmount0Code code for 0 exit amount error
	ErrExitAmount0Code apiErrorCode = 5
	// ErrExitAmount0Type type for 0 exit amount error
	ErrExitAmount0Type apiErrorType = "ErrExitAmount0"

	// ErrInvalidTxTypeOrTxIDCode code for invalid tx type or txID error
	ErrInvalidTxTypeOrTxIDCode apiErrorCode = 6
	// ErrInvalidTxTypeOrTxIDType type for invalid tx type or txID error
	ErrInvalidTxTypeOrTxIDType apiErrorType = "ErrInvalidTxTypeOrTxID"

	// ErrFeeOverflowCode code for fee overflow code error
	ErrFeeOverflowCode apiErrorCode = 7
	// ErrFeeOverflowType type for fee overflow code type
	ErrFeeOverflowType apiErrorType = "ErrFeeOverflow"

	// ErrGettingSenderAccountCode code for getting sender account error
	ErrGettingSenderAccountCode apiErrorCode = 8
	// ErrGettingSenderAccountType type for getting sender account error
	ErrGettingSenderAccountType apiErrorType = "ErrGettingSenderAccount"

	// ErrAccountTokenNotEqualTxTokenCode code for account token not equal tx token error
	ErrAccountTokenNotEqualTxTokenCode apiErrorCode = 9
	// ErrAccountTokenNotEqualTxTokenType type for account token not equal tx token type
	ErrAccountTokenNotEqualTxTokenType apiErrorType = "ErrAccountTokenNotEqualTxToken"

	// ErrInvalidNonceCode code for invalid nonce error
	ErrInvalidNonceCode apiErrorCode = 10
	// ErrInvalidNonceType type for invalid nonce error
	ErrInvalidNonceType apiErrorType = "ErrInvalidNonce"

	// ErrInvalidSignatureCode code for invalid signature error
	ErrInvalidSignatureCode apiErrorCode = 11
	// ErrInvalidSignatureType type for invalid signature error
	ErrInvalidSignatureType apiErrorType = "ErrInvalidSignature"

	// ErrGettingReceiverAccountCode code for getting receiver account error
	ErrGettingReceiverAccountCode apiErrorCode = 12
	// ErrGettingReceiverAccountType type for getting receiver account error
	ErrGettingReceiverAccountType apiErrorType = "ErrGettingReceiverAccount"

	// ErrCantSendToEthAddrCode code when can't send to eth addr code error appeared
	ErrCantSendToEthAddrCode apiErrorCode = 13
	// ErrCantSendToEthAddrType type when can't send to eth addr code error appeared
	ErrCantSendToEthAddrType apiErrorType = "ErrCantSendToEthAddr"

	// Internal error messages (used for logs or handling errors returned from internal components)

	// errCtxTimeout error message received internally when context reaches timeout
	errCtxTimeout = "context deadline exceeded"

	// ErrIsAtomic filter atomic transactions on POST /transactions-pool
	ErrIsAtomic = "this endpoint does not accept atomic transactions"
	// ErrIsAtomicCode code filter atomic transactions on POST /transactions-pool
	ErrIsAtomicCode = 14
	// ErrIsAtomicType type filter atomic transactions on POST /transactions-pool
	ErrIsAtomicType = "ErrIsAtomic"
)

type apiError struct {
	Err  error
	Code apiErrorCode
	Type apiErrorType
}

type apiErrorResponse struct {
	Message string       `json:"message"`
	Code    apiErrorCode `json:"code"`
	Type    apiErrorType `json:"type"`
}

func (a apiError) Error() string {
	return a.Err.Error()
}
