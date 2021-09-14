package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration adds the column `gas_price` and `gas_used` on `batch` table

type migrationTest0011 struct{}

func (m migrationTest0011) InsertData(db *sqlx.DB) error {
	// insert tx
	const queryInsert = `
	INSERT INTO block
	(eth_block_num, "timestamp", hash)
	VALUES(48295, '2021-09-13 08:28:39.000', decode('2AB24E7021318D6CF0686E8F8FBFB0A63CB79A9FB5CDECE7C09FD4438E67242F','hex'));
	INSERT INTO block
	(eth_block_num, "timestamp", hash)
	VALUES(48286, '2021-09-13 08:28:39.000', decode('2AB24E7021318D6CF0686E8F8FBFB0A63CB79A9FB5CDECE7C09FD4438E67242A','hex'));
	INSERT INTO block
	(eth_block_num, "timestamp", hash)
	VALUES(48278, '2021-09-13 08:28:39.000', decode('2AB24E7021318D6CF0686E8F8FBFB0A63CB79A9FB5CDECE7C09FD4438E67242E','hex'));

	INSERT INTO batch
	(item_id, batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, eth_tx_hash)
	VALUES(1420, 1420, 48295, decode('DCC5DD922FB1D0FD0C450A0636A8CE827521F0ED','hex'), decode('7B7D0A','hex'), decode('5B5D0A','hex'), 0, 0, 255, 0, 1419, 1205, 0, decode('AE80AB27E97213DEC805C78ED9C637E0414A541D489377F766B3372170F4AD66','hex'));
	INSERT INTO batch
	(item_id, batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, eth_tx_hash)
	VALUES(1419, 1419, 48286, decode('DCC5DD922FB1D0FD0C450A0636A8CE827521F0ED','hex'), decode('7B7D0A','hex'), decode('5B5D0A','hex'), 0, 0, 255, 0, 1418, 1205, 0, decode('4BC9C94E8CF93AD475F8C8394BC934AF5EB0802FE4009D13F58AE25F6047DA95','hex'));
	`
	_, err := db.Exec(queryInsert)
	return err
}

func (m migrationTest0011) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	// check that the batch inserted in previous step is persisted with same content
	const queryGetBatch = `SELECT COUNT(*) FROM batch WHERE eth_tx_hash = decode('4BC9C94E8CF93AD475F8C8394BC934AF5EB0802FE4009D13F58AE25F6047DA95','hex');`
	row := db.QueryRow(queryGetBatch)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)

	insert := `INSERT INTO batch
	(item_id, batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, eth_tx_hash, gas_price, gas_used, ether_price_usd)
	VALUES(1418, 1418, 48278, decode('DCC5DD922FB1D0FD0C450A0636A8CE827521F0ED','hex'), decode('7B7D0A','hex'), decode('5B5D0A','hex'), 0, 0, 255, 0, 1417, 1204, 0, decode('285CE6A154901AF5197382DC8A5CCE02588BDA1B078768C5077B6996FA2EA0A7','hex'), 500000000000, 15000000, 3492.21);
	`
	_, err := db.Exec(insert)
	assert.NoError(t, err)
}

func (m migrationTest0011) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the batch inserted in previous step is persisted with same content
	const queryGetTx = `SELECT COUNT(*) FROM batch WHERE eth_tx_hash = decode('4BC9C94E8CF93AD475F8C8394BC934AF5EB0802FE4009D13F58AE25F6047DA95','hex');`
	row := db.QueryRow(queryGetTx)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)

	// check that eth_tx_hash and l1_fee fields don't exist anymore
	const queryCheckGasPrice = `SELECT COUNT(*) FROM batch WHERE gas_price = 0;`
	row = db.QueryRow(queryCheckGasPrice)
	assert.Equal(t, `pq: column "gas_price" does not exist`, row.Scan(&result).Error())
	const queryCheckGasUsed = `SELECT COUNT(*) FROM batch WHERE gas_used = 0;`
	row = db.QueryRow(queryCheckGasUsed)
	assert.Equal(t, `pq: column "gas_used" does not exist`, row.Scan(&result).Error())
	const queryCheckEtherPrice = `SELECT COUNT(*) FROM batch WHERE ether_price_usd = 0;`
	row = db.QueryRow(queryCheckEtherPrice)
	assert.Equal(t, `pq: column "ether_price_usd" does not exist`, row.Scan(&result).Error())
}

func TestMigration0011(t *testing.T) {
	runMigrationTest(t, 11, migrationTest0011{})
}
