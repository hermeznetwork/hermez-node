package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration adds the column `eth_tx_hash` on `tx`

type migrationTest0010 struct{}

func (m migrationTest0010) InsertData(db *sqlx.DB) error {
	// insert tx
	const queryInsert = `
	INSERT INTO block (eth_block_num, "timestamp", hash)
	VALUES(163, '2021-08-16 10:34:30.000', decode('2A4D566DE659EADD5188D0ACF7155E4010BD331033D14CD2610C467285AFFF51','hex'));
	INSERT INTO block (eth_block_num, "timestamp", hash)
	VALUES(187, '2021-08-16 10:34:30.000', decode('2A4D566DE659EADD5188D0ACF7155E4010BD331033D14CD2610C467285AFFF5E','hex'));
	INSERT INTO block (eth_block_num, "timestamp", hash)
	VALUES(191, '2021-08-16 10:34:42.000', decode('29F4F4128E6E8165DB13A1638C9C0B52B759B8FAEA9F339754D10F0050F3ACA1','hex'));
	INSERT INTO block (eth_block_num, "timestamp", hash)
	VALUES(197, '2021-08-16 10:34:42.000', decode('29F4F4128E6E8165DB13A1638C9C0B52B759B8FAEA9F339754D10F0050F3ACA3','hex'));
	INSERT INTO block (eth_block_num, "timestamp", hash)
	VALUES(206, '2021-08-16 10:34:42.000', decode('29F4F4128E6E8165DB13A1638C9C0B52B759B8FAEA9F339754D10F0050F3ACA2','hex'));

	INSERT INTO batch (item_id, batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, eth_tx_hash)
	VALUES(6, 6, 197, decode('DCC5DD922FB1D0FD0C450A0636A8CE827521F0ED','hex'), decode('7B7D0A','hex'), decode('5B5D0A','hex'), 18824406808947086769639114603381416975270151877490721532469555446804684467592, 1, 256, 0, 5, 2, 0, decode('21BCAEC5472E6F578A5938516A76B438AF7D1739CB10C65A86D5F0FCF19A514F','hex'));
	INSERT INTO batch (item_id, batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, eth_tx_hash)
	VALUES(7, 7, 206, decode('DCC5DD922FB1D0FD0C450A0636A8CE827521F0ED','hex'), decode('7B7D0A','hex'), decode('5B5D0A','hex'), 7234362616256345251637054085980477919716486801577186942795287784595347908748, 1, 257, 0, 6, 3, 0, decode('7BEFD7FEE85583F26C236BAF3DAA98D0C35E60F5DBCF53865230B067ABFD31E1','hex'));

	INSERT INTO token (item_id, token_id, eth_block_num, eth_addr, "name", symbol, decimals, usd, usd_update)
	VALUES(2, 11, 163, decode('66AC1B9605D439F2ECDCDE0C5C1FAA41A66537A1','hex'), 'ERC20_0', '20_0', 18, NULL, NULL);

	INSERT INTO account (item_id, idx, token_id, batch_num, bjj, eth_addr)
	VALUES(1, 256, 11, 6, decode('D746824F7D0AC5044A573F51B278ACB56D823BEC39551D1D7BF7378B68A1B021','hex'), decode('27E9727FD9B8CDDDD0854F56712AD9DF647FAB74','hex'));
	INSERT INTO account (item_id, idx, token_id, batch_num, bjj, eth_addr)
	VALUES(2, 257, 11, 7, decode('4D05C307400C65795F02DB96B1B81C60386FD53E947D9D3F749F3D99B1853909','hex'), decode('9766D2E7FFDE358AD0A40BB87C4B88D9FAC3F4DD','hex'));

	INSERT INTO tx (item_id, is_l1, id, "type", "position", from_idx, effective_from_idx, from_eth_addr, from_bjj, to_idx, to_eth_addr, to_bjj, amount, amount_success, amount_f, token_id, amount_usd, batch_num, eth_block_num, to_forge_l1_txs_num, user_origin, deposit_amount, deposit_amount_success, deposit_amount_f, deposit_amount_usd, fee, fee_usd, nonce)
	VALUES(3, true, decode('00C33F316240F8D33A973DB2D0E901E4AC1C96DE30B185FCC6B63DAC4D0E147BD4','hex'), 'CreateAccountDeposit', 0, 0, 256, decode('27E9727FD9B8CDDDD0854F56712AD9DF647FAB74','hex'), decode('D746824F7D0AC5044A573F51B278ACB56D823BEC39551D1D7BF7378B68A1B021','hex'), 0, NULL, NULL, 0, true, 0, 11, NULL, 7, 187, 6, true, 1000000000000000000, true, 1000000000000000000, NULL, NULL, NULL, NULL);
	`
	_, err := db.Exec(queryInsert)
	return err
}

func (m migrationTest0010) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	// check that the tx inserted in previous step is persisted with same content
	const queryGetTx = `SELECT COUNT(*) FROM tx WHERE
	item_id = 3 AND is_l1 = true AND id = decode('00C33F316240F8D33A973DB2D0E901E4AC1C96DE30B185FCC6B63DAC4D0E147BD4','hex');`
	row := db.QueryRow(queryGetTx)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)

	insert := `INSERT INTO tx
	(item_id, is_l1, id, "type", "position", from_idx, effective_from_idx, from_eth_addr, from_bjj, to_idx, to_eth_addr, to_bjj, amount, amount_success, amount_f, token_id, amount_usd, batch_num, eth_block_num, to_forge_l1_txs_num, user_origin, deposit_amount, deposit_amount_success, deposit_amount_f, deposit_amount_usd, fee, fee_usd, nonce, eth_tx_hash, l1_fee)
	VALUES(4, true, decode('00B55F0882C5229D1BE3D9D3C1A076290F249CD0BAE5AE6E609234606BEFB91233','hex'), 'CreateAccountDeposit', 1, 0, 257, decode('9766D2E7FFDE358AD0A40BB87C4B88D9FAC3F4DD','hex'), decode('4D05C307400C65795F02DB96B1B81C60386FD53E947D9D3F749F3D99B1853909','hex'), 0, NULL, NULL, 0, true, 0, 11, NULL, 7, 191, 6, true, 1000000000000000000, true, 1000000000000000000, NULL, NULL, NULL, NULL, DECODE('6b9cb28230289dcfb748859531f17ebb305c638a970c5c84377fc680fddfcf80', 'hex'), 3007800000);`
	_, err := db.Exec(insert)
	assert.NoError(t, err)
}

func (m migrationTest0010) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the tx inserted in previous step is persisted with same content
	const queryGetTx = `SELECT COUNT(*) FROM tx WHERE
	item_id = 4 AND is_l1 = true AND id = decode('00B55F0882C5229D1BE3D9D3C1A076290F249CD0BAE5AE6E609234606BEFB91233','hex');`
	row := db.QueryRow(queryGetTx)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)

	// check that eth_tx_hash and l1_fee fields don't exist anymore
	const queryCheckEthTxHash = `SELECT COUNT(*) FROM tx WHERE eth_tx_hash IS NULL;`
	row = db.QueryRow(queryCheckEthTxHash)
	assert.Equal(t, `pq: column "eth_tx_hash" does not exist`, row.Scan(&result).Error())
	const queryCheckFee = `SELECT COUNT(*) FROM tx WHERE l1_fee IS NULL;`
	row = db.QueryRow(queryCheckFee)
	assert.Equal(t, `pq: column "l1_fee" does not exist`, row.Scan(&result).Error())
}

func TestMigration0010(t *testing.T) {
	runMigrationTest(t, 10, migrationTest0010{})
}
