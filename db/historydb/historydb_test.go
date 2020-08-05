package historydb

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

var historyDB *HistoryDB

// In order to run the test you need to run a Posgres DB with
// a database named "history" that is accessible by
// user: "hermez"
// pass: set it using the env var POSTGRES_PASS
// This can be achieved by running: POSTGRES_PASS=your_strong_pass && sudo docker run --rm --name hermez-db-test -p 5432:5432 -e POSTGRES_DB=history -e POSTGRES_USER=hermez -e POSTGRES_PASSWORD=$POSTGRES_PASS -d postgres && sleep 2s && sudo docker exec -it hermez-db-test psql -a history -U hermez -c "CREATE DATABASE l2;"
// After running the test you can stop the container by running: sudo docker kill hermez-db-test
// If you already did that for the L2DB you don't have to do it again

func TestMain(m *testing.M) {
	// init DB
	var err error
	pass := os.Getenv("POSTGRES_PASS")
	historyDB, err = NewHistoryDB(5432, "localhost", "hermez", pass, "history")
	if err != nil {
		panic(err)
	}
	// Run tests
	result := m.Run()
	// Close DB
	if err := historyDB.Close(); err != nil {
		fmt.Println("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func TestAddBlock(t *testing.T) {
	var fromBlock, toBlock uint64
	fromBlock = 1
	toBlock = 5
	// Delete peviously created rows (clean previous test execs)
	assert.NoError(t, historyDB.reorg(fromBlock-1))
	// Generate fake blocks
	blocks := genBlocks(fromBlock, toBlock)
	// Insert blocks into DB
	err := historyDB.addBlocks(blocks)
	assert.NoError(t, err)
	// Get blocks from DB
	fetchedBlocks, err := historyDB.GetBlocks(fromBlock, toBlock)
	// Compare generated vs getted blocks
	assert.NoError(t, err)
	for i, fetchedBlock := range fetchedBlocks {
		assert.Equal(t, blocks[i].EthBlockNum, fetchedBlock.EthBlockNum)
		assert.Equal(t, blocks[i].Hash, fetchedBlock.Hash)
		assert.Equal(t, blocks[i].Timestamp.Unix(), fetchedBlock.Timestamp.Unix())
	}
}

func TestBids(t *testing.T) {
	const fromBlock uint64 = 1
	const toBlock uint64 = 5
	const bidsPerSlot = 5
	// Prepare blocks in the DB
	setTestBlocks(fromBlock, toBlock)
	// Generate fake bids
	bids := make([]common.Bid, 0, (toBlock-fromBlock)*bidsPerSlot)
	for i := fromBlock; i < toBlock; i++ {
		for j := 0; j < bidsPerSlot; j++ {
			bids = append(bids, common.Bid{
				SlotNum:     common.SlotNum(i),
				BidValue:    big.NewInt(int64(j)),
				EthBlockNum: i,
				ForgerAddr:  eth.BigToAddress(big.NewInt(int64(j))),
			})
		}
	}
	err := historyDB.addBids(bids)
	assert.NoError(t, err)
	// Fetch bids
	fetchedBidsPtr, err := historyDB.GetBidsByBlock(fromBlock, toBlock)
	assert.NoError(t, err)
	// Compare fetched bids vs generated bids
	fetchedBids := make([]common.Bid, 0, (toBlock-fromBlock)*bidsPerSlot)
	for _, bid := range fetchedBidsPtr {
		fetchedBids = append(fetchedBids, *bid)
	}
	assert.Equal(t, bids, fetchedBids)
}

// setTestBlocks WARNING: this will delete the blocks and recreate them
func setTestBlocks(from, to uint64) {
	if from == 0 {
		if err := historyDB.reorg(from); err != nil {
			panic(err)
		}
	} else {
		if err := historyDB.reorg(from - 1); err != nil {
			panic(err)
		}
	}
	blocks := genBlocks(from, to)
	if err := historyDB.addBlocks(blocks); err != nil {
		panic(err)
	}
}

func genBlocks(from, to uint64) []common.Block {
	var blocks []common.Block
	for i := from; i < to; i++ {
		blocks = append(blocks, common.Block{
			EthBlockNum: i,
			Timestamp:   time.Now().Add(time.Second * 13).UTC(),
			Hash:        eth.BigToHash(big.NewInt(int64(i))),
		})
	}
	return blocks
}
