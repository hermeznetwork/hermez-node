package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration adds the column `rq_offset` and `atomic_group_id` on `tx_pool`

type migrationTest0007 struct{}

func (m migrationTest0007) InsertData(db *sqlx.DB) error {
	// insert block to respect the FKey of token
	const queryInsertBlock = `INSERT INTO block (
		eth_block_num,"timestamp",hash
	) VALUES (
		4417296,'2021-03-10 16:44:06.000',decode('C4D46677F3B2511D96389521C2BDFFE91127DE214423FF14899A6177631D2105','hex')
	);`
	// insert token to respect the FKey of tx_pool
	const queryInsertToken = `INSERT INTO "token" (
		token_id,eth_block_num,eth_addr,"name",symbol,decimals,usd,usd_update
	) VALUES (
		2,4417296,decode('1B36A4DED4DF40248C0E0E52CEA5EDC9A298B721','hex'),'Dai Stablecoin','DAI',18,1.01,'2021-04-17 20:21:16.870'
	);`
	// insert batch to respect the FKey of account
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
	// isert account to set effective_from_eth_addr, effective_from_bjj through trigger
	const queryInsertAccount = `INSERT INTO account (
		idx,token_id,batch_num,bjj,eth_addr
	) VALUES (
		789,2,6758,decode('FDDACE21457376B0952CCD19CE66B854FDD7C6E45905B0A0A75747C87D41719A','hex'),decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F','hex')
	);`
	// insert a row in tx_pool with all the fields setted
	const queryInsertTxPool = `INSERT INTO tx_pool (
		tx_id, from_idx, effective_from_eth_addr, effective_from_bjj, to_idx, to_eth_addr, to_bjj, effective_to_eth_addr, effective_to_bjj,
		token_id, amount, amount_f, fee, nonce, state, info, signature, "timestamp", batch_num,
		rq_from_idx,rq_to_idx,rq_to_eth_addr,rq_to_bjj,rq_token_id,rq_amount,rq_fee,rq_nonce,tx_type,client_ip,external_delete,rq_offset,atomic_group_id
	) VALUES (
		decode('023A0D72BEB1095C28A7130D896F484CC9D465C1C95F1617C0A7B2094E3E1F11FF', 'hex'),
		789,
		decode('FF', 'hex'), -- Note that this field will be replaced by trigger with account table values
		decode('FF', 'hex'), -- Note that this field will be replaced by trigger with account table values
		1,
		decode('1224456678907543564567567567567657567567', 'hex'),
		decode('1224456678907543564567567567567657567567000000000000000000000000', 'hex'),
		decode('1224456678907543564567567567567657567567', 'hex'),
		decode('1224456678907543564567567567567657567567000000000000000000000000', 'hex'),
		2,
		5,
		5,
		227,
		3,
		'pend',
		'Exits with amount 0 have no sense, not accepting to prevent unintended transactions',
		decode('9C6A159C57D7FC58E3E5D3510FBC64EAC9C0D56A1B3144D94D6BBA4C23B9402CEE57D0CFF4A3BE135CBD2393AB8FD2A1840A62281B1721801DBF708D27F1DF00', 'hex'),
		'2021-05-06 15:02:47.616',
		32,
		765,
		567,
		decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex'),
		decode('FDDACE21457376B0952CCD19CE66B854FDD7C6E45905B0A0A75747C87D41719A', 'hex'),
		5,
		345345345,
		33,
		4,
		'Exit',
		'93.176.174.84',
		true,
		1,
		decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex')
	);`
	_, err := db.Exec(queryInsertBlock +
		queryInsertToken +
		queryInsertBatch +
		queryInsertAccount +
		queryInsertTxPool,
	)
	return err
}

func (m migrationTest0007) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	// check that the tx_pool inserted in previous step is persisted
	// with same content, except item_id is added
	const queryGetTxPool = `SELECT COUNT(*) FROM tx_pool WHERE
		tx_id = decode('023A0D72BEB1095C28A7130D896F484CC9D465C1C95F1617C0A7B2094E3E1F11FF', 'hex') AND
		from_idx = 789 AND
		effective_from_eth_addr = decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex') AND
		effective_from_bjj = decode('FDDACE21457376B0952CCD19CE66B854FDD7C6E45905B0A0A75747C87D41719A', 'hex') AND
		to_idx = 1 AND
		to_eth_addr = decode('1224456678907543564567567567567657567567', 'hex') AND
		to_bjj = decode('1224456678907543564567567567567657567567000000000000000000000000', 'hex') AND
		effective_to_eth_addr = decode('1224456678907543564567567567567657567567', 'hex') AND
		effective_to_bjj = decode('1224456678907543564567567567567657567567000000000000000000000000', 'hex') AND
		token_id = 2 AND
		amount = 5 AND
		amount_f = 5 AND
		fee = 227 AND
		nonce = 3 AND
		state = 'pend' AND
		info = 'Exits with amount 0 have no sense, not accepting to prevent unintended transactions' AND
		signature = decode('9C6A159C57D7FC58E3E5D3510FBC64EAC9C0D56A1B3144D94D6BBA4C23B9402CEE57D0CFF4A3BE135CBD2393AB8FD2A1840A62281B1721801DBF708D27F1DF00', 'hex') AND
		"timestamp" = '2021-05-06 15:02:47.616' AND
		batch_num = 32 AND
		rq_from_idx = 765 AND
		rq_to_idx = 567 AND
		rq_to_eth_addr = decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex') AND
		rq_to_bjj = decode('FDDACE21457376B0952CCD19CE66B854FDD7C6E45905B0A0A75747C87D41719A', 'hex') AND
		rq_token_id = 5 AND
		rq_amount = 345345345 AND
		rq_fee = 33 AND
		rq_nonce = 4 AND
		tx_type = 'Exit' AND
		client_ip = '93.176.174.84' AND
		external_delete = true AND
		item_id = 1 AND -- Note that item_id is an autoincremental column, so this value is setted automaticallyAND
		rq_offset = 1 AND
		atomic_group_id = decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex') AND
		max_num_batch IS NULL;`
	row := db.QueryRow(queryGetTxPool)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0007) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the tx_pool inserted in previous step is persisted with same content
	const queryGetTxPool = `SELECT COUNT(*) FROM tx_pool WHERE
		tx_id = decode('023A0D72BEB1095C28A7130D896F484CC9D465C1C95F1617C0A7B2094E3E1F11FF', 'hex') AND
		from_idx = 789 AND
		effective_from_eth_addr = decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex') AND
		effective_from_bjj = decode('FDDACE21457376B0952CCD19CE66B854FDD7C6E45905B0A0A75747C87D41719A', 'hex') AND
		to_idx = 1 AND
		to_eth_addr = decode('1224456678907543564567567567567657567567', 'hex') AND
		to_bjj = decode('1224456678907543564567567567567657567567000000000000000000000000', 'hex') AND
		effective_to_eth_addr = decode('1224456678907543564567567567567657567567', 'hex') AND
		effective_to_bjj = decode('1224456678907543564567567567567657567567000000000000000000000000', 'hex') AND
		token_id = 2 AND
		amount = 5 AND
		amount_f = 5 AND
		fee = 227 AND
		nonce = 3 AND
		state = 'pend' AND
		info = 'Exits with amount 0 have no sense, not accepting to prevent unintended transactions' AND
		signature = decode('9C6A159C57D7FC58E3E5D3510FBC64EAC9C0D56A1B3144D94D6BBA4C23B9402CEE57D0CFF4A3BE135CBD2393AB8FD2A1840A62281B1721801DBF708D27F1DF00', 'hex') AND
		"timestamp" = '2021-05-06 15:02:47.616' AND
		batch_num = 32 AND
		rq_from_idx = 765 AND
		rq_to_idx = 567 AND
		rq_to_eth_addr = decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex') AND
		rq_to_bjj = decode('FDDACE21457376B0952CCD19CE66B854FDD7C6E45905B0A0A75747C87D41719A', 'hex') AND
		rq_token_id = 5 AND
		rq_amount = 345345345 AND
		rq_fee = 33 AND
		rq_nonce = 4 AND
		tx_type = 'Exit' AND
		client_ip = '93.176.174.84' AND
		item_id = 1 AND
		external_delete = true AND
		rq_offset = 1 AND
		atomic_group_id = decode('A631BE6995643E6085330A31B9E1AF48DD5D6B7F', 'hex');`
	row := db.QueryRow(queryGetTxPool)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
	// check that max_num_batch colum doesn't exist anymore
	const queryCheckRqOffset = `SELECT COUNT(*) FROM tx_pool WHERE max_num_batch IS NULL;`
	row = db.QueryRow(queryCheckRqOffset)
	assert.Equal(t, `pq: column "max_num_batch" does not exist`, row.Scan(&result).Error())
}

func TestMigration0007(t *testing.T) {
	runMigrationTest(t, 7, migrationTest0007{})
}
