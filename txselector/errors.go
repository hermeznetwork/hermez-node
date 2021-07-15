package txselector

type txSelectorErrorCode uint
type txSelectorType string

const (
	// Error messages showed in the info field from tx table

	// ErrExitAmount error message returned when an exit with amount 0 is received
	ErrExitAmount = "Exits with amount 0 make no sense, not accepting to prevent unintended transactions"
	// ErrExitAmountCode error code
	ErrExitAmountCode txSelectorErrorCode = 1
	// ErrExitAmountType error type
	ErrExitAmountType txSelectorType = "ErrExit0Amount"

	// ErrUnsupportedMaxNumBatch error message returned when the maximum batch number is exceeded
	ErrUnsupportedMaxNumBatch = "MaxNumBatch exceeded"
	// ErrUnsupportedMaxNumBatchCode error code
	ErrUnsupportedMaxNumBatchCode txSelectorErrorCode = 2
	// ErrUnsupportedMaxNumBatchType error type
	ErrUnsupportedMaxNumBatchType txSelectorType = "ErrUnsupportedMaxNumBatch"

	// ErrSenderNotEnoughBalance error message returned if the sender doesn't have enough balance to send the tx
	ErrSenderNotEnoughBalance = "Tx not selected due to not enough Balance at the sender. "
	// ErrSenderNotEnoughBalanceCode error code
	ErrSenderNotEnoughBalanceCode txSelectorErrorCode = 11
	// ErrSenderNotEnoughBalanceType error type
	ErrSenderNotEnoughBalanceType txSelectorType = "ErrSenderNotEnoughBalance"

	// ErrNoCurrentNonce error message returned if the sender doesn't use the current nonce
	ErrNoCurrentNonce = "Tx not selected due to not current Nonce. "
	// ErrNoCurrentNonceCode error code
	ErrNoCurrentNonceCode txSelectorErrorCode = 12
	// ErrNoCurrentNonceType error type
	ErrNoCurrentNonceType txSelectorType = "ErrNoCurrentNonce"

	// ErrNotEnoughSpaceL1Coordinator error message returned if L2Tx depends on a L1CoordinatorTx and there is not enough space for L1Coordinator
	ErrNotEnoughSpaceL1Coordinator = "Tx not selected because the L2Tx depends on a L1CoordinatorTx and there is not enough space for L1Coordinator"
	// ErrNotEnoughSpaceL1CoordinatorCode error code
	ErrNotEnoughSpaceL1CoordinatorCode txSelectorErrorCode = 13
	// ErrNotEnoughSpaceL1CoordinatorType error type
	ErrNotEnoughSpaceL1CoordinatorType txSelectorType = "ErrNotEnoughSpaceL1Coordinator"

	// ErrTxDiscartedInProcessTxToEthAddrBJJ error message returned if tx is discarted in processTxToEthAddrBJJ
	ErrTxDiscartedInProcessTxToEthAddrBJJ = "Tx not selected (in processTxToEthAddrBJJ)"
	// ErrTxDiscartedInProcessTxToEthAddrBJJCode error code
	ErrTxDiscartedInProcessTxToEthAddrBJJCode txSelectorErrorCode = 14
	// ErrTxDiscartedInProcessTxToEthAddrBJJType error type
	ErrTxDiscartedInProcessTxToEthAddrBJJType txSelectorType = "ErrTxDiscartedInProcessTxToEthAddrBJJ"

	// ErrToIdxNotFound error message returned if the toIdx is not found in the stateDB
	ErrToIdxNotFound = "Tx not selected due to tx.ToIdx not found in StateDB. "
	// ErrToIdxNotFoundCode error code
	ErrToIdxNotFoundCode txSelectorErrorCode = 15
	// ErrToIdxNotFoundType error type
	ErrToIdxNotFoundType txSelectorType = "ErrToIdxNotFound"

	// ErrTxDiscartedInProcessL2Tx error message returned if tx is discarted in ProcessL2Tx
	ErrTxDiscartedInProcessL2Tx = "Tx not selected (in ProcessL2Tx)"
	// ErrTxDiscartedInProcessL2TxCode error code
	ErrTxDiscartedInProcessL2TxCode txSelectorErrorCode = 16
	// ErrTxDiscartedInProcessL2TxType error type
	ErrTxDiscartedInProcessL2TxType txSelectorType = "ErrTxDiscartedInProcessL2Tx"

	// ErrUnselectableAtomicGroup error message returned if tx is discarted in ProcessL2Tx
	ErrUnselectableAtomicGroup = "Unselectable atomic group"
	// ErrUnselectableAtomicGroupCode error code
	ErrUnselectableAtomicGroupCode txSelectorErrorCode = 17
	// ErrUnselectableAtomicGroupType error type
	ErrUnselectableAtomicGroupType txSelectorType = "ErrUnselectableAtomicGroup"
)

type txSelectorError struct {
	Message string
	Code    txSelectorErrorCode
	Type    txSelectorType
}
