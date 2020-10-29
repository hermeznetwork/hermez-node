package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

// SlotAPI is a repesentation of a slot information
type SlotAPI struct {
	ItemID      int               `json:"itemId"`
	SlotNum     int64             `json:"slotNum"`
	FirstBlock  int64             `json:"firstBlock"`
	LastBlock   int64             `json:"lastBlock"`
	OpenAuction bool              `json:"openAuction"`
	WinnerBid   *historydb.BidAPI `json:"winnerBid"`
	TotalItems  int               `json:"-"`
	FirstItem   int               `json:"-"`
	LastItem    int               `json:"-"`
}

func getFirstLastBlock(slotNum int64) (int64, int64) {
	genesisBlock := cg.AuctionConstants.GenesisBlockNum
	blocksPerSlot := int64(cg.AuctionConstants.BlocksPerSlot)
	firstBlock := slotNum*blocksPerSlot + genesisBlock
	lastBlock := (slotNum+1)*blocksPerSlot + genesisBlock - 1
	return firstBlock, lastBlock
}

func getCurrentSlot(currentBlock int64) int64 {
	genesisBlock := cg.AuctionConstants.GenesisBlockNum
	blocksPerSlot := int64(cg.AuctionConstants.BlocksPerSlot)
	currentSlot := (currentBlock - genesisBlock) / blocksPerSlot
	return currentSlot
}

func isOpenAuction(currentBlock, slotNum int64, auctionVars common.AuctionVariables) bool {
	currentSlot := getCurrentSlot(currentBlock)
	closedAuctionSlots := currentSlot + int64(auctionVars.ClosedAuctionSlots)
	openAuctionSlots := int64(auctionVars.OpenAuctionSlots)
	if slotNum > closedAuctionSlots && slotNum <= (closedAuctionSlots+openAuctionSlots) {
		return true
	}
	return false
}

func getPagination(totalItems int, minSlotNum, maxSlotNum *int64) *db.Pagination {
	// itemID is slotNum
	firstItem := *minSlotNum
	lastItem := *maxSlotNum
	pagination := &db.Pagination{
		TotalItems: int(totalItems),
		FirstItem:  int(firstItem),
		LastItem:   int(lastItem),
	}
	return pagination
}

func newSlotAPI(slotNum, currentBlockNum int64, bid *historydb.BidAPI, auctionVars *common.AuctionVariables) SlotAPI {
	firstBlock, lastBlock := getFirstLastBlock(slotNum)
	openAuction := isOpenAuction(currentBlockNum, slotNum, *auctionVars)
	slot := SlotAPI{
		ItemID:      int(slotNum),
		SlotNum:     slotNum,
		FirstBlock:  firstBlock,
		LastBlock:   lastBlock,
		OpenAuction: openAuction,
		WinnerBid:   bid,
	}
	return slot
}

func newSlotsAPIFromWinnerBids(fromItem *uint, order string, bids []historydb.BidAPI, currentBlockNum int64, auctionVars *common.AuctionVariables) (slots []SlotAPI) {
	for i := range bids {
		slotNum := bids[i].SlotNum
		slot := newSlotAPI(slotNum, currentBlockNum, &bids[i], auctionVars)
		if order == historydb.OrderAsc {
			if slot.ItemID >= int(*fromItem) {
				slots = append(slots, slot)
			}
		} else {
			if slot.ItemID <= int(*fromItem) {
				slots = append(slots, slot)
			}
		}
	}
	return slots
}

func addEmptySlot(slots []SlotAPI, slotNum int64, currentBlockNum int64, auctionVars *common.AuctionVariables, fromItem *uint, order string) ([]SlotAPI, error) {
	emptySlot := newSlotAPI(slotNum, currentBlockNum, nil, auctionVars)
	if order == historydb.OrderAsc {
		if emptySlot.ItemID >= int(*fromItem) {
			slots = append(slots, emptySlot)
		}
	} else {
		if emptySlot.ItemID <= int(*fromItem) {
			slots = append([]SlotAPI{emptySlot}, slots...)
		}
	}
	return slots, nil
}

func getSlot(c *gin.Context) {
	slotNumUint, err := parseParamUint("slotNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	currentBlock, err := h.GetLastBlock()
	if err != nil {
		retBadReq(err, c)
		return
	}
	auctionVars, err := h.GetAuctionVars()
	if err != nil {
		retBadReq(err, c)
		return
	}

	slotNum := int64(*slotNumUint)
	bid, err := h.GetBestBidAPI(&slotNum)
	if err != nil && err != sql.ErrNoRows {
		retSQLErr(err, c)
		return
	}

	var slot SlotAPI
	if err == sql.ErrNoRows {
		slot = newSlotAPI(slotNum, currentBlock.EthBlockNum, nil, auctionVars)
	} else {
		slot = newSlotAPI(bid.SlotNum, currentBlock.EthBlockNum, &bid, auctionVars)
	}

	// JSON response
	c.JSON(http.StatusOK, slot)
}

func getLimits(minSlotNum, maxSlotNum *int64, fromItem, limit *uint, order string) (int64, int64) {
	var minLim, maxLim int64
	if fromItem != nil {
		if order == historydb.OrderAsc {
			if int64(*fromItem) > *minSlotNum {
				minLim = int64(*fromItem)
			} else {
				minLim = *minSlotNum
			}
			if (minLim + int64(*limit-1)) < *maxSlotNum {
				maxLim = minLim + int64(*limit-1)
			} else {
				maxLim = *maxSlotNum
			}
		} else {
			if int64(*fromItem) < *maxSlotNum {
				maxLim = int64(*fromItem)
			} else {
				maxLim = *maxSlotNum
			}
			if (maxLim - int64(*limit-1)) < *minSlotNum {
				minLim = *minSlotNum
			} else {
				minLim = maxLim - int64(*limit-1)
			}
		}
	}
	return minLim, maxLim
}

func getLimitsWithAddr(minSlotNum, maxSlotNum *int64, fromItem, limit *uint, order string) (int64, int64) {
	var minLim, maxLim int64
	if fromItem != nil {
		if order == historydb.OrderAsc {
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
	}
	return minLim, maxLim
}

func getSlots(c *gin.Context) {
	var slots []SlotAPI
	minSlotNumDflt := int64(0)

	// Get filters
	minSlotNum, maxSlotNum, wonByEthereumAddress, finishedAuction, err := parseSlotFilters(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	currentBlock, err := h.GetLastBlock()
	if err != nil {
		retBadReq(err, c)
		return
	}
	auctionVars, err := h.GetAuctionVars()
	if err != nil {
		retBadReq(err, c)
		return
	}

	// Check filters
	if maxSlotNum == nil && finishedAuction == nil {
		retBadReq(errors.New("It is necessary to add maxSlotNum filter"), c)
		return
	} else if finishedAuction != nil {
		if maxSlotNum == nil && !*finishedAuction {
			retBadReq(errors.New("It is necessary to add maxSlotNum filter"), c)
			return
		} else if *finishedAuction {
			currentBlock, err := h.GetLastBlock()
			if err != nil {
				retBadReq(err, c)
				return
			}
			currentSlot := getCurrentSlot(currentBlock.EthBlockNum)
			auctionVars, err := h.GetAuctionVars()
			if err != nil {
				retBadReq(err, c)
				return
			}
			closedAuctionSlots := currentSlot + int64(auctionVars.ClosedAuctionSlots)
			if maxSlotNum == nil {
				maxSlotNum = &closedAuctionSlots
			} else if closedAuctionSlots < *maxSlotNum {
				maxSlotNum = &closedAuctionSlots
			}
		}
	} else if maxSlotNum != nil && minSlotNum != nil {
		if *minSlotNum > *maxSlotNum {
			retBadReq(errors.New("It is necessary to add valid filter (minSlotNum <= maxSlotNum)"), c)
			return
		}
	}
	if minSlotNum == nil {
		minSlotNum = &minSlotNumDflt
	}

	// Get bids and pagination according to filters
	var slotMinLim, slotMaxLim int64
	var bids []historydb.BidAPI
	var pag *db.Pagination
	totalItems := 0
	if wonByEthereumAddress == nil {
		slotMinLim, slotMaxLim = getLimits(minSlotNum, maxSlotNum, fromItem, limit, order)
		// Get best bids in range maxSlotNum - minSlotNum
		bids, _, err = h.GetBestBidsAPI(&slotMinLim, &slotMaxLim, wonByEthereumAddress, nil, order)
		if err != nil && err != sql.ErrNoRows {
			retSQLErr(err, c)
			return
		}
		totalItems = int(*maxSlotNum) - int(*minSlotNum) + 1
	} else {
		slotMinLim, slotMaxLim = getLimitsWithAddr(minSlotNum, maxSlotNum, fromItem, limit, order)
		bids, pag, err = h.GetBestBidsAPI(&slotMinLim, &slotMaxLim, wonByEthereumAddress, limit, order)
		if err != nil && err != sql.ErrNoRows {
			retSQLErr(err, c)
			return
		}
		if len(bids) > 0 {
			totalItems = pag.TotalItems
			*maxSlotNum = int64(pag.LastItem)
			*minSlotNum = int64(pag.FirstItem)
		}
	}

	// Build the slot information with previous bids
	var slotsBids []SlotAPI
	if len(bids) > 0 {
		slotsBids = newSlotsAPIFromWinnerBids(fromItem, order, bids, currentBlock.EthBlockNum, auctionVars)
		if err != nil {
			retBadReq(err, c)
			return
		}
	}

	// Build the other slots
	if wonByEthereumAddress == nil {
		// Build hte information of the slots with bids or not
		for i := slotMinLim; i <= slotMaxLim; i++ {
			found := false
			for j := range slotsBids {
				if slotsBids[j].SlotNum == i {
					found = true
					if order == historydb.OrderAsc {
						if slotsBids[j].ItemID >= int(*fromItem) {
							slots = append(slots, slotsBids[j])
						}
					} else {
						if slotsBids[j].ItemID <= int(*fromItem) {
							slots = append([]SlotAPI{slotsBids[j]}, slots...)
						}
					}
					break
				}
			}
			if !found {
				slots, err = addEmptySlot(slots, i, currentBlock.EthBlockNum, auctionVars, fromItem, order)
				if err != nil {
					retBadReq(err, c)
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

	// Build succesfull response
	type slotsResponse struct {
		Slots      []SlotAPI      `json:"slots"`
		Pagination *db.Pagination `json:"pagination"`
	}
	c.JSON(http.StatusOK, &slotsResponse{
		Slots:      slots,
		Pagination: getPagination(totalItems, minSlotNum, maxSlotNum),
	})
}
