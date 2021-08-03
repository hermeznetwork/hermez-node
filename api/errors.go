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
	ErrSQLTimeout = "The node is under heavy pressure, please try again later"
	// ErrSQLTimeoutCode code for sql timeout error
	ErrSQLTimeoutCode apiErrorCode = 3
	// ErrSQLTimeoutType type for sql timeout type
	ErrSQLTimeoutType apiErrorType = "ErrSQLTimeout"

	// ErrSQLNoRows error message returned when there is no such records
	ErrSQLNoRows = "record(s) were not found for this query and/or the parameters entered"
	// ErrSQLNoRowsCode code for no rows error
	ErrSQLNoRowsCode apiErrorCode = 4
	// ErrSQLNoRowsType type for now rows error
	ErrSQLNoRowsType apiErrorType = "ErrSQLNoRows"

	// ErrExitAmount0 error message returned when receiving (and rejecting) a tx of type exit with amount 0
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

	// ErrNotAtomicTxsInPostPoolTx error message returned when received an non-atomic transaction inside atomic pool
	ErrNotAtomicTxsInPostPoolTx = "atomic transactions are only accepted in POST /atomic-pool"
	// ErrNotAtomicTxsInPostPoolTxCode code filter atomic transactions on POST /transactions-pool
	ErrNotAtomicTxsInPostPoolTxCode apiErrorCode = 14
	// ErrNotAtomicTxsInPostPoolTxType type filter atomic transactions on POST /transactions-pool
	ErrNotAtomicTxsInPostPoolTxType apiErrorType = "ErrNotAtomicTxsInPostPoolTx"

	// ErrFailedToGetCurrentBlockCode code when can't get current block in /slots request
	ErrFailedToGetCurrentBlockCode apiErrorCode = 15
	// ErrFailedToGetCurrentBlockType type when can't get current block in /slots request
	ErrFailedToGetCurrentBlockType apiErrorType = "ErrFailedToGetCurrentBlock"

	// ErrFailedToGetAuctionVarsCode code when can't get auction vars in /slots request
	ErrFailedToGetAuctionVarsCode apiErrorCode = 16
	// ErrFailedToGetAuctionVarsType type when can't get auction vars in /slots request
	ErrFailedToGetAuctionVarsType apiErrorType = "ErrFailedToGetAuctionVars"

	// ErrFailedToAddEmptySlotCode code when can't add empty slot in /slots request
	ErrFailedToAddEmptySlotCode apiErrorCode = 17
	// ErrFailedToAddEmptySlotType type when can't add empty slot in /slots request
	ErrFailedToAddEmptySlotType apiErrorType = "ErrFailedToAddEmptySlot"

	// ErrTxsNotAtomic error message returned when receiving (and rejecting) txs in the atomic endpoint with not all txs being atomic
	ErrTxsNotAtomic = "there is at least one transaction in the payload that could be forged without the others"
	// ErrTxsNotAtomicCode error code
	ErrTxsNotAtomicCode apiErrorCode = 18
	// ErrTxsNotAtomicType error type
	ErrTxsNotAtomicType apiErrorType = "ErrTxsNotAtomic"

	// ErrSingleTxInAtomicEndpoint only one tx sent to the atomic-pool endpoint
	ErrSingleTxInAtomicEndpoint = "to use the atomic-pool endpoint at least two transactions are required"
	// ErrSingleTxInAtomicEndpointCode error code
	ErrSingleTxInAtomicEndpointCode apiErrorCode = 19
	// ErrSingleTxInAtomicEndpointType error type
	ErrSingleTxInAtomicEndpointType apiErrorType = "ErrSingleTxInAtomicEndpoint"

	// ErrInvalidRqOffset error message returned when received an invalid request offset
	ErrInvalidRqOffset = "invalid requestOffset. Valid values goes from 0 to 7"

	// ErrRqOffsetOutOfBounds error message returned when transaction tries to access another out the bounds of the array
	ErrRqOffsetOutOfBounds = "one of the transactions requested another one outside the bounds of the provided array"
	// ErrRqOffsetOutOfBoundsCode error code
	ErrRqOffsetOutOfBoundsCode = 20
	// ErrRqOffsetOutOfBoundsType error type
	ErrRqOffsetOutOfBoundsType = "ErrRqOffsetOutOfBounds"

	// ErrInvalidAtomicGroupID error message returned when received an invalid AtomicGroupID
	ErrInvalidAtomicGroupID = "invalid atomicGroupId"
	// ErrInvalidAtomicGroupIDCode error code
	ErrInvalidAtomicGroupIDCode = 21
	// ErrInvalidAtomicGroupIDType error type
	ErrInvalidAtomicGroupIDType = "ErrInvalidAtomicGroupID"

	// ErrFailedToFindOffsetToRelativePositionCode error code when can't find offset to relative position
	ErrFailedToFindOffsetToRelativePositionCode = 22
	// ErrFailedToFindOffsetToRelativePositionType error type
	ErrFailedToFindOffsetToRelativePositionType = "ErrFailedToFindOffsetToRelativePosition"

	// ErrFeeTooLowCode code for fee too low error
	ErrFeeTooLowCode apiErrorCode = 23
	// ErrFeeTooLowType type for fee too low error
	ErrFeeTooLowType apiErrorType = "ErrFeeTooLow"

	// ErrFeeTooBigCode code for fee too big error
	ErrFeeTooBigCode apiErrorCode = 24
	// ErrFeeTooBigType type for fee too big error
	ErrFeeTooBigType apiErrorType = "ErrFeeTooBig"

	// ErrNothingToUpdateCode code for nothing to update error
	ErrNothingToUpdateCode apiErrorCode = 25
	// ErrNothingToUpdateType type for nothing to update type
	ErrNothingToUpdateType apiErrorType = "ErrNothingToUpdate"

	// ErrUnsupportedMaxNumBatch error message returned when tx.MaxNumBatch != 0 until the feature is fully implemented
	ErrUnsupportedMaxNumBatch = "currently only supported value for maxNumBatch is 0, this will change soon when the feature is fully implemented"

	// Internal error messages (used for logs or handling errors returned from internal components)

	// errCtxTimeout error message received internally when context reaches timeout
	errCtxTimeout = "context deadline exceeded"
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
