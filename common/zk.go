package common

// circuit parameters
// absolute maximum of L1 or L2 transactions allowed
// uncomment: type nTx uint32

// merkle tree depth
// uncomment: type nLevels uint32

// absolute maximum of L1 transaction allowed
// uncomment: type maxL1Tx uint32

//absolute maximum of fee transactions allowed
// uncomment: type maxFeeTx uint32

// ZKInputs represents the inputs that will be used to generate the zkSNARK proof
type ZKInputs struct {
	// inputs for final `hashGlobalInputs`
	// oldLastIdx is the last index assigned to an account
	// uncomment: oldLastIdx *big.Int // uint64 (max nLevels bits)
	// oldStateRoot is the current state merkle tree root
	// uncomment: oldStateRoot *big.Int // Hash
	// globalChainID is the blockchain ID (0 for Ethereum mainnet). This value can be get from the smart contract.
	// uncomment: globalChainID *big.Int // uint16
	// feeIdxs is an array of merkle tree indexes where the coordinator will receive the accumulated fees
	// uncomment: feeIdxs []*big.Int // uint64 (max nLevels bits), len: [maxFeeTx]

	// accumulate fees
	// feePlanTokens contains all the tokenIDs for which the fees are being accumulated
	// uncomment: feePlanTokens []*big.Int // uint32 (max 32 bits), len: [maxFeeTx]

	// Intermediary States to parallelize witness computation
	// decode-tx
	// imOnChain indicates if tx is L1 (true) or L2 (false)
	// uncomment: imOnChain []*big.Int // bool, len: [nTx - 1]
	// imOutIdx current index account for each Tx
	// uncomment: imOutIdx []*big.Int // uint64 (max nLevels bits), len: [nTx - 1]
	// rollup-tx
	// imStateRoot root at the moment of the Tx, the state root value once the Tx is processed into the state tree
	// uncomment: imStateRoot []*big.Int // Hash, len: [nTx - 1]
	// imExitTree root at the moment of the Tx the value once the Tx is processed into the exit tree
	// uncomment: imExitRoot []*big.Int // Hash, len: [nTx - 1]
	// imAccFeeOut accumulated fees once the Tx is processed
	// uncomment: imAccFeeOut [][]*big.Int // big.Int, len: [nTx - 1][maxFeeTx]
	// fee-tx
	// imStateRootFee root at the moment of the Tx, the state root value once the Tx is processed into the state tree
	// uncomment: imStateRootFee []*big.Int // Hash, len: [maxFeeTx - 1]
	// imInitStateRootFee state root once all L1-L2 tx are processed (before computing the fees-tx)
	// uncomment: imInitStateRootFee *big.Int // Hash
	// imFinalAccFee final accumulated fees (before computing the fees-tx)
	// uncomment: imFinalAccFee []*big.Int // big.Int, len: [maxFeeTx - 1]

	// transaction L1-L2
	// txCompressedData
	// uncomment: txCompressedData []*big.Int // big.Int (max 251 bits), len: [nTx]
	// txCompressedDataV2
	// uncomment: txCompressedDataV2 []*big.Int // big.Int (max 193 bits), len: [nTx]
	// fromIdx
	// uncomment: fromIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// auxFromIdx is the Idx of the new created account which is consequence of a L1CreateAccountTx
	// uncomment: auxFromIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]

	// toIdx
	// uncomment: toIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// auxToIdx is the Idx of the Tx that has 'toIdx==0', is the coordinator who will find which Idx corresponds to the 'toBjjAy' or 'toEthAddr'
	// uncomment: auxToIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// toBjjAy
	// uncomment: toBjjAy []*big.Int // big.Int, len: [nTx]
	// toEthAddr
	// uncomment: toEthAddr []*big.Int // ethCommon.Address, len: [nTx]

	// onChain determines if is L1 (1/true) or L2 (0/false)
	// uncomment: onChain []*big.Int // bool, len: [nTx]
	// newAccount boolean (0/1) flag to set L1 tx creates a new account
	// uncomment: newAccount []*big.Int // bool, len: [nTx]
	// rqOffset relative transaction position to be linked. Used to perform atomic transactions.
	// uncomment: rqOffset []*big.Int // uint8 (max 3 bits), len: [nTx]

	// transaction L2 request data
	// rqTxCompressedDataV2
	// uncomment: rqTxCompressedDataV2 []*big.Int // big.Int (max 251 bits), len: [nTx]
	// rqToEthAddr
	// uncomment: rqToEthAddr []*big.Int // ethCommon.Address, len: [nTx]
	// rqToBjjAy
	// uncomment: rqToBjjAy []*big.Int // big.Int, len: [nTx]

	// transaction L2 signature
	// s
	// uncomment: s []*big.Int // big.Int, len: [nTx]
	// r8x
	// uncomment: r8x []*big.Int // big.Int, len: [nTx]
	// r8y
	// uncomment: r8y []*big.Int // big.Int, len: [nTx]

	// transaction L1
	// loadAmountF encoded as float16
	// uncomment: loadAmountF []*big.Int // uint16, len: [nTx]
	// fromEthAddr
	// uncomment: fromEthAddr []*big.Int // ethCommon.Address, len: [nTx]
	// fromBjjCompressed boolean encoded where each value is a *big.Int
	// uncomment: fromBjjCompressed [][]*big.Int // bool array, len: [nTx][256]

	// state 1, value of the sender (from) account leaf
	// uncomment: tokenID1  []*big.Int   // uint32, len: [nTx]
	// uncomment: nonce1    []*big.Int   // uint64 (max 40 bits), len: [nTx]
	// uncomment: sign1     []*big.Int   // bool, len: [nTx]
	// uncomment: balance1  []*big.Int   // big.Int (max 192 bits), len: [nTx]
	// uncomment: ay1       []*big.Int   // big.Int, len: [nTx]
	// uncomment: ethAddr1  []*big.Int   // ethCommon.Address, len: [nTx]
	// uncomment: siblings1 [][]*big.Int // big.Int, len: [nTx][nLevels + 1]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	// uncomment: isOld0_1  []*big.Int // bool, len: [nTx]
	// uncomment: oldKey1   []*big.Int // uint64 (max 40 bits), len: [nTx]
	// uncomment: oldValue1 []*big.Int // Hash, len: [nTx]

	// state 2, value of the receiver (to) account leaf
	// uncomment: tokenID2  []*big.Int   // uint32, len: [nTx]
	// uncomment: nonce2    []*big.Int   // uint64 (max 40 bits), len: [nTx]
	// uncomment: sign2     []*big.Int   // bool, len: [nTx]
	// uncomment: balance2  []*big.Int   // big.Int (max 192 bits), len: [nTx]
	// uncomment: ay2       []*big.Int   // big.Int, len: [nTx]
	// uncomment: ethAddr2  []*big.Int   // ethCommon.Address, len: [nTx]
	// uncomment: siblings2 [][]*big.Int // big.Int, len: [nTx][nLevels + 1]
	// newExit determines if an exit transaction has to create a new leaf in the exit tree
	// uncomment: newExit []*big.Int // bool, len: [nTx]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	// uncomment: isOld0_2  []*big.Int // bool, len: [nTx]
	// uncomment: oldKey2   []*big.Int // uint64 (max 40 bits), len: [nTx]
	// uncomment: oldValue2 []*big.Int // Hash, len: [nTx]

	// state 3, value of the account leaf receiver of the Fees
	// fee tx
	// State fees
	// uncomment: tokenID3  []*big.Int   // uint32, len: [maxFeeTx]
	// uncomment: nonce3    []*big.Int   // uint64 (max 40 bits), len: [maxFeeTx]
	// uncomment: sign3     []*big.Int   // bool, len: [maxFeeTx]
	// uncomment: balance3  []*big.Int   // big.Int (max 192 bits), len: [maxFeeTx]
	// uncomment: ay3       []*big.Int   // big.Int, len: [maxFeeTx]
	// uncomment: ethAddr3  []*big.Int   // ethCommon.Address, len: [maxFeeTx]
	// uncomment: siblings3 [][]*big.Int // Hash, len: [maxFeeTx][nLevels + 1]
}

// CallDataForge TBD
type CallDataForge struct {
	// TBD
}
