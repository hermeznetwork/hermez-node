package txselector

var (
	// ErrAtomicGroupFail one or more atomic transaction failed, failing all others
	ErrAtomicGroupFail = "one or more atomic transaction failed"
	// ErrCoordIdxNotFound coordinator idx not found for the token id
	ErrCoordIdxNotFound = "coordinator idx not found, waiting the next batch"
	// ErrInvalidToIdx trying to make a TxTypeTransferToEthAddr or TxTypeTransferToBJJ with a ToIdx greater then zero
	ErrInvalidToIdx = "this transfer cannot have a toIdx"
	// ErrInvalidToFAddr TransferToBJJ requires hez:0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
	ErrInvalidToFAddr = "the destination ETH address must be hez:0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF to perform a TransferToBJJ"
	// ErrInvalidToEthAddr making a transfer with a empty or a 0xFFF.. ETH address
	ErrInvalidToEthAddr = "invalid ETH address"
	// ErrInvalidToBjjAddr making a transfer with a empty BJJ address
	ErrInvalidToBjjAddr = "invalid BJJ address"
	// ErrInsufficientFunds making a transfer without sufficient balance
	ErrInsufficientFunds = "insufficient funds"
	// ErrSenderNotFound transaction sender not found
	ErrSenderNotFound = "transaction sender not found"
	// ErrRecipientNotFound transaction recipient not found
	ErrRecipientNotFound = "transaction recipient not found"
	// ErrInvalidNonce invalid nonce
	ErrInvalidNonce = "invalid nonce"
	// ErrExitZeroAmount exit with amount zero
	ErrExitZeroAmount = "invalid exit, zero amount"
	// ErrInvalidExitToIdx invalid ToIdx for exit transaction
	ErrInvalidExitToIdx = "exit transactions must have ToIdx equal to one"
	// ErrMaxL2TxSlot tx not selected due not available slots for L2 transactions
	ErrMaxL2TxSlot = "tx not selected due not available slots for L2Txs"
	// ErrInvalidRqTx request transaction id not found
	ErrInvalidRqTx = "request transaction id not found"
	// ErrUnexpectedRqOffset One of the transactions within an atomic group has RqOffset == 0
	ErrUnexpectedRqOffset = "One of the transactions within an atomic group has RqOffset == 0"
)
