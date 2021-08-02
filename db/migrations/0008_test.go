package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration updates the tx_pool table

type migrationTest0008 struct{}

func (m migrationTest0008) InsertData(db *sqlx.DB) error {
	return nil
}

func (m migrationTest0008) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	//Insert data in the tx_pool table
	const queryInsert = `INSERT INTO tx_pool (tx_id,
		from_idx,
		effective_from_eth_addr,
		effective_from_bjj,
		to_idx,
		to_eth_addr,
		to_bjj,
		effective_to_eth_addr,
		effective_to_bjj,
		token_id,
		amount,
		amount_f,
		fee,
		nonce,
		state,
		info,
		signature,
		"timestamp",
		batch_num,
		rq_from_idx,
		rq_to_idx,
		rq_to_eth_addr,
		rq_to_bjj,
		rq_token_id,
		rq_amount,
		rq_fee,
		rq_nonce,
		tx_type,
		client_ip,
		external_delete,
		item_id,
		error_code,
		error_type)	VALUES(decode('03A193BC53932580F2EF91B5DA038AF611D9F1D896D518CDD65B1D766CBD835E30','hex'),
		3142,
		decode('380ED8BD696C78395FB1961BDA42739D2F5242A1','hex'),
		decode('CA780AF6B4C6164157DF737CE3E1E0A29EC9523F5B2CB4ADC26560379BFD5080','hex'),
		NULL,
		decode('15868E0C2DFC14A47FFC7360A93ADDC994386B11','hex'),
		NULL,
		decode('15868E0C2DFC14A47FFC7360A93ADDC994386B11','hex'),
		NULL,
		0,
		34565432000000,
		34565432000000,
		20,
		28,
		'fged',
		NULL,
		decode('226B72179B58EC2D2106EAF40D828DF31F1FA92F2ED7DAC263E04259BDCE3085C803B7EC7F57E44E0C63234E52BFD28404332204B2F53A4589CB0B83531B0B05','hex'),
		'2021-07-16 10:19:15.671',
		6164,
		NULL, NULL,	NULL, NULL,	NULL, NULL, NULL, NULL,
		'TransferToEthAddr',
		'95.127.153.55',
		false,
		28494,
		15,
		'ErrToIdxNotFound');
	`
	_, err := db.Exec(queryInsert)
	assert.NoError(t, err)
	const queryGetNumberItems = `select count(*) from tx_pool;`
	row := db.QueryRow(queryGetNumberItems)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0008) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the new fields can't be inserted in tx_pool table
	const queryInsert = `INSERT INTO tx_pool (tx_id,
		from_idx,
		effective_from_eth_addr,
		effective_from_bjj,
		to_idx,
		to_eth_addr,
		to_bjj,
		effective_to_eth_addr,
		effective_to_bjj,
		token_id,
		amount,
		amount_f,
		fee,
		nonce,
		state,
		info,
		signature,
		"timestamp",
		batch_num,
		rq_from_idx,
		rq_to_idx,
		rq_to_eth_addr,
		rq_to_bjj,
		rq_token_id,
		rq_amount,
		rq_fee,
		rq_nonce,
		tx_type,
		client_ip,
		external_delete,
		item_id,
		error_code,
		error_type)	VALUES(decode('02A194BC53932580F2EF91B5DA038AF611D9F1D896D518CDD65B1D766CBD835E31','hex'),
		3142,
		decode('380ED8BD696C78395FB1961BDA42739D2F5242A1','hex'),
		decode('CA780AF6B4C6164157DF737CE3E1E0A29EC9523F5B2CB4ADC26560379BFD5080','hex'),
		NULL,
		decode('15868E0C2DFC14A47FFC7360A93ADDC994386B11','hex'),
		NULL,
		decode('15868E0C2DFC14A47FFC7360A93ADDC994386B11','hex'),
		NULL,
		0,
		34565432000000,
		34565432000000,
		20,
		28,
		'fged',
		NULL,
		decode('226B72179B58EC2D2106EAF40D828DF31F1FA92F2ED7DAC263E04259BDCE3085C803B7EC7F57E44E0C63234E52BFD28404332204B2F53A4589CB0B83531B0B05','hex'),
		'2021-07-16 10:19:15.671',
		6164,
		NULL, NULL,	NULL, NULL,	NULL, NULL, NULL, NULL,
		'TransferToEthAddr',
		'95.127.153.55',
		false,
		28495,
		15,
		'ErrToIdxNotFound');
	`
	_, err := db.Exec(queryInsert)
	if assert.Error(t, err) {
		assert.Equal(t, `pq: column "error_code" of relation "tx_pool" does not exist`, err.Error())
	}
	const query2Insert = `INSERT INTO tx_pool (tx_id,
		from_idx,
		effective_from_eth_addr,
		effective_from_bjj,
		to_idx,
		to_eth_addr,
		to_bjj,
		effective_to_eth_addr,
		effective_to_bjj,
		token_id,
		amount,
		amount_f,
		fee,
		nonce,
		state,
		info,
		signature,
		"timestamp",
		batch_num,
		rq_from_idx,
		rq_to_idx,
		rq_to_eth_addr,
		rq_to_bjj,
		rq_token_id,
		rq_amount,
		rq_fee,
		rq_nonce,
		tx_type,
		client_ip,
		external_delete,
		item_id) VALUES(decode('02A193BC63932580F2EF91B5DA038AF611D9F1D896D518CDD65B1D766CBD835E32','hex'),
		3142,
		decode('380ED8BD696C78395FB1961BDA42739D2F5242A1','hex'),
		decode('CA780AF6B4C6164157DF737CE3E1E0A29EC9523F5B2CB4ADC26560379BFD5080','hex'),
		NULL,
		decode('15868E0C2DFC14A47FFC7360A93ADDC994386B11','hex'),
		NULL,
		decode('15868E0C2DFC14A47FFC7360A93ADDC994386B11','hex'),
		NULL,
		0,
		34565432000000,
		34565432000000,
		20,
		28,
		'fged',
		NULL,
		decode('226B72179B58EC2D2106EAF40D828DF31F1FA92F2ED7DAC263E04259BDCE3085C803B7EC7F57E44E0C63234E52BFD28404332204B2F53A4589CB0B83531B0B05','hex'),
		'2021-07-16 10:19:15.671',
		6164,
		NULL, NULL,	NULL, NULL,	NULL, NULL, NULL, NULL,
		'TransferToEthAddr',
		'95.127.153.55',
		false,
		28496);
	`
	_, err = db.Exec(query2Insert)
	assert.NoError(t, err)
}

func TestMigration0008(t *testing.T) {
	runMigrationTest(t, 8, migrationTest0008{})
}
