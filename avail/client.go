package avail

import (
	"fmt"
	"math/big"
	"strings"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
)

var appID int

type Client struct {
	api        *gsrpc.SubstrateAPI
	seedPhrase string
}

func NewClient(url, seedPhrase string) (*Client, error) {
	api, err := gsrpc.NewSubstrateAPI(url)
	if err != nil {
		return nil, err
	}
	return &Client{api: api, seedPhrase: seedPhrase}, nil
}

func (cl *Client) GetLastBlock() (*common.BlockAvail, error) {
	header, err := cl.api.RPC.Chain.GetHeaderLatest()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	blockAvail := &common.BlockAvail{
		Num:        uint32(header.Number),
		Hash:       header.StateRoot.Hex(),
		ParentHash: header.ParentHash.Hex(),
	}

	return blockAvail, nil
}

func (cl *Client) GetBlockByNumber(num uint64) (*common.BlockAvail, error) {
	hash, err := cl.api.RPC.Chain.GetBlockHash(num)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	block, err := cl.api.RPC.Chain.GetBlock(hash)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	nLevels := int64(32)
	lenL1L2TxsBytes := int((nLevels/8)*2 + common.Float40BytesLength + 1) //nolint:gomnd
	var (
		l2TxsData []byte
		stateRoot string
	)

	for _, ext := range block.Block.Extrinsics {
		methodArgs := string(ext.Method.Args)
		if strings.Contains(methodArgs, "root") {
			args := strings.Split(methodArgs, "root")
			l2TxsData = []byte(args[1])
			stateRoot = args[0]
		}
	}

	var l2txs []common.L2Tx
	numTxsL2 := len(l2TxsData) / lenL1L2TxsBytes
	for i := 0; i < numTxsL2; i++ {
		l2tx, err := common.L2TxFromBytesDataAvailability(l2TxsData[i*lenL1L2TxsBytes:(i+1)*lenL1L2TxsBytes], int(nLevels))
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l2txs = append(l2txs, *l2tx)
	}

	blockAvail := &common.BlockAvail{
		Num:        uint32(block.Block.Header.Number),
		Hash:       hash.Hex(),
		ParentHash: block.Block.Header.ParentHash.Hex(),
		L2Txs:      l2txs,
		StateRoot:  stateRoot,
	}
	return blockAvail, nil
}

func (cl *Client) SendTxs(stateRoot *big.Int, l1UserTxs, l1CoordTxs []common.L1Tx, l2Txs []common.L2Tx) error {
	nLevels := 32
	// L1L2TxData
	var l1l2TxData []byte
	for i := 0; i < len(l1UserTxs); i++ {
		l1User := l1UserTxs[i]
		bytesl1User, err := l1User.BytesDataAvailability(uint32(nLevels))
		if err != nil {
			return tracerr.Wrap(err)
		}
		l1l2TxData = append(l1l2TxData, bytesl1User[:]...)
	}
	for i := 0; i < len(l1CoordTxs); i++ {
		l1Coord := l1CoordTxs[i]
		bytesl1Coord, err := l1Coord.BytesDataAvailability(uint32(nLevels))
		if err != nil {
			return tracerr.Wrap(err)
		}
		l1l2TxData = append(l1l2TxData, bytesl1Coord[:]...)
	}
	for i := 0; i < len(l2Txs); i++ {
		l2 := l2Txs[i]
		bytesl2, err := l2.BytesDataAvailability(uint32(nLevels))
		if err != nil {
			return tracerr.Wrap(err)
		}
		l1l2TxData = append(l1l2TxData, bytesl2[:]...)
	}

	_, err := submitData(cl.api, stateRoot.String()+"root"+string(l1l2TxData), cl.seedPhrase, 0)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// Submit data sends the extrinsic data to Substrate
// seed is used for keyring generation, 42 is the network number for Substrate
func submitData(api *gsrpc.SubstrateAPI, data string, seed string, appID int) (types.Hash, error) {

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return types.Hash{}, err
	}

	c, err := types.NewCall(meta, "DataAvailability.submit_data", types.NewBytes([]byte(data)))
	if err != nil {
		return types.Hash{}, fmt.Errorf("error creating new call: %s", err)
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error getting genesis hash: %s", err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return types.Hash{}, fmt.Errorf("error retrieveing runtime version: %s", err)
	}

	keyringPair, err := signature.KeyringPairFromSecret(seed, 42)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error creating keyring pair: %s", err)
	}

	// if testing locally with Alice account, use signature.TestKeyringPairAlice.PublicKey as last param
	key, err := types.CreateStorageKey(meta, "System", "Account", keyringPair.PublicKey)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error createStorageKey: %s", err)
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return types.Hash{}, fmt.Errorf("error GetStorageLatest: %s", err)
	}

	nonce := uint32(accountInfo.Nonce)
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		AppID:              types.NewU32(uint32(appID)),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account
	err = ext.Sign(keyringPair, o)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error signing tx: %s", err.Error())
	}

	// Send the extrinsic
	hash, err := api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error submitting extrinsic: %s", err.Error())
	}

	return hash, nil
}
