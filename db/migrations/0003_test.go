package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration adds the column `eth_tx_hash` on `batch`

type migrationTest0003 struct{}

func (m migrationTest0003) InsertData(db *sqlx.DB) error {
	// insert block to respect the FKey of batch
	const queryInsertBlock = `INSERT INTO block (
		eth_block_num,"timestamp",hash
	) VALUES (
		4417296,'2021-03-10 16:44:06.000',decode('C4D46677F3B2511D96389521C2BDFFE91127DE214423FF14899A6177631D2105','hex')
	);`
	// insert batch
	const queryInsertBatch = `INSERT INTO batch (
		batch_num, 
		eth_block_num, 
		forger_addr, 
		fees_collected, 
		fee_idxs_coordinator, 
		state_root, 
		num_accounts, 
		last_idx, 
		exit_root, 
		forge_l1_txs_num, 
		slot_num, 
		total_fees_usd
	) VALUES (
		6758, 
		4417296, 
		decode('459264CC7D2BF350AFDDA828C273E81367729C1F', 'hex'),
		decode('7B2230223A34383337383531313632323134343030307D0A', 'hex'),
		decode('5B3236335D0A', 'hex'),
		12898140512818699175738765060248919016800434587665040485377676113605873428098, 
		256, 
		1044, 
		0, 
		NULL, 
		717, 
		115.047487133272
	);`
	_, err := db.Exec(queryInsertBlock + queryInsertBatch)
	return err
}

func (m migrationTest0003) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	// check that the batch inserted in previous step is persisted with same content
	const queryGetBatch = `SELECT COUNT(*) FROM batch WHERE
		batch_num = 6758 AND
		eth_block_num = 4417296 AND
		forger_addr = decode('459264CC7D2BF350AFDDA828C273E81367729C1F', 'hex') AND
		fees_collected = decode('7B2230223A34383337383531313632323134343030307D0A', 'hex') AND
		fee_idxs_coordinator = decode('5B3236335D0A', 'hex') AND
		state_root = 12898140512818699175738765060248919016800434587665040485377676113605873428098 AND
		num_accounts = 256 AND
		last_idx = 1044 AND
		exit_root = 0 AND
		forge_l1_txs_num IS NULL AND
		slot_num = 717 AND
		total_fees_usd = 115.047487133272 AND
		eth_tx_hash = DECODE('0000000000000000000000000000000000000000000000000000000000000000', 'hex');
	`
	row := db.QueryRow(queryGetBatch)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0003) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the batch inserted in previous step is persisted with same content
	const queryGetBatch = `SELECT COUNT(*) FROM batch WHERE
		batch_num = 6758 AND
		eth_block_num = 4417296 AND
		forger_addr = decode('459264CC7D2BF350AFDDA828C273E81367729C1F', 'hex') AND
		fees_collected = decode('7B2230223A34383337383531313632323134343030307D0A', 'hex') AND
		fee_idxs_coordinator = decode('5B3236335D0A', 'hex') AND
		state_root = 12898140512818699175738765060248919016800434587665040485377676113605873428098 AND
		num_accounts = 256 AND
		last_idx = 1044 AND
		exit_root = 0 AND
		forge_l1_txs_num IS NULL AND
		slot_num = 717 AND
		total_fees_usd = 115.047487133272;
	`
	row := db.QueryRow(queryGetBatch)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
	// check that eth_tx_hash table doesn't exist anymore
	const queryCheckItemID = `SELECT COUNT(*) FROM batch WHERE eth_tx_hash IS NULL;`
	row = db.QueryRow(queryCheckItemID)
	assert.Equal(t, `pq: column "eth_tx_hash" does not exist`, row.Scan(&result).Error())
}

func TestMigration0003(t *testing.T) {
	runMigrationTest(t, 3, migrationTest0003{})
}
