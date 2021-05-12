package api

import (
	"fmt"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
)

type testBid struct {
	ItemID      uint64            `json:"itemId"`
	SlotNum     int64             `json:"slotNum"`
	BidValue    string            `json:"bidValue"`
	EthBlockNum int64             `json:"ethereumBlockNum"`
	Bidder      ethCommon.Address `json:"bidderAddr"`
	Forger      ethCommon.Address `json:"forgerAddr"`
	URL         string            `json:"URL"`
	Timestamp   time.Time         `json:"timestamp"`
}

type testBidsResponse struct {
	Bids         []testBid `json:"bids"`
	PendingItems uint64    `json:"pendingItems"`
}

func (t testBidsResponse) GetPending() (pendingItems, lastItemID uint64) {
	if len(t.Bids) == 0 {
		return 0, 0
	}
	pendingItems = t.PendingItems
	lastItemID = t.Bids[len(t.Bids)-1].ItemID
	return pendingItems, lastItemID
}

func (t testBidsResponse) Len() int {
	return len(t.Bids)
}

func (t testBidsResponse) New() Pendinger { return &testBidsResponse{} }

func genTestBids(blocks []common.Block, coordinators []historydb.CoordinatorAPI, bids []common.Bid) []testBid {
	tBids := []testBid{}
	for _, bid := range bids {
		block := getBlockByNum(bid.EthBlockNum, blocks)
		coordinator := getCoordinatorByBidder(bid.Bidder, coordinators)
		tBid := testBid{
			SlotNum:     bid.SlotNum,
			BidValue:    bid.BidValue.String(),
			EthBlockNum: bid.EthBlockNum,
			Bidder:      bid.Bidder,
			Forger:      coordinator.Forger,
			URL:         coordinator.URL,
			Timestamp:   block.Timestamp,
		}
		tBids = append(tBids, tBid)
	}
	return tBids
}

func TestGetBids(t *testing.T) {
	endpoint := apiURL + "bids"
	fetchedBids := []testBid{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testBidsResponse).Bids); i++ {
			tmp, err := copystructure.Copy(intr.(*testBidsResponse).Bids[i])
			if err != nil {
				panic(err)
			}
			fetchedBids = append(fetchedBids, tmp.(testBid))
		}
	}

	limit := 3
	// bidderAddress
	fetchedBids = []testBid{}
	bidderAddress := tc.bids[3].Bidder
	path := fmt.Sprintf("%s?bidderAddr=%s&limit=%d", endpoint, bidderAddress.String(), limit)
	err := doGoodReqPaginated(path, db.OrderAsc, &testBidsResponse{}, appendIter)
	assert.NoError(t, err)
	bidderAddrBids := []testBid{}
	for i := 0; i < len(tc.bids); i++ {
		if tc.bids[i].Bidder == bidderAddress {
			bidderAddrBids = append(bidderAddrBids, tc.bids[i])
		}
	}
	assertBids(t, bidderAddrBids, fetchedBids)

	// slotNum
	fetchedBids = []testBid{}
	slotNum := tc.bids[3].SlotNum
	path = fmt.Sprintf("%s?slotNum=%d&limit=%d", endpoint, slotNum, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testBidsResponse{}, appendIter)
	assert.NoError(t, err)
	slotNumBids := []testBid{}
	for i := 0; i < len(tc.bids); i++ {
		if tc.bids[i].SlotNum == slotNum {
			slotNumBids = append(slotNumBids, tc.bids[i])
		}
	}
	assertBids(t, slotNumBids, fetchedBids)

	// slotNum, in reverse order
	fetchedBids = []testBid{}
	path = fmt.Sprintf("%s?slotNum=%d&limit=%d", endpoint, slotNum, limit)
	err = doGoodReqPaginated(path, db.OrderDesc, &testBidsResponse{}, appendIter)
	assert.NoError(t, err)
	flippedBids := []testBid{}
	for i := len(slotNumBids) - 1; i >= 0; i-- {
		flippedBids = append(flippedBids, slotNumBids[i])
	}
	assertBids(t, flippedBids, fetchedBids)

	// Mixed filters
	fetchedBids = []testBid{}
	bidderAddress = tc.bids[1].Bidder
	slotNum = tc.bids[1].SlotNum
	path = fmt.Sprintf("%s?bidderAddr=%s&slotNum=%d&limit=%d", endpoint, bidderAddress.String(), slotNum, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testBidsResponse{}, appendIter)
	assert.NoError(t, err)
	slotNumBidderAddrBids := []testBid{}
	for i := 0; i < len(tc.bids); i++ {
		if tc.bids[i].Bidder == bidderAddress && tc.bids[i].SlotNum == slotNum {
			slotNumBidderAddrBids = append(slotNumBidderAddrBids, tc.bids[i])
		}
	}
	assertBids(t, slotNumBidderAddrBids, fetchedBids)

	// Empty array
	fetchedBids = []testBid{}
	path = fmt.Sprintf("%s?slotNum=%d&bidderAddr=%s", endpoint, 5, tc.bids[1].Bidder.String())
	err = doGoodReqPaginated(path, db.OrderAsc, &testBidsResponse{}, appendIter)
	assert.NoError(t, err)
	assertBids(t, []testBid{}, fetchedBids)

	// 400
	// No filters
	path = fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// Invalid slotNum
	path = fmt.Sprintf("%s?slotNum=%d", endpoint, -2)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// Invalid bidderAddress
	path = fmt.Sprintf("%s?bidderAddr=%s", endpoint, "0xG0000001")
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
}

func assertBids(t *testing.T, expected, actual []testBid) {
	assert.Equal(t, len(expected), len(actual))
	for i := 0; i < len(expected); i++ {
		assertBid(t, expected[i], actual[i])
	}
}

func assertBid(t *testing.T, expected, actual testBid) {
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp
	actual.ItemID = expected.ItemID
	assert.Equal(t, expected, actual)
}
