package api

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
)

type testBatch struct {
	ItemID        uint64                    `json:"itemId"`
	BatchNum      common.BatchNum           `json:"batchNum"`
	EthBlockNum   int64                     `json:"ethereumBlockNum"`
	EthBlockHash  ethCommon.Hash            `json:"ethereumBlockHash"`
	Timestamp     time.Time                 `json:"timestamp"`
	ForgerAddr    ethCommon.Address         `json:"forgerAddr"`
	CollectedFees map[common.TokenID]string `json:"collectedFees"`
	TotalFeesUSD  *float64                  `json:"historicTotalCollectedFeesUSD"`
	StateRoot     string                    `json:"stateRoot"`
	NumAccounts   int                       `json:"numAccounts"`
	ExitRoot      string                    `json:"exitRoot"`
	ForgeL1TxsNum *int64                    `json:"forgeL1TransactionsNum"`
	SlotNum       int64                     `json:"slotNum"`
}
type testBatchesResponse struct {
	Batches      []testBatch `json:"batches"`
	PendingItems uint64      `json:"pendingItems"`
}

func (t testBatchesResponse) GetPending() (pendingItems, lastItemID uint64) {
	pendingItems = t.PendingItems
	lastItemID = t.Batches[len(t.Batches)-1].ItemID
	return pendingItems, lastItemID
}

func (t testBatchesResponse) Len() int {
	return len(t.Batches)
}

func (t testBatchesResponse) New() Pendinger { return &testBatchesResponse{} }

type testFullBatch struct {
	Batch testBatch `json:"batch"`
	Txs   []testTx  `json:"transactions"`
}

func genTestBatches(
	blocks []common.Block,
	cBatches []common.Batch,
	txs []testTx,
) ([]testBatch, []testFullBatch) {
	tBatches := []testBatch{}
	for i := 0; i < len(cBatches); i++ {
		block := common.Block{}
		found := false
		for _, b := range blocks {
			if b.Num == cBatches[i].EthBlockNum {
				block = b
				found = true
				break
			}
		}
		if !found {
			panic("block not found")
		}
		collectedFees := make(map[common.TokenID]string)
		for k, v := range cBatches[i].CollectedFees {
			collectedFees[k] = v.String()
		}
		tBatch := testBatch{
			BatchNum:      cBatches[i].BatchNum,
			EthBlockNum:   cBatches[i].EthBlockNum,
			EthBlockHash:  block.Hash,
			Timestamp:     block.Timestamp,
			ForgerAddr:    cBatches[i].ForgerAddr,
			CollectedFees: collectedFees,
			TotalFeesUSD:  cBatches[i].TotalFeesUSD,
			StateRoot:     cBatches[i].StateRoot.String(),
			NumAccounts:   cBatches[i].NumAccounts,
			ExitRoot:      cBatches[i].ExitRoot.String(),
			ForgeL1TxsNum: cBatches[i].ForgeL1TxsNum,
			SlotNum:       cBatches[i].SlotNum,
		}
		tBatches = append(tBatches, tBatch)
	}
	fullBatches := []testFullBatch{}
	for i := 0; i < len(tBatches); i++ {
		forgedTxs := []testTx{}
		for j := 0; j < len(txs); j++ {
			if txs[j].BatchNum != nil && *txs[j].BatchNum == tBatches[i].BatchNum {
				forgedTxs = append(forgedTxs, txs[j])
			}
		}
		fullBatches = append(fullBatches, testFullBatch{
			Batch: tBatches[i],
			Txs:   forgedTxs,
		})
	}
	return tBatches, fullBatches
}

func TestGetBatches(t *testing.T) {
	endpoint := apiURL + "batches"
	fetchedBatches := []testBatch{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testBatchesResponse).Batches); i++ {
			tmp, err := copystructure.Copy(intr.(*testBatchesResponse).Batches[i])
			if err != nil {
				panic(err)
			}
			fetchedBatches = append(fetchedBatches, tmp.(testBatch))
		}
	}
	// Get all (no filters)
	limit := 3
	path := fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	assertBatches(t, tc.batches, fetchedBatches)

	// minBatchNum
	fetchedBatches = []testBatch{}
	limit = 2
	minBatchNum := tc.batches[len(tc.batches)/2].BatchNum
	path = fmt.Sprintf("%s?minBatchNum=%d&limit=%d&fromItem=", endpoint, minBatchNum, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	minBatchNumBatches := []testBatch{}
	for i := 0; i < len(tc.batches); i++ {
		if tc.batches[i].BatchNum > minBatchNum {
			minBatchNumBatches = append(minBatchNumBatches, tc.batches[i])
		}
	}
	assertBatches(t, minBatchNumBatches, fetchedBatches)

	// maxBatchNum
	fetchedBatches = []testBatch{}
	limit = 1
	maxBatchNum := tc.batches[len(tc.batches)/2].BatchNum
	path = fmt.Sprintf("%s?maxBatchNum=%d&limit=%d&fromItem=", endpoint, maxBatchNum, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	maxBatchNumBatches := []testBatch{}
	for i := 0; i < len(tc.batches); i++ {
		if tc.batches[i].BatchNum < maxBatchNum {
			maxBatchNumBatches = append(maxBatchNumBatches, tc.batches[i])
		}
	}
	assertBatches(t, maxBatchNumBatches, fetchedBatches)

	// slotNum
	fetchedBatches = []testBatch{}
	limit = 5
	slotNum := tc.batches[len(tc.batches)/2].SlotNum
	path = fmt.Sprintf("%s?slotNum=%d&limit=%d&fromItem=", endpoint, slotNum, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	slotNumBatches := []testBatch{}
	for i := 0; i < len(tc.batches); i++ {
		if tc.batches[i].SlotNum == slotNum {
			slotNumBatches = append(slotNumBatches, tc.batches[i])
		}
	}
	assertBatches(t, slotNumBatches, fetchedBatches)

	// forgerAddr
	fetchedBatches = []testBatch{}
	limit = 10
	forgerAddr := tc.batches[len(tc.batches)/2].ForgerAddr
	path = fmt.Sprintf("%s?forgerAddr=%s&limit=%d&fromItem=", endpoint, forgerAddr.String(), limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	forgerAddrBatches := []testBatch{}
	for i := 0; i < len(tc.batches); i++ {
		if tc.batches[i].ForgerAddr == forgerAddr {
			forgerAddrBatches = append(forgerAddrBatches, tc.batches[i])
		}
	}
	assertBatches(t, forgerAddrBatches, fetchedBatches)

	// All, in reverse order
	fetchedBatches = []testBatch{}
	limit = 6
	path = fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err = doGoodReqPaginated(path, historydb.OrderDesc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	flippedBatches := []testBatch{}
	for i := len(tc.batches) - 1; i >= 0; i-- {
		flippedBatches = append(flippedBatches, tc.batches[i])
	}
	assertBatches(t, flippedBatches, fetchedBatches)

	// Mixed filters
	fetchedBatches = []testBatch{}
	limit = 1
	maxBatchNum = tc.batches[len(tc.batches)-len(tc.batches)/4].BatchNum
	minBatchNum = tc.batches[len(tc.batches)/4].BatchNum
	path = fmt.Sprintf("%s?minBatchNum=%d&maxBatchNum=%d&limit=%d&fromItem=", endpoint, minBatchNum, maxBatchNum, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testBatchesResponse{}, appendIter)
	assert.NoError(t, err)
	minMaxBatchNumBatches := []testBatch{}
	for i := 0; i < len(tc.batches); i++ {
		if tc.batches[i].BatchNum < maxBatchNum && tc.batches[i].BatchNum > minBatchNum {
			minMaxBatchNumBatches = append(minMaxBatchNumBatches, tc.batches[i])
		}
	}
	assertBatches(t, minMaxBatchNumBatches, fetchedBatches)

	// 400
	// Invalid minBatchNum
	path = fmt.Sprintf("%s?minBatchNum=%d", endpoint, -2)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// Invalid forgerAddr
	path = fmt.Sprintf("%s?forgerAddr=%s", endpoint, "0xG0000001")
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// 404
	path = fmt.Sprintf("%s?slotNum=%d&minBatchNum=%d", endpoint, 1, 25)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
}

func TestGetBatch(t *testing.T) {
	endpoint := apiURL + "batches/"
	for _, batch := range tc.batches {
		fetchedBatch := testBatch{}
		assert.NoError(
			t, doGoodReq(
				"GET",
				endpoint+strconv.Itoa(int(batch.BatchNum)),
				nil, &fetchedBatch,
			),
		)
		assertBatch(t, batch, fetchedBatch)
	}
	// 400
	assert.NoError(t, doBadReq("GET", endpoint+"foo", nil, 400))
	// 404
	assert.NoError(t, doBadReq("GET", endpoint+"99999", nil, 404))
}

func TestGetFullBatch(t *testing.T) {
	endpoint := apiURL + "full-batches/"
	for _, fullBatch := range tc.fullBatches {
		fetchedFullBatch := testFullBatch{}
		assert.NoError(
			t, doGoodReq(
				"GET",
				endpoint+strconv.Itoa(int(fullBatch.Batch.BatchNum)),
				nil, &fetchedFullBatch,
			),
		)
		assertBatch(t, fullBatch.Batch, fetchedFullBatch.Batch)
		assertTxs(t, fullBatch.Txs, fetchedFullBatch.Txs)
	}
	// 400
	assert.NoError(t, doBadReq("GET", endpoint+"foo", nil, 400))
	// 404
	assert.NoError(t, doBadReq("GET", endpoint+"99999", nil, 404))
}

func assertBatches(t *testing.T, expected, actual []testBatch) {
	assert.Equal(t, len(expected), len(actual))
	for i := 0; i < len(expected); i++ {
		assertBatch(t, expected[i], actual[i])
	}
}

func assertBatch(t *testing.T, expected, actual testBatch) {
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp
	actual.ItemID = expected.ItemID
	assert.Equal(t, expected, actual)
}
