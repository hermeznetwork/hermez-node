package txselector

const (
	// Error messages showed in the info field from tx table

	// ErrExitAmount error message returned when an exit with amount 0 is received
	ErrExitAmount = "Exits with amount 0 make no sense, not accepting to prevent unintended transactions"
	// ErrExitAmountCode error code
	ErrExitAmountCode int = 1
	// ErrExitAmountType error type
	ErrExitAmountType string = "ErrExit0Amount"

	// ErrUnsupportedMaxNumBatch error message returned when the maximum batch number is exceeded
	ErrUnsupportedMaxNumBatch = "MaxNumBatch exceeded"
	// ErrUnsupportedMaxNumBatchCode error code
	ErrUnsupportedMaxNumBatchCode int = 2
	// ErrUnsupportedMaxNumBatchType error type
	ErrUnsupportedMaxNumBatchType string = "ErrUnsupportedMaxNumBatch"

	// ErrSenderNotEnoughBalance error message returned if the sender doesn't have enough balance to send the tx
	ErrSenderNotEnoughBalance = "Tx not selected due to not enough Balance at the sender. "
	// ErrSenderNotEnoughBalanceCode error code
	ErrSenderNotEnoughBalanceCode int = 11
	// ErrSenderNotEnoughBalanceType error type
	ErrSenderNotEnoughBalanceType string = "ErrSenderNotEnoughBalance"

	// ErrNoCurrentNonce error message returned if the sender doesn't use the current nonce
	ErrNoCurrentNonce = "Tx not selected due to not current Nonce. "
	// ErrNoCurrentNonceCode error code
	ErrNoCurrentNonceCode int = 12
	// ErrNoCurrentNonceType error type
	ErrNoCurrentNonceType string = "ErrNoCurrentNonce"

	// ErrNotEnoughSpaceL1Coordinator error message returned if L2Tx depends on a L1CoordinatorTx and there is not enough space for L1Coordinator
	ErrNotEnoughSpaceL1Coordinator = "Tx not selected because the L2Tx depends on a L1CoordinatorTx and there is not enough space for L1Coordinator"
	// ErrNotEnoughSpaceL1CoordinatorCode error code
	ErrNotEnoughSpaceL1CoordinatorCode int = 13
	// ErrNotEnoughSpaceL1CoordinatorType error type
	ErrNotEnoughSpaceL1CoordinatorType string = "ErrNotEnoughSpaceL1Coordinator"

	// ErrTxDiscartedInProcessTxToEthAddrBJJ error message returned if tx is discarted in processTxToEthAddrBJJ
	ErrTxDiscartedInProcessTxToEthAddrBJJ = "Tx not selected (in processTxToEthAddrBJJ)"
	// ErrTxDiscartedInProcessTxToEthAddrBJJCode error code
	ErrTxDiscartedInProcessTxToEthAddrBJJCode int = 14
	// ErrTxDiscartedInProcessTxToEthAddrBJJType error type
	ErrTxDiscartedInProcessTxToEthAddrBJJType string = "ErrTxDiscartedInProcessTxToEthAddrBJJ"

	// ErrToIdxNotFound error message returned if the toIdx is not found in the stateDB
	ErrToIdxNotFound = "Tx not selected due to tx.ToIdx not found in StateDB. "
	// ErrToIdxNotFoundCode error code
	ErrToIdxNotFoundCode int = 15
	// ErrToIdxNotFoundType error type
	ErrToIdxNotFoundType string = "ErrToIdxNotFound"

	// ErrTxDiscartedInProcessL2Tx error message returned if tx is discarted in ProcessL2Tx
	ErrTxDiscartedInProcessL2Tx = "Tx not selected (in ProcessL2Tx)"
	// ErrTxDiscartedInProcessL2TxCode error code
	ErrTxDiscartedInProcessL2TxCode int = 16
	// ErrTxDiscartedInProcessL2TxType error type
	ErrTxDiscartedInProcessL2TxType string = "ErrTxDiscartedInProcessL2Tx"

	// ErrNoAvailableSlots error message returned if there is no available slots for L2Txs
	ErrNoAvailableSlots = "Tx not selected due not available slots for L2Txs"
	// ErrNoAvailableSlotsCode error code
	ErrNoAvailableSlotsCode int = 17
	// ErrNoAvailableSlotsType error type
	ErrNoAvailableSlotsType string = "ErrNoAvailableSlots"

	// ErrInvalidAtomicGroup error message returned if an atomic group is malformed
	ErrInvalidAtomicGroup = "Tx not selected because it belongs to an atomic group with missing transactions or bad requested transaction"
	// ErrInvalidAtomicGroupCode error code
	ErrInvalidAtomicGroupCode int = 18
	// ErrInvalidAtomicGroupType error type
	ErrInvalidAtomicGroupType string = "ErrInvalidAtomicGroup"
)
