// Package common contains all the common data structures used at the
// hermez-node, zk.go contains the zkSnark inputs used to generate the proof
package common

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	cryptoConstants "github.com/iden3/go-iden3-crypto/constants"
	"github.com/iden3/go-merkletree"
	"github.com/mitchellh/mapstructure"
)

// ZKMetadata contains ZKInputs metadata that is not used directly in the
// ZKInputs result, but to calculate values for Hash check
type ZKMetadata struct {
	// Circuit parameters
	// absolute maximum of L1 or L2 transactions allowed
	NTx uint32
	// merkle tree depth
	NLevels   uint32
	MaxLevels uint32
	// absolute maximum of L1 transaction allowed
	MaxL1Tx uint32
	// total txs allowed
	MaxTx uint32
	// Maximum number of Idxs where Fees can be send in a batch (currently
	// is constant for all circuits: 64)
	MaxFeeIdxs uint32

	L1TxsData             [][]byte
	L1TxsDataAvailability [][]byte
	L2TxsData             [][]byte
	ChainID               uint16

	NewLastIdxRaw   Idx
	NewStateRootRaw *merkletree.Hash
	NewExitRootRaw  *merkletree.Hash
}

// ZKInputs represents the inputs that will be used to generate the zkSNARK proof
type ZKInputs struct {
	Metadata ZKMetadata `json:"-"`

	//
	// General
	//

	// CurrentNumBatch is the current batch number processed
	CurrentNumBatch *big.Int `json:"currentNumBatch"` // uint32
	// inputs for final `hashGlobalInputs`
	// OldLastIdx is the last index assigned to an account
	OldLastIdx *big.Int `json:"oldLastIdx"` // uint64 (max nLevels bits)
	// OldStateRoot is the current state merkle tree root
	OldStateRoot *big.Int `json:"oldStateRoot"` // Hash
	// GlobalChainID is the blockchain ID (0 for Ethereum mainnet). This
	// value can be get from the smart contract.
	GlobalChainID *big.Int `json:"globalChainID"` // uint16
	// FeeIdxs is an array of merkle tree indexes where the coordinator
	// will receive the accumulated fees
	FeeIdxs []*big.Int `json:"feeIdxs"` // uint64 (max nLevels bits), len: [maxFeeIdxs]

	// accumulate fees
	// FeePlanTokens contains all the tokenIDs for which the fees are being accumulated
	FeePlanTokens []*big.Int `json:"feePlanTokens"` // uint32 (max nLevels bits), len: [maxFeeIdxs]

	//
	// Txs (L1&L2)
	//

	// transaction L1-L2
	// TxCompressedData
	TxCompressedData []*big.Int `json:"txCompressedData"` // big.Int (max 251 bits), len: [nTx]
	// TxCompressedDataV2, only used in L2Txs, in L1Txs is set to 0
	TxCompressedDataV2 []*big.Int `json:"txCompressedDataV2"` // big.Int (max 193 bits), len: [nTx]
	// MaxNumBatch is the maximum allowed batch number when the transaction
	// can be processed
	MaxNumBatch []*big.Int `json:"maxNumBatch"` // uint32

	// FromIdx
	FromIdx []*big.Int `json:"fromIdx"` // uint64 (max nLevels bits), len: [nTx]
	// AuxFromIdx is the Idx of the new created account which is consequence of a L1CreateAccountTx
	AuxFromIdx []*big.Int `json:"auxFromIdx"` // uint64 (max nLevels bits), len: [nTx]

	// ToIdx
	ToIdx []*big.Int `json:"toIdx"` // uint64 (max nLevels bits), len: [nTx]
	// AuxToIdx is the Idx of the Tx that has 'toIdx==0', is the
	// coordinator who will find which Idx corresponds to the 'toBJJAy' or
	// 'toEthAddr'
	AuxToIdx []*big.Int `json:"auxToIdx"` // uint64 (max nLevels bits), len: [nTx]
	// ToBJJAy
	ToBJJAy []*big.Int `json:"toBjjAy"` // big.Int, len: [nTx]
	// ToEthAddr
	ToEthAddr []*big.Int `json:"toEthAddr"` // ethCommon.Address, len: [nTx]

	// OnChain determines if is L1 (1/true) or L2 (0/false)
	OnChain []*big.Int `json:"onChain"` // bool, len: [nTx]

	//
	// Txs/L1Txs
	//
	// NewAccount boolean (0/1) flag set 'true' when L1 tx creates a new account (fromIdx==0)
	NewAccount []*big.Int `json:"newAccount"` // bool, len: [nTx]
	// DepositAmountF encoded as float16
	DepositAmountF []*big.Int `json:"loadAmountF"` // uint16, len: [nTx]
	// FromEthAddr
	FromEthAddr []*big.Int `json:"fromEthAddr"` // ethCommon.Address, len: [nTx]
	// FromBJJCompressed boolean encoded where each value is a *big.Int
	FromBJJCompressed [][256]*big.Int `json:"fromBjjCompressed"` // bool array, len: [nTx][256]

	//
	// Txs/L2Txs
	//

	// RqOffset relative transaction position to be linked. Used to perform atomic transactions.
	RqOffset []*big.Int `json:"rqOffset"` // uint8 (max 3 bits), len: [nTx]

	// transaction L2 request data
	// RqTxCompressedDataV2
	RqTxCompressedDataV2 []*big.Int `json:"rqTxCompressedDataV2"` // big.Int (max 251 bits), len: [nTx]
	// RqToEthAddr
	RqToEthAddr []*big.Int `json:"rqToEthAddr"` // ethCommon.Address, len: [nTx]
	// RqToBJJAy
	RqToBJJAy []*big.Int `json:"rqToBjjAy"` // big.Int, len: [nTx]

	// transaction L2 signature
	// S
	S []*big.Int `json:"s"` // big.Int, len: [nTx]
	// R8x
	R8x []*big.Int `json:"r8x"` // big.Int, len: [nTx]
	// R8y
	R8y []*big.Int `json:"r8y"` // big.Int, len: [nTx]

	//
	// State MerkleTree Leafs transitions
	//

	// state 1, value of the sender (from) account leaf. The values at the
	// moment pre-smtprocessor of the update (before updating the Sender
	// leaf).
	TokenID1  []*big.Int   `json:"tokenID1"`  // uint32, len: [nTx]
	Nonce1    []*big.Int   `json:"nonce1"`    // uint64 (max 40 bits), len: [nTx]
	Sign1     []*big.Int   `json:"sign1"`     // bool, len: [nTx]
	Ay1       []*big.Int   `json:"ay1"`       // big.Int, len: [nTx]
	Balance1  []*big.Int   `json:"balance1"`  // big.Int (max 192 bits), len: [nTx]
	EthAddr1  []*big.Int   `json:"ethAddr1"`  // ethCommon.Address, len: [nTx]
	Siblings1 [][]*big.Int `json:"siblings1"` // big.Int, len: [nTx][nLevels + 1]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	IsOld0_1  []*big.Int `json:"isOld0_1"`  // bool, len: [nTx]
	OldKey1   []*big.Int `json:"oldKey1"`   // uint64 (max 40 bits), len: [nTx]
	OldValue1 []*big.Int `json:"oldValue1"` // Hash, len: [nTx]

	// state 2, value of the receiver (to) account leaf
	// if Tx is an Exit, state 2 is used for the Exit Merkle Proof
	TokenID2  []*big.Int   `json:"tokenID2"`  // uint32, len: [nTx]
	Nonce2    []*big.Int   `json:"nonce2"`    // uint64 (max 40 bits), len: [nTx]
	Sign2     []*big.Int   `json:"sign2"`     // bool, len: [nTx]
	Ay2       []*big.Int   `json:"ay2"`       // big.Int, len: [nTx]
	Balance2  []*big.Int   `json:"balance2"`  // big.Int (max 192 bits), len: [nTx]
	EthAddr2  []*big.Int   `json:"ethAddr2"`  // ethCommon.Address, len: [nTx]
	Siblings2 [][]*big.Int `json:"siblings2"` // big.Int, len: [nTx][nLevels + 1]
	// newExit determines if an exit transaction has to create a new leaf in the exit tree
	NewExit []*big.Int `json:"newExit"` // bool, len: [nTx]
	// Required for inserts and deletes, values of the CircomProcessorProof (smt insert proof)
	IsOld0_2  []*big.Int `json:"isOld0_2"`  // bool, len: [nTx]
	OldKey2   []*big.Int `json:"oldKey2"`   // uint64 (max 40 bits), len: [nTx]
	OldValue2 []*big.Int `json:"oldValue2"` // Hash, len: [nTx]

	// state 3, value of the account leaf receiver of the Fees
	// fee tx
	// State fees
	TokenID3  []*big.Int   `json:"tokenID3"`  // uint32, len: [maxFeeIdxs]
	Nonce3    []*big.Int   `json:"nonce3"`    // uint64 (max 40 bits), len: [maxFeeIdxs]
	Sign3     []*big.Int   `json:"sign3"`     // bool, len: [maxFeeIdxs]
	Ay3       []*big.Int   `json:"ay3"`       // big.Int, len: [maxFeeIdxs]
	Balance3  []*big.Int   `json:"balance3"`  // big.Int (max 192 bits), len: [maxFeeIdxs]
	EthAddr3  []*big.Int   `json:"ethAddr3"`  // ethCommon.Address, len: [maxFeeIdxs]
	Siblings3 [][]*big.Int `json:"siblings3"` // Hash, len: [maxFeeIdxs][nLevels + 1]

	//
	// Intermediate States
	//

	// Intermediate States to parallelize witness computation
	// Note: the Intermediate States (IS) of the last transaction does not
	// exist. Meaning that transaction 3 (4th) will fill the parameters
	// FromIdx[3] and ISOnChain[3], but last transaction (nTx-1) will fill
	// FromIdx[nTx-1] but will not fill ISOnChain. That's why IS have
	// length of nTx-1, while the other parameters have length of nTx.
	// Last transaction does not need intermediate state since its output
	// will not be used.

	// decode-tx
	// ISOnChain indicates if tx is L1 (true (1)) or L2 (false (0))
	ISOnChain []*big.Int `json:"imOnChain"` // bool, len: [nTx - 1]
	// ISOutIdx current index account for each Tx
	// Contains the index of the created account in case that the tx is of
	// account creation type.
	ISOutIdx []*big.Int `json:"imOutIdx"` // uint64 (max nLevels bits), len: [nTx - 1]
	// rollup-tx
	// ISStateRoot root at the moment of the Tx (once processed), the state
	// root value once the Tx is processed into the state tree
	ISStateRoot []*big.Int `json:"imStateRoot"` // Hash, len: [nTx - 1]
	// ISExitTree root at the moment (once processed) of the Tx the value
	// once the Tx is processed into the exit tree
	ISExitRoot []*big.Int `json:"imExitRoot"` // Hash, len: [nTx - 1]
	// ISAccFeeOut accumulated fees once the Tx is processed.  Contains the
	// array of FeeAccount Balances at each moment of each Tx processed.
	ISAccFeeOut [][]*big.Int `json:"imAccFeeOut"` // big.Int, len: [nTx - 1][maxFeeIdxs]
	// fee-tx:
	// ISStateRootFee root at the moment of the Tx (once processed), the
	// state root value once the Tx is processed into the state tree
	ISStateRootFee []*big.Int `json:"imStateRootFee"` // Hash, len: [maxFeeIdxs - 1]
	// ISInitStateRootFee state root once all L1-L2 tx are processed
	// (before computing the fees-tx)
	ISInitStateRootFee *big.Int `json:"imInitStateRootFee"` // Hash
	// ISFinalAccFee final accumulated fees (before computing the fees-tx).
	// Contains the final values of the ISAccFeeOut parameter
	ISFinalAccFee []*big.Int `json:"imFinalAccFee"` // big.Int, len: [maxFeeIdxs - 1]
}

func bigIntsToStrings(v interface{}) interface{} {
	switch c := v.(type) {
	case *big.Int:
		return c.String()
	case []*big.Int:
		r := make([]interface{}, len(c))
		for i := range c {
			r[i] = bigIntsToStrings(c[i])
		}
		return r
	case [256]*big.Int:
		r := make([]interface{}, len(c))
		for i := range c {
			r[i] = bigIntsToStrings(c[i])
		}
		return r
	case [][]*big.Int:
		r := make([]interface{}, len(c))
		for i := range c {
			r[i] = bigIntsToStrings(c[i])
		}
		return r
	case [][256]*big.Int:
		r := make([]interface{}, len(c))
		for i := range c {
			r[i] = bigIntsToStrings(c[i])
		}
		return r
	case map[string]interface{}:
		// avoid printing a warning when there is a struct type
	default:
		log.Warnf("bigIntsToStrings unexpected type: %T\n", v)
	}
	return nil
}

// MarshalJSON implements the json marshaler for ZKInputs
func (z ZKInputs) MarshalJSON() ([]byte, error) {
	var m map[string]interface{}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &m,
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	err = dec.Decode(z)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	for k, v := range m {
		m[k] = bigIntsToStrings(v)
	}
	return json.Marshal(m)
}

// NewZKInputs returns a pointer to an initialized struct of ZKInputs
func NewZKInputs(nTx, maxL1Tx, maxTx, maxFeeIdxs, nLevels uint32, currentNumBatch *big.Int) *ZKInputs {
	zki := &ZKInputs{}
	zki.Metadata.NTx = nTx
	zki.Metadata.MaxFeeIdxs = maxFeeIdxs
	zki.Metadata.MaxLevels = uint32(48) //nolint:gomnd
	zki.Metadata.NLevels = nLevels
	zki.Metadata.MaxL1Tx = maxL1Tx
	zki.Metadata.MaxTx = maxTx

	// General
	zki.CurrentNumBatch = currentNumBatch
	zki.OldLastIdx = big.NewInt(0)
	zki.OldStateRoot = big.NewInt(0)
	zki.GlobalChainID = big.NewInt(0) // TODO pass by parameter
	zki.FeeIdxs = newSlice(maxFeeIdxs)
	zki.FeePlanTokens = newSlice(maxFeeIdxs)

	// Txs
	zki.TxCompressedData = newSlice(nTx)
	zki.TxCompressedDataV2 = newSlice(nTx)
	zki.MaxNumBatch = newSlice(nTx)
	zki.FromIdx = newSlice(nTx)
	zki.AuxFromIdx = newSlice(nTx)
	zki.ToIdx = newSlice(nTx)
	zki.AuxToIdx = newSlice(nTx)
	zki.ToBJJAy = newSlice(nTx)
	zki.ToEthAddr = newSlice(nTx)
	zki.OnChain = newSlice(nTx)
	zki.NewAccount = newSlice(nTx)

	// L1
	zki.DepositAmountF = newSlice(nTx)
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
	zki.Ay1 = newSlice(nTx)
	zki.Balance1 = newSlice(nTx)
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
	zki.Ay2 = newSlice(nTx)
	zki.Balance2 = newSlice(nTx)
	zki.EthAddr2 = newSlice(nTx)
	zki.Siblings2 = make([][]*big.Int, nTx)
	for i := 0; i < len(zki.Siblings2); i++ {
		zki.Siblings2[i] = newSlice(nLevels + 1)
	}
	zki.NewExit = newSlice(nTx)
	zki.IsOld0_2 = newSlice(nTx)
	zki.OldKey2 = newSlice(nTx)
	zki.OldValue2 = newSlice(nTx)

	zki.TokenID3 = newSlice(maxFeeIdxs)
	zki.Nonce3 = newSlice(maxFeeIdxs)
	zki.Sign3 = newSlice(maxFeeIdxs)
	zki.Ay3 = newSlice(maxFeeIdxs)
	zki.Balance3 = newSlice(maxFeeIdxs)
	zki.EthAddr3 = newSlice(maxFeeIdxs)
	zki.Siblings3 = make([][]*big.Int, maxFeeIdxs)
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
		zki.ISAccFeeOut[i] = newSlice(maxFeeIdxs)
	}
	zki.ISStateRootFee = newSlice(maxFeeIdxs - 1)
	zki.ISInitStateRootFee = big.NewInt(0)
	zki.ISFinalAccFee = newSlice(maxFeeIdxs - 1)

	return zki
}

// newSlice returns a []*big.Int slice of length n with values initialized at
// 0.
// Is used to initialize all *big.Ints of the ZKInputs data structure, so when
// the transactions are processed and the ZKInputs filled, there is no need to
// set all the elements, and if a transaction does not use a parameter, can be
// leaved as it is in the ZKInputs, as will be 0, so later when using the
// ZKInputs to generate the zkSnark proof there is no 'nil'/'null' values.
func newSlice(n uint32) []*big.Int {
	s := make([]*big.Int, n)
	for i := 0; i < len(s); i++ {
		s[i] = big.NewInt(0)
	}
	return s
}

// HashGlobalData returns the HashGlobalData
func (z ZKInputs) HashGlobalData() (*big.Int, error) {
	b, err := z.ToHashGlobalData()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	h := sha256.New()
	_, err = h.Write(b)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	r := new(big.Int).SetBytes(h.Sum(nil))
	v := r.Mod(r, cryptoConstants.Q)

	return v, nil
}

// ToHashGlobalData returns the data to be hashed in the method HashGlobalData
func (z ZKInputs) ToHashGlobalData() ([]byte, error) {
	var b []byte
	bytesMaxLevels := int(z.Metadata.MaxLevels / 8) //nolint:gomnd
	bytesNLevels := int(z.Metadata.NLevels / 8)     //nolint:gomnd

	// [MAX_NLEVELS bits] oldLastIdx
	oldLastIdx := make([]byte, bytesMaxLevels)
	copy(oldLastIdx, z.OldLastIdx.Bytes())
	b = append(b, SwapEndianness(oldLastIdx)...)

	// [MAX_NLEVELS bits] newLastIdx
	newLastIdx := make([]byte, bytesMaxLevels)
	newLastIdxBytes, err := z.Metadata.NewLastIdxRaw.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(newLastIdx, newLastIdxBytes[len(newLastIdxBytes)-bytesMaxLevels:])
	b = append(b, newLastIdx...)

	// [256 bits] oldStRoot
	oldStateRoot := make([]byte, 32)
	copy(oldStateRoot, z.OldStateRoot.Bytes())
	b = append(b, oldStateRoot...)

	// [256 bits] newStateRoot
	newStateRoot := make([]byte, 32)
	copy(newStateRoot, z.Metadata.NewStateRootRaw.Bytes())
	b = append(b, newStateRoot...)

	// [256 bits] newExitRoot
	newExitRoot := make([]byte, 32)
	copy(newExitRoot, z.Metadata.NewExitRootRaw.Bytes())
	b = append(b, newExitRoot...)

	// [MAX_L1_TX * (2 * MAX_NLEVELS + 480) bits] L1TxsData
	l1TxDataLen := (2*z.Metadata.MaxLevels + 480)
	l1TxsDataLen := (z.Metadata.MaxL1Tx * l1TxDataLen)
	l1TxsData := make([]byte, l1TxsDataLen/8) //nolint:gomnd
	for i := 0; i < len(z.Metadata.L1TxsData); i++ {
		dataLen := int(l1TxDataLen) / 8 //nolint:gomnd
		pos0 := i * dataLen
		pos1 := i*dataLen + dataLen
		copy(l1TxsData[pos0:pos1], z.Metadata.L1TxsData[i])
	}
	b = append(b, l1TxsData...)

	var l1TxsDataAvailability []byte
	for i := 0; i < len(z.Metadata.L1TxsDataAvailability); i++ {
		l1TxsDataAvailability = append(l1TxsDataAvailability, z.Metadata.L1TxsDataAvailability[i]...)
	}
	b = append(b, l1TxsDataAvailability...)

	// [MAX_TX*(2*NLevels + 24) bits] L2TxsData
	var l2TxsData []byte
	l2TxDataLen := 2*z.Metadata.NLevels + 24 //nolint:gomnd
	l2TxsDataLen := (z.Metadata.MaxTx * l2TxDataLen)
	expectedL2TxsDataLen := l2TxsDataLen / 8 //nolint:gomnd
	for i := 0; i < len(z.Metadata.L2TxsData); i++ {
		l2TxsData = append(l2TxsData, z.Metadata.L2TxsData[i]...)
	}
	if len(l2TxsData) > int(expectedL2TxsDataLen) {
		return nil, tracerr.Wrap(fmt.Errorf("len(l2TxsData): %d, expected: %d", len(l2TxsData), expectedL2TxsDataLen))
	}

	b = append(b, l2TxsData...)
	l2TxsPadding := make([]byte, (int(z.Metadata.MaxTx)-len(z.Metadata.L1TxsDataAvailability)-len(z.Metadata.L2TxsData))*int(l2TxDataLen)/8) //nolint:gomnd
	b = append(b, l2TxsPadding...)

	// [NLevels * MAX_TOKENS_FEE bits] feeTxsData
	for i := 0; i < len(z.FeeIdxs); i++ {
		feeIdx := make([]byte, bytesNLevels) //nolint:gomnd
		feeIdxBytes := z.FeeIdxs[i].Bytes()
		copy(feeIdx[len(feeIdx)-len(feeIdxBytes):], feeIdxBytes[:])
		b = append(b, feeIdx...)
	}

	// [16 bits] chainID
	var chainID [2]byte
	binary.BigEndian.PutUint16(chainID[:], z.Metadata.ChainID)
	b = append(b, chainID[:]...)

	// [32 bits] currentNumBatch
	currNumBatchBytes := z.CurrentNumBatch.Bytes()
	var currNumBatch [4]byte
	copy(currNumBatch[4-len(currNumBatchBytes):], currNumBatchBytes)
	b = append(b, currNumBatch[:]...)

	return b, nil
}
