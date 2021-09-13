package historydb

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/tracerr"
	"github.com/russross/meddler"
)

// Period represents a time period in ethereum
type Period struct {
	SlotNum       int64     `json:"slotNum"`
	FromBlock     int64     `json:"fromBlock"`
	ToBlock       int64     `json:"toBlock"`
	FromTimestamp time.Time `json:"fromTimestamp"`
	ToTimestamp   time.Time `json:"toTimestamp"`
}

// NextForgerAPI represents the next forger exposed via the API
type NextForgerAPI struct {
	Coordinator CoordinatorAPI `json:"coordinator"`
	Period      Period         `json:"period"`
}

// NetworkAPI is the network state exposed via the API
type NetworkAPI struct {
	LastEthBlock  int64           `json:"lastEthereumBlock"`
	LastSyncBlock int64           `json:"lastSynchedBlock"`
	LastBatch     *BatchAPI       `json:"lastBatch"`
	CurrentSlot   int64           `json:"currentSlot"`
	NextForgers   []NextForgerAPI `json:"nextForgers"`
	PendingL1Txs  int             `json:"pendingL1Transactions"`
}

// NodePublicInfo is the configuration and metrics of the node that is exposed via API
type NodePublicInfo struct {
	// ForgeDelay in seconds
	ForgeDelay float64 `json:"forgeDelay"`
	// PoolLoad amount of transactions in the pool
	PoolLoad int64 `json:"poolLoad"`
}

// StateAPI is an object representing the node and network state exposed via the API
type StateAPI struct {
	NodePublicInfo    NodePublicInfo           `json:"node"`
	Network           NetworkAPI               `json:"network"`
	Metrics           MetricsAPI               `json:"metrics"`
	Rollup            RollupVariablesAPI       `json:"rollup"`
	Auction           AuctionVariablesAPI      `json:"auction"`
	WithdrawalDelayer common.WDelayerVariables `json:"withdrawalDelayer"`
	RecommendedFee    common.RecommendedFee    `json:"recommendedFee"`
}

// Constants contains network constants
type Constants struct {
	common.SCConsts
	ChainID       uint16
	HermezAddress ethCommon.Address
}

// NodeConfig contains the node config exposed in the API
type NodeConfig struct {
	MaxPoolTxs uint32
	MinFeeUSD  float64
	MaxFeeUSD  float64
	ForgeDelay float64
}

// NodeInfo contains information about he node used when serving the API
type NodeInfo struct {
	ItemID     int         `meddler:"item_id,pk"`
	StateAPI   *StateAPI   `meddler:"state,json"`
	NodeConfig *NodeConfig `meddler:"config,json"`
	Constants  *Constants  `meddler:"constants,json"`
}

// GetNodeInfo returns the NodeInfo
func (hdb *HistoryDB) GetNodeInfo() (*NodeInfo, error) {
	ni := &NodeInfo{}
	err := meddler.QueryRow(
		hdb.dbRead, ni, `SELECT * FROM node_info WHERE item_id = 1;`,
	)
	return ni, tracerr.Wrap(err)
}

// GetConstants returns the Constats
func (hdb *HistoryDB) GetConstants() (*Constants, error) {
	var nodeInfo NodeInfo
	err := meddler.QueryRow(
		hdb.dbRead, &nodeInfo,
		"SELECT constants FROM node_info WHERE item_id = 1;",
	)
	return nodeInfo.Constants, tracerr.Wrap(err)
}

// SetConstants sets the Constants
func (hdb *HistoryDB) SetConstants(constants *Constants) error {
	_constants := struct {
		Constants *Constants `meddler:"constants,json"`
	}{constants}
	values, err := meddler.Default.Values(&_constants, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	_, err = hdb.dbWrite.Exec(
		"UPDATE node_info SET constants = $1 WHERE item_id = 1;",
		values[0],
	)
	return tracerr.Wrap(err)
}

// GetStateInternalAPI returns the StateAPI
func (hdb *HistoryDB) GetStateInternalAPI() (*StateAPI, error) {
	return hdb.getStateAPI(hdb.dbRead)
}

func (hdb *HistoryDB) getStateAPI(d meddler.DB) (*StateAPI, error) {
	var nodeInfo NodeInfo
	err := meddler.QueryRow(
		d, &nodeInfo,
		"SELECT state FROM node_info WHERE item_id = 1;",
	)
	return nodeInfo.StateAPI, tracerr.Wrap(err)
}

// SetStateInternalAPI sets the StateAPI
func (hdb *HistoryDB) SetStateInternalAPI(stateAPI *StateAPI) error {
	if stateAPI.Network.LastBatch != nil {
		stateAPI.Network.LastBatch.CollectedFeesAPI =
			apitypes.NewCollectedFeesAPI(stateAPI.Network.LastBatch.CollectedFeesDB)
	}
	_stateAPI := struct {
		StateAPI *StateAPI `meddler:"state,json"`
	}{stateAPI}
	values, err := meddler.Default.Values(&_stateAPI, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	_, err = hdb.dbWrite.Exec(
		"UPDATE node_info SET state = $1 WHERE item_id = 1;",
		values[0],
	)
	return tracerr.Wrap(err)
}

// GetNodeConfig returns the NodeConfig
func (hdb *HistoryDB) GetNodeConfig() (*NodeConfig, error) {
	var nodeInfo NodeInfo
	err := meddler.QueryRow(
		hdb.dbRead, &nodeInfo,
		"SELECT config FROM node_info WHERE item_id = 1;",
	)
	return nodeInfo.NodeConfig, tracerr.Wrap(err)
}

// SetNodeConfig sets the NodeConfig
func (hdb *HistoryDB) SetNodeConfig(nodeConfig *NodeConfig) error {
	_nodeConfig := struct {
		NodeConfig *NodeConfig `meddler:"config,json"`
	}{nodeConfig}
	values, err := meddler.Default.Values(&_nodeConfig, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	_, err = hdb.dbWrite.Exec(
		"UPDATE node_info SET config = $1 WHERE item_id = 1;",
		values[0],
	)
	return tracerr.Wrap(err)
}
