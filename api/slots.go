package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

// SlotAPI is a representation of a slot information
type SlotAPI struct {
	ItemID      uint64            `json:"itemId"`
	SlotNum     int64             `json:"slotNum"`
	FirstBlock  int64             `json:"firstBlock"`
	LastBlock   int64             `json:"lastBlock"`
	OpenAuction bool              `json:"openAuction"`
	WinnerBid   *historydb.BidAPI `json:"bestBid"`
}

func (a *API) getFirstLastBlock(slotNum int64) (int64, int64) {
	genesisBlock := a.config.AuctionConstants.GenesisBlockNum
	blocksPerSlot := int64(a.config.AuctionConstants.BlocksPerSlot)
	firstBlock := slotNum*blocksPerSlot + genesisBlock
	lastBlock := (slotNum+1)*blocksPerSlot + genesisBlock - 1
	return firstBlock, lastBlock
}

func (a *API) getCurrentSlot(currentBlock int64) int64 {
	genesisBlock := a.config.AuctionConstants.GenesisBlockNum
	blocksPerSlot := int64(a.config.AuctionConstants.BlocksPerSlot)
	currentSlot := (currentBlock - genesisBlock) / blocksPerSlot
	return currentSlot
}

func (a *API) isOpenAuction(currentBlock, slotNum int64, auctionVars common.AuctionVariables) bool {
	currentSlot := a.getCurrentSlot(currentBlock)
	closedAuctionSlots := currentSlot + int64(auctionVars.ClosedAuctionSlots)
	openAuctionSlots := int64(auctionVars.OpenAuctionSlots)
	if slotNum > closedAuctionSlots && slotNum <= (closedAuctionSlots+openAuctionSlots) {
		return true
	}
	return false
}

func (a *API) newSlotAPI(slotNum, currentBlockNum int64, bid *historydb.BidAPI, auctionVars *common.AuctionVariables) SlotAPI {
	firstBlock, lastBlock := a.getFirstLastBlock(slotNum)
	openAuction := a.isOpenAuction(currentBlockNum, slotNum, *auctionVars)
	slot := SlotAPI{
		ItemID:      uint64(slotNum),
		SlotNum:     slotNum,
		FirstBlock:  firstBlock,
		LastBlock:   lastBlock,
		OpenAuction: openAuction,
		WinnerBid:   bid,
	}
	return slot
}

func (a *API) newSlotsAPIFromWinnerBids(fromItem *uint, order string, bids []historydb.BidAPI, currentBlockNum int64, auctionVars *common.AuctionVariables) (slots []SlotAPI) {
	for i := range bids {
		slotNum := bids[i].SlotNum
		slot := a.newSlotAPI(slotNum, currentBlockNum, &bids[i], auctionVars)
		if order == db.OrderAsc {
			if fromItem == nil || slot.ItemID >= uint64(*fromItem) {
				slots = append(slots, slot)
			}
		} else {
			if fromItem == nil || slot.ItemID <= uint64(*fromItem) {
				slots = append(slots, slot)
			}
		}
	}
	return slots
}

func (a *API) addEmptySlot(slots []SlotAPI, slotNum int64, currentBlockNum int64, auctionVars *common.AuctionVariables, fromItem *uint, order string) ([]SlotAPI, error) {
	emptySlot := a.newSlotAPI(slotNum, currentBlockNum, nil, auctionVars)
	if order == db.OrderAsc {
		if fromItem == nil || emptySlot.ItemID >= uint64(*fromItem) {
			slots = append(slots, emptySlot)
		}
	} else {
		if fromItem == nil || emptySlot.ItemID <= uint64(*fromItem) {
			slots = append([]SlotAPI{emptySlot}, slots...)
		}
	}
	return slots, nil
}

func (a *API) getSlot(c *gin.Context) {
	slotNumUint, err := parsers.ParseSlotFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	currentBlock, err := a.historyDB.GetLastBlockAPI()
	if err != nil {
		retBadReq(err, c)
		return
	}
	auctionVars, err := a.historyDB.GetAuctionVarsAPI()
	if err != nil {
		retBadReq(err, c)
		return
	}

	slotNum := int64(*slotNumUint)
	bid, err := a.historyDB.GetBestBidAPI(&slotNum)
	if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
		retSQLErr(err, c)
		return
	}

	var slot SlotAPI
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		slot = a.newSlotAPI(slotNum, currentBlock.Num, nil, auctionVars)
	} else {
		slot = a.newSlotAPI(bid.SlotNum, currentBlock.Num, &bid, auctionVars)
	}

	// JSON response
	c.JSON(http.StatusOK, slot)
}

func getLimits(
	minSlotNum, maxSlotNum int64, fromItem, limit *uint, order string,
) (minLimit, maxLimit int64, pendingItems uint64) {
	if order == db.OrderAsc {
		if fromItem != nil && int64(*fromItem) > minSlotNum {
			minLimit = int64(*fromItem)
		} else {
			minLimit = minSlotNum
		}
		if limit != nil && (minLimit+int64(*limit-1)) < maxSlotNum {
			maxLimit = minLimit + int64(*limit-1)
		} else {
			maxLimit = maxSlotNum
		}
		pendingItems = uint64(maxSlotNum - maxLimit)
	} else {
		if fromItem != nil && int64(*fromItem) < maxSlotNum {
			maxLimit = int64(*fromItem)
		} else {
			maxLimit = maxSlotNum
		}
		if limit != nil && (maxLimit-int64(*limit-1)) < minSlotNum {
			minLimit = minSlotNum
		} else {
			minLimit = maxLimit - int64(*limit-1)
		}
		pendingItems = uint64(-(minSlotNum - minLimit))
	}
	return minLimit, maxLimit, pendingItems
}

func getLimitsWithAddr(minSlotNum, maxSlotNum *int64, fromItem, limit *uint, order string) (int64, int64) {
	var minLim, maxLim int64
	if fromItem != nil {
		if order == db.OrderAsc {
			maxLim = *maxSlotNum
			if int64(*fromItem) > *minSlotNum {
				minLim = int64(*fromItem)
			} else {
				minLim = *minSlotNum
			}
		} else {
			minLim = *minSlotNum
			if int64(*fromItem) < *maxSlotNum {
				maxLim = int64(*fromItem)
			} else {
				maxLim = *maxSlotNum
			}
		}
	} else {
		maxLim = *maxSlotNum
		minLim = *minSlotNum
	}
	return minLim, maxLim
}

func (a *API) getSlots(c *gin.Context) {
	var slots []SlotAPI
	minSlotNumDflt := int64(0)

	// Get filters
	filters, err := parsers.ParseSlotsFilters(c, a.validate)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	currentBlock, err := a.historyDB.GetLastBlockAPI()
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrFailedToGetCurrentBlockCode,
			Type: ErrFailedToGetCurrentBlockType,
		}, c)
		return
	}
	auctionVars, err := a.historyDB.GetAuctionVarsAPI()
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrFailedToGetAuctionVarsCode,
			Type: ErrFailedToGetAuctionVarsType,
		}, c)
		return
	}

	// Check filters
	if filters.FinishedAuction != nil && *filters.FinishedAuction {
		currentSlot := a.getCurrentSlot(currentBlock.Num)
		closedAuctionSlots := currentSlot + int64(auctionVars.ClosedAuctionSlots)
		if filters.MaxSlotNum == nil || closedAuctionSlots < *filters.MaxSlotNum {
			filters.MaxSlotNum = &closedAuctionSlots
		}
	}
	if filters.MinSlotNum == nil {
		filters.MinSlotNum = &minSlotNumDflt
	}

	// Get bids and pagination according to filters
	var slotMinLim, slotMaxLim int64
	var bids []historydb.BidAPI
	var pendingItems uint64
	if filters.BidderAddr == nil {
		slotMinLim, slotMaxLim, pendingItems = getLimits(*filters.MinSlotNum, *filters.MaxSlotNum, filters.FromItem, filters.Limit, filters.Order)
		// Get best bids in range maxSlotNum - minSlotNum
		bids, _, err = a.historyDB.GetBestBidsAPI(historydb.GetBestBidsAPIRequest{
			MinSlotNum: &slotMinLim,
			MaxSlotNum: &slotMaxLim,
			BidderAddr: filters.BidderAddr,
			Limit:      nil,
			Order:      filters.Order,
		})
		if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
			retSQLErr(err, c)
			return
		}
	} else {
		slotMinLim, slotMaxLim = getLimitsWithAddr(filters.MinSlotNum, filters.MaxSlotNum, filters.FromItem, filters.Limit, filters.Order)
		bids, pendingItems, err = a.historyDB.GetBestBidsAPI(historydb.GetBestBidsAPIRequest{
			MinSlotNum: &slotMinLim,
			MaxSlotNum: &slotMaxLim,
			BidderAddr: filters.BidderAddr,
			Limit:      filters.Limit,
			Order:      filters.Order,
		})
		if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
			retSQLErr(err, c)
			return
		}
	}

	// Build the slot information with previous bids
	var slotsBids []SlotAPI
	if len(bids) > 0 {
		slotsBids = a.newSlotsAPIFromWinnerBids(filters.FromItem, filters.Order, bids, currentBlock.Num, auctionVars)
	}

	// Build the other slots
	if filters.BidderAddr == nil {
		// Build hte information of the slots with bids or not
		for i := slotMinLim; i <= slotMaxLim; i++ {
			found := false
			for j := range slotsBids {
				if slotsBids[j].SlotNum == i {
					found = true
					if filters.Order == db.OrderAsc {
						if filters.FromItem == nil || slotsBids[j].ItemID >= uint64(*filters.FromItem) {
							slots = append(slots, slotsBids[j])
						}
					} else {
						if filters.FromItem == nil || slotsBids[j].ItemID <= uint64(*filters.FromItem) {
							slots = append([]SlotAPI{slotsBids[j]}, slots...)
						}
					}
					break
				}
			}
			if !found {
				slots, err = a.addEmptySlot(slots, i, currentBlock.Num, auctionVars, filters.FromItem, filters.Order)
				if err != nil {
					retBadReq(&apiError{
						Err:  err,
						Code: ErrFailedToAddEmptySlotCode,
						Type: ErrFailedToAddEmptySlotType,
					}, c)
					return
				}
			}
		}
	} else if len(slotsBids) > 0 {
		slots = slotsBids
	}

	if len(slots) == 0 {
		retSQLErr(sql.ErrNoRows, c)
		return
	}

	// Build successful response
	type slotsResponse struct {
		Slots        []SlotAPI `json:"slots"`
		PendingItems uint64    `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &slotsResponse{
		Slots:        slots,
		PendingItems: pendingItems,
	})
}
