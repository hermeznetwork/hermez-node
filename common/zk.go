// Package common contains all the common data structures used at the
// hermez-node, zk.go contains the zkSnark inputs used to generate the proof
//nolint:deadcode,structcheck, unused
package common

import "math/big"

// circuit parameters
// absolute maximum of L1 or L2 transactions allowed
type nTx uint32

// merkle tree depth
type nLevels uint32

// absolute maximum of L1 transaction allowed
type maxL1Tx uint32

//absolute maximum of fee transactions allowed
type maxFeeTx uint32

// ZKInputs represents the inputs that will be used to generate the zkSNARK proof
type ZKInputs struct {
	// inputs for final `hashGlobalInputs`
	// oldLastIdx is the last index assigned to an account
	oldLastIdx *big.Int // uint64 (max nLevels bits)
	// oldStateRoot is the current state merkle tree root
	oldStateRoot *big.Int // Hash
	// globalChainID is the blockchain ID (0 for Ethereum mainnet). This value can be get from the smart contract.
	globalChainID *big.Int // uint16
	// feeIdxs is an array of merkle tree indexes where the coordinator will receive the accumulated fees
	feeIdxs []*big.Int // uint64 (max nLevels bits), len: [maxFeeTx]

	// accumulate fees
	// feePlanTokens contains all the tokenIDs for which the fees are being accumulated
	feePlanTokens []*big.Int // uint32 (max 32 bits), len: [maxFeeTx]

	// Intermediary States to parallelize witness computation
	// decode-tx
	// imOnChain indicates if tx is L1 (true) or L2 (false)
	imOnChain []*big.Int // bool, len: [nTx - 1]
	// imOutIdx current index account for each Tx
	imOutIdx []*big.Int // uint64 (max nLevels bits), len: [nTx - 1]
	// rollup-tx
	// imStateRoot root at the moment of the Tx, the state root value once the Tx is processed into the state tree
	imStateRoot []*big.Int // Hash, len: [nTx - 1]
	// imExitTree root at the moment of the Tx the value once the Tx is processed into the exit tree
	imExitRoot []*big.Int // Hash, len: [nTx - 1]
	// imAccFeeOut accumulated fees once the Tx is processed
	imAccFeeOut [][]*big.Int // big.Int, len: [nTx - 1][maxFeeTx]
	// fee-tx
	// imStateRootFee root at the moment of the Tx, the state root value once the Tx is processed into the state tree
	imStateRootFee []*big.Int // Hash, len: [maxFeeTx - 1]
	// imInitStateRootFee state root once all L1-L2 tx are processed (before computing the fees-tx)
	imInitStateRootFee *big.Int // Hash
	// imFinalAccFee final accumulated fees (before computing the fees-tx)
	imFinalAccFee []*big.Int // big.Int, len: [maxFeeTx - 1]

	// transaction L1-L2
	// txCompressedData
	txCompressedData []*big.Int // big.Int (max 251 bits), len: [nTx]
	// txCompressedDataV2
	txCompressedDataV2 []*big.Int // big.Int (max 193 bits), len: [nTx]
	// fromIdx
	fromIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// auxFromIdx is the Idx of the new created account which is consequence of a L1CreateAccountTx
	auxFromIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]

	// toIdx
	toIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// auxToIdx is the Idx of the Tx that has 'toIdx==0', is the coordinator who will find which Idx corresponds to the 'toBjjAy' or 'toEthAddr'
	auxToIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// toBjjAy
	toBjjAy []*big.Int // big.Int, len: [nTx]
	// toEthAddr
	toEthAddr []*big.Int // ethCommon.Address, len: [nTx]

	// onChain determines if is L1 (1/true) or L2 (0/false)
	onChain []*big.Int // bool, len: [nTx]
	// newAccount boolean (0/1) flag to set L1 tx creates a new account
	newAccount []*big.Int // bool, len: [nTx]
	// rqOffset relative transaction position to be linked. Used to perform atomic transactions.
	rqOffset []*big.Int // uint8 (max 3 bits), len: [nTx]

	// transaction L2 request data
	// rqTxCompressedDataV2
	rqTxCompressedDataV2 []*big.Int // big.Int (max 251 bits), len: [nTx]
	// rqToEthAddr
	rqToEthAddr []*big.Int // ethCommon.Address, len: [nTx]
	// rqToBjjAy
	rqToBjjAy []*big.Int // big.Int, len: [nTx]

	// transaction L2 signature
	// s
	s []*big.Int // big.Int, len: [nTx]
	// r8x
	r8x []*big.Int // big.Int, len: [nTx]
	// r8y
	r8y []*big.Int // big.Int, len: [nTx]

	// transaction L1
	// loadAmountF encoded as float16
	loadAmountF []*big.Int // uint16, len: [nTx]
	// fromEthAddr
	fromEthAddr []*big.Int // ethCommon.Address, len: [nTx]
	// fromBjjCompressed boolean encoded where each value is a *big.Int
	fromBjjCompressed [][]*big.Int // bool array, len: [nTx][256]

	// state 1, value of the sender (from) account leaf
	tokenID1  []*big.Int   // uint32, len: [nTx]
	nonce1    []*big.Int   // uint64 (max 40 bits), len: [nTx]
	sign1     []*big.Int   // bool, len: [nTx]
	balance1  []*big.Int   // big.Int (max 192 bits), len: [nTx]
	ay1       []*big.Int   // big.Int, len: [nTx]
	ethAddr1  []*big.Int   // ethCommon.Address, len: [nTx]
	siblings1 [][]*big.Int // big.Int, len: [nTx][nLevels + 1]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	isOld0_1  []*big.Int // bool, len: [nTx]
	oldKey1   []*big.Int // uint64 (max 40 bits), len: [nTx]
	oldValue1 []*big.Int // Hash, len: [nTx]

	// state 2, value of the receiver (to) account leaf
	tokenID2  []*big.Int   // uint32, len: [nTx]
	nonce2    []*big.Int   // uint64 (max 40 bits), len: [nTx]
	sign2     []*big.Int   // bool, len: [nTx]
	balance2  []*big.Int   // big.Int (max 192 bits), len: [nTx]
	ay2       []*big.Int   // big.Int, len: [nTx]
	ethAddr2  []*big.Int   // ethCommon.Address, len: [nTx]
	siblings2 [][]*big.Int // big.Int, len: [nTx][nLevels + 1]
	// newExit determines if an exit transaction has to create a new leaf in the exit tree
	newExit []*big.Int // bool, len: [nTx]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	isOld0_2  []*big.Int // bool, len: [nTx]
	oldKey2   []*big.Int // uint64 (max 40 bits), len: [nTx]
	oldValue2 []*big.Int // Hash, len: [nTx]

	// state 3, value of the account leaf receiver of the Fees
	// fee tx
	// State fees
	tokenID3  []*big.Int   // uint32, len: [maxFeeTx]
	nonce3    []*big.Int   // uint64 (max 40 bits), len: [maxFeeTx]
	sign3     []*big.Int   // bool, len: [maxFeeTx]
	balance3  []*big.Int   // big.Int (max 192 bits), len: [maxFeeTx]
	ay3       []*big.Int   // big.Int, len: [maxFeeTx]
	ethAddr3  []*big.Int   // ethCommon.Address, len: [maxFeeTx]
	siblings3 [][]*big.Int // Hash, len: [maxFeeTx][nLevels + 1]
}

// CallDataForge TBD
type CallDataForge struct {
	// TBD
}
