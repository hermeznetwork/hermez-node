// Package common contains all the common data structures used at the
// hermez-node, zk.go contains the zkSnark inputs used to generate the proof
//nolint:deadcode,structcheck,unused
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
	//
	// General
	//

	// inputs for final `hashGlobalInputs`
	// OldLastIdx is the last index assigned to an account
	OldLastIdx *big.Int // uint64 (max nLevels bits)
	// OldStateRoot is the current state merkle tree root
	OldStateRoot *big.Int // Hash
	// GlobalChainID is the blockchain ID (0 for Ethereum mainnet). This value can be get from the smart contract.
	GlobalChainID *big.Int // uint16
	// FeeIdxs is an array of merkle tree indexes where the coordinator will receive the accumulated fees
	FeeIdxs []*big.Int // uint64 (max nLevels bits), len: [maxFeeTx]

	// accumulate fees
	// FeePlanTokens contains all the tokenIDs for which the fees are being accumulated
	FeePlanTokens []*big.Int // uint32 (max 32 bits), len: [maxFeeTx]

	//
	// Txs (L1&L2)
	//

	// transaction L1-L2
	// TxCompressedData
	TxCompressedData []*big.Int // big.Int (max 251 bits), len: [nTx]
	// TxCompressedDataV2, only used in L2Txs, in L1Txs is set to 0
	TxCompressedDataV2 []*big.Int // big.Int (max 193 bits), len: [nTx]

	// FromIdx
	FromIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// AuxFromIdx is the Idx of the new created account which is consequence of a L1CreateAccountTx
	AuxFromIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]

	// ToIdx
	ToIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// AuxToIdx is the Idx of the Tx that has 'toIdx==0', is the coordinator who will find which Idx corresponds to the 'toBJJAy' or 'toEthAddr'
	AuxToIdx []*big.Int // uint64 (max nLevels bits), len: [nTx]
	// ToBJJAy
	ToBJJAy []*big.Int // big.Int, len: [nTx]
	// ToEthAddr
	ToEthAddr []*big.Int // ethCommon.Address, len: [nTx]

	// OnChain determines if is L1 (1/true) or L2 (0/false)
	OnChain []*big.Int // bool, len: [nTx]
	// NewAccount boolean (0/1) flag set 'true' when L1 tx creates a new account (fromIdx==0)
	NewAccount []*big.Int // bool, len: [nTx]

	//
	// Txs/L1Txs
	//
	// transaction L1
	// LoadAmountF encoded as float16
	LoadAmountF []*big.Int // uint16, len: [nTx]
	// FromEthAddr
	FromEthAddr []*big.Int // ethCommon.Address, len: [nTx]
	// FromBJJCompressed boolean encoded where each value is a *big.Int
	FromBJJCompressed [][256]*big.Int // bool array, len: [nTx][256]

	//
	// Txs/L2Txs
	//

	// RqOffset relative transaction position to be linked. Used to perform atomic transactions.
	RqOffset []*big.Int // uint8 (max 3 bits), len: [nTx]

	// transaction L2 request data
	// RqTxCompressedDataV2
	RqTxCompressedDataV2 []*big.Int // big.Int (max 251 bits), len: [nTx]
	// RqToEthAddr
	RqToEthAddr []*big.Int // ethCommon.Address, len: [nTx]
	// RqToBJJAy
	RqToBJJAy []*big.Int // big.Int, len: [nTx]

	// transaction L2 signature
	// S
	S []*big.Int // big.Int, len: [nTx]
	// R8x
	R8x []*big.Int // big.Int, len: [nTx]
	// R8y
	R8y []*big.Int // big.Int, len: [nTx]

	//
	// State MerkleTree Leafs transitions
	//

	// state 1, value of the sender (from) account leaf
	TokenID1  []*big.Int   // uint32, len: [nTx]
	Nonce1    []*big.Int   // uint64 (max 40 bits), len: [nTx]
	Sign1     []*big.Int   // bool, len: [nTx]
	Balance1  []*big.Int   // big.Int (max 192 bits), len: [nTx]
	Ay1       []*big.Int   // big.Int, len: [nTx]
	EthAddr1  []*big.Int   // ethCommon.Address, len: [nTx]
	Siblings1 [][]*big.Int // big.Int, len: [nTx][nLevels + 1]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	IsOld0_1  []*big.Int // bool, len: [nTx]
	OldKey1   []*big.Int // uint64 (max 40 bits), len: [nTx]
	OldValue1 []*big.Int // Hash, len: [nTx]

	// state 2, value of the receiver (to) account leaf
	// if Tx is an Exit, state 2 is used for the Exit Merkle Proof
	TokenID2  []*big.Int   // uint32, len: [nTx]
	Nonce2    []*big.Int   // uint64 (max 40 bits), len: [nTx]
	Sign2     []*big.Int   // bool, len: [nTx]
	Balance2  []*big.Int   // big.Int (max 192 bits), len: [nTx]
	Ay2       []*big.Int   // big.Int, len: [nTx]
	EthAddr2  []*big.Int   // ethCommon.Address, len: [nTx]
	Siblings2 [][]*big.Int // big.Int, len: [nTx][nLevels + 1]
	// newExit determines if an exit transaction has to create a new leaf in the exit tree
	NewExit []*big.Int // bool, len: [nTx]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	IsOld0_2  []*big.Int // bool, len: [nTx]
	OldKey2   []*big.Int // uint64 (max 40 bits), len: [nTx]
	OldValue2 []*big.Int // Hash, len: [nTx]

	// state 3, value of the account leaf receiver of the Fees
	// fee tx
	// State fees
	TokenID3  []*big.Int   // uint32, len: [maxFeeTx]
	Nonce3    []*big.Int   // uint64 (max 40 bits), len: [maxFeeTx]
	Sign3     []*big.Int   // bool, len: [maxFeeTx]
	Balance3  []*big.Int   // big.Int (max 192 bits), len: [maxFeeTx]
	Ay3       []*big.Int   // big.Int, len: [maxFeeTx]
	EthAddr3  []*big.Int   // ethCommon.Address, len: [maxFeeTx]
	Siblings3 [][]*big.Int // Hash, len: [maxFeeTx][nLevels + 1]

	//
	// Intermediate States
	//

	// Intermediate States to parallelize witness computation
	// decode-tx
	// ISOnChain indicates if tx is L1 (true) or L2 (false)
	ISOnChain []*big.Int // bool, len: [nTx - 1]
	// ISOutIdx current index account for each Tx
	ISOutIdx []*big.Int // uint64 (max nLevels bits), len: [nTx - 1]
	// rollup-tx
	// ISStateRoot root at the moment of the Tx, the state root value once the Tx is processed into the state tree
	ISStateRoot []*big.Int // Hash, len: [nTx - 1]
	// ISExitTree root at the moment of the Tx the value once the Tx is processed into the exit tree
	ISExitRoot []*big.Int // Hash, len: [nTx - 1]
	// ISAccFeeOut accumulated fees once the Tx is processed
	ISAccFeeOut [][]*big.Int // big.Int, len: [nTx - 1][maxFeeTx]
	// fee-tx
	// ISStateRootFee root at the moment of the Tx, the state root value once the Tx is processed into the state tree
	ISStateRootFee []*big.Int // Hash, len: [maxFeeTx - 1]
	// ISInitStateRootFee state root once all L1-L2 tx are processed (before computing the fees-tx)
	ISInitStateRootFee *big.Int // Hash
	// ISFinalAccFee final accumulated fees (before computing the fees-tx)
	ISFinalAccFee []*big.Int // big.Int, len: [maxFeeTx - 1]
}

// NewZKInputs returns a pointer to an initialized struct of ZKInputs
func NewZKInputs(nTx, maxFeeTx, nLevels int) *ZKInputs {
	zki := &ZKInputs{}

	// General
	zki.OldLastIdx = big.NewInt(0)
	zki.OldStateRoot = big.NewInt(0)
	zki.GlobalChainID = big.NewInt(0)
	zki.FeeIdxs = newSlice(maxFeeTx)
	zki.FeePlanTokens = newSlice(maxFeeTx)

	// Txs
	zki.TxCompressedData = newSlice(nTx)
	zki.TxCompressedDataV2 = newSlice(nTx)
	zki.FromIdx = newSlice(nTx)
	zki.AuxFromIdx = newSlice(nTx)
	zki.ToIdx = newSlice(nTx)
	zki.AuxToIdx = newSlice(nTx)
	zki.ToBJJAy = newSlice(nTx)
	zki.ToEthAddr = newSlice(nTx)
	zki.OnChain = newSlice(nTx)
	zki.NewAccount = newSlice(nTx)

	// L1
	zki.LoadAmountF = newSlice(nTx)
	zki.FromEthAddr = newSlice(nTx)
	zki.FromBJJCompressed = make([][256]*big.Int, nTx)
	for i := 0; i < len(zki.FromBJJCompressed); i++ {
		// zki.FromBJJCompressed[i] = newSlice(256)
		for j := 0; j < 256; j++ {
			zki.FromBJJCompressed[i][j] = big.NewInt(0)
		}
	}

	// L2
	zki.RqOffset = newSlice(nTx)
	zki.RqTxCompressedDataV2 = newSlice(nTx)
	zki.RqToEthAddr = newSlice(nTx)
	zki.RqToBJJAy = newSlice(nTx)
	zki.S = newSlice(nTx)
	zki.R8x = newSlice(nTx)
	zki.R8y = newSlice(nTx)

	// State MerkleTree Leafs transitions
	zki.TokenID1 = newSlice(nTx)
	zki.Nonce1 = newSlice(nTx)
	zki.Sign1 = newSlice(nTx)
	zki.Balance1 = newSlice(nTx)
	zki.Ay1 = newSlice(nTx)
	zki.EthAddr1 = newSlice(nTx)
	zki.Siblings1 = make([][]*big.Int, nTx)
	for i := 0; i < len(zki.Siblings1); i++ {
		zki.Siblings1[i] = newSlice(nLevels + 1)
	}
	zki.IsOld0_1 = newSlice(nTx)
	zki.OldKey1 = newSlice(nTx)
	zki.OldValue1 = newSlice(nTx)

	zki.TokenID2 = newSlice(nTx)
	zki.Nonce2 = newSlice(nTx)
	zki.Sign2 = newSlice(nTx)
	zki.Balance2 = newSlice(nTx)
	zki.Ay2 = newSlice(nTx)
	zki.EthAddr2 = newSlice(nTx)
	zki.Siblings2 = make([][]*big.Int, nTx)
	for i := 0; i < len(zki.Siblings2); i++ {
		zki.Siblings2[i] = newSlice(nLevels + 1)
	}
	zki.NewExit = newSlice(nTx)
	zki.IsOld0_2 = newSlice(nTx)
	zki.OldKey2 = newSlice(nTx)
	zki.OldValue2 = newSlice(nTx)

	zki.TokenID3 = newSlice(maxFeeTx)
	zki.Nonce3 = newSlice(maxFeeTx)
	zki.Sign3 = newSlice(maxFeeTx)
	zki.Balance3 = newSlice(maxFeeTx)
	zki.Ay3 = newSlice(maxFeeTx)
	zki.EthAddr3 = newSlice(maxFeeTx)
	zki.Siblings3 = make([][]*big.Int, maxFeeTx)
	for i := 0; i < len(zki.Siblings3); i++ {
		zki.Siblings3[i] = newSlice(nLevels + 1)
	}

	// Intermediate States
	zki.ISOnChain = newSlice(nTx - 1)
	zki.ISOutIdx = newSlice(nTx - 1)
	zki.ISStateRoot = newSlice(nTx - 1)
	zki.ISExitRoot = newSlice(nTx - 1)
	zki.ISAccFeeOut = make([][]*big.Int, nTx-1)
	for i := 0; i < len(zki.ISAccFeeOut); i++ {
		zki.ISAccFeeOut[i] = newSlice(maxFeeTx)
	}
	zki.ISStateRootFee = newSlice(maxFeeTx - 1)
	zki.ISInitStateRootFee = big.NewInt(0)
	zki.ISFinalAccFee = newSlice(maxFeeTx - 1)

	return zki
}

// newSlice returns a []*big.Int slice of length n with values initialized at
// 0.
// Is used to initialize all *big.Ints of the ZKInputs data structure, so when
// the transactions are processed and the ZKInputs filled, there is no need to
// set all the elements, and if a transaction does not use a parameter, can be
// leaved as it is in the ZKInputs, as will be 0, so later when using the
// ZKInputs to generate the zkSnark proof there is no 'nil'/'null' values.
func newSlice(n int) []*big.Int {
	s := make([]*big.Int, n)
	for i := 0; i < len(s); i++ {
		s[i] = big.NewInt(0)
	}
	return s
}
