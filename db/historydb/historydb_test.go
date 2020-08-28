package historydb

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
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

func TestBlocks(t *testing.T) {
	var fromBlock, toBlock uint64
	fromBlock = 1
	toBlock = 5
	// Delete peviously created rows (clean previous test execs)
	assert.NoError(t, historyDB.Reorg(fromBlock-1))
	// Generate fake blocks
	blocks := genBlocks(fromBlock, toBlock)
	// Insert blocks into DB
	for i := 0; i < len(blocks); i++ {
		err := historyDB.AddBlock(&blocks[i])
		assert.NoError(t, err)
	}
	// Get all blocks from DB
	fetchedBlocks, err := historyDB.GetBlocks(fromBlock, toBlock)
	assert.Equal(t, len(blocks), len(fetchedBlocks))
	// Compare generated vs getted blocks
	assert.NoError(t, err)
	for i, fetchedBlock := range fetchedBlocks {
		assertEqualBlock(t, &blocks[i], fetchedBlock)
	}
	// Get blocks from the DB one by one
	for i := fromBlock; i < toBlock; i++ {
		fetchedBlock, err := historyDB.GetBlock(i)
		assert.NoError(t, err)
		assertEqualBlock(t, &blocks[i-1], fetchedBlock)
	}
	// Get last block
	lastBlock, err := historyDB.GetLastBlock()
	assert.NoError(t, err)
	assertEqualBlock(t, &blocks[len(blocks)-1], lastBlock)
}

func assertEqualBlock(t *testing.T, expected *common.Block, actual *common.Block) {
	assert.Equal(t, expected.EthBlockNum, actual.EthBlockNum)
	assert.Equal(t, expected.Hash, actual.Hash)
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
}

func TestBatches(t *testing.T) {
	const fromBlock uint64 = 1
	const toBlock uint64 = 3
	const nBatchesPerBlock = 3
	// Prepare blocks in the DB
	setTestBlocks(fromBlock, toBlock)
	// Generate fake batches
	var batches []common.Batch
	collectedFees := make(map[common.TokenID]*big.Int)
	for i := 0; i < 64; i++ {
		collectedFees[common.TokenID(i)] = big.NewInt(int64(i))
	}
	for i := fromBlock; i < toBlock; i++ {
		for j := 0; j < nBatchesPerBlock; j++ {
			batch := common.Batch{
				BatchNum:      common.BatchNum(int(i-1)*nBatchesPerBlock + j),
				EthBlockNum:   uint64(i),
				ForgerAddr:    eth.BigToAddress(big.NewInt(239457111187)),
				CollectedFees: collectedFees,
				StateRoot:     common.Hash([]byte("duhdqlwiucgwqeiu")),
				NumAccounts:   j,
				ExitRoot:      common.Hash([]byte("tykertheuhtgenuer3iuw3b")),
				SlotNum:       common.SlotNum(j),
			}
			if j%2 == 0 {
				batch.ForgeL1TxsNum = uint32(i)
			}
			batches = append(batches, batch)
		}
	}
	// Add batches to the DB
	err := historyDB.addBatches(batches)
	assert.NoError(t, err)
	// Get batches from the DB
	fetchedBatches, err := historyDB.GetBatches(0, common.BatchNum(int(toBlock-fromBlock)*nBatchesPerBlock))
	assert.NoError(t, err)
	for i, fetchedBatch := range fetchedBatches {
		assert.Equal(t, batches[i], *fetchedBatch)
	}
	// Test GetLastBatchNum
	fetchedLastBatchNum, err := historyDB.GetLastBatchNum()
	assert.NoError(t, err)
	assert.Equal(t, batches[len(batches)-1].BatchNum, fetchedLastBatchNum)
	// Test GetLastL1TxsNum
	fetchedLastL1TxsNum, err := historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
	assert.Equal(t, batches[len(batches)-1-(int(toBlock-fromBlock+1)%nBatchesPerBlock)].ForgeL1TxsNum, fetchedLastL1TxsNum)
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
	var fetchedBids []*common.Bid
	for i := fromBlock; i < toBlock; i++ {
		fetchedBidsSlot, err := historyDB.GetBidsBySlot(common.SlotNum(i))
		assert.NoError(t, err)
		fetchedBids = append(fetchedBids, fetchedBidsSlot...)
	}
	// Compare fetched bids vs generated bids
	for i, bid := range fetchedBids {
		assert.Equal(t, bids[i], *bid)
	}
}

// setTestBlocks WARNING: this will delete the blocks and recreate them
func setTestBlocks(from, to uint64) {
	if from == 0 {
		if err := historyDB.Reorg(from); err != nil {
			panic(err)
		}
	} else {
		if err := historyDB.Reorg(from - 1); err != nil {
			panic(err)
		}
	}
	blocks := genBlocks(from, to)
	if err := addBlocks(blocks); err != nil {
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

// addBlocks insert blocks into the DB. TODO: move method to test
func addBlocks(blocks []common.Block) error {
	return db.BulkInsert(
		historyDB.db,
		"INSERT INTO block (eth_block_num, timestamp, hash) VALUES %s",
		blocks[:],
	)
}
