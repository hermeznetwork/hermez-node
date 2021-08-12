package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type migrationTest0009 struct{}

func (m migrationTest0009) InsertData(db *sqlx.DB) error {
	return nil
}

func (m migrationTest0009) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	// Test that update to_eth_addr on tx_pool also affects effective_to_eth_addr
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
	const queryUpdate = `UPDATE tx_pool SET to_eth_addr = 'hez:0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf' WHERE tx_id = decode('03A193BC53932580F2EF91B5DA038AF611D9F1D896D518CDD65B1D766CBD835E30','hex')`
	_, err = db.Exec(queryUpdate)
	assert.NoError(t, err)
	const getQuery = `SELECT COUNT(*) FROM tx_pool WHERE effective_to_eth_addr = 'hez:0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf'`
	row = db.QueryRow(getQuery)
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
	const queryDelete = `DELETE FROM tx_pool WHERE tx_id = decode('03A193BC53932580F2EF91B5DA038AF611D9F1D896D518CDD65B1D766CBD835E30','hex')`
	_, err = db.Exec(queryDelete)
	assert.NoError(t, err)
	row = db.QueryRow(getQuery)
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 0, result)
}

func (m migrationTest0009) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// Test that update on tx_pool doesn't affect effective_to_eth_addr
	// Insert data in the tx_pool table
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
	const queryUpdate = `UPDATE tx_pool SET to_eth_addr = 'hez:0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf' WHERE tx_id = decode('03A193BC53932580F2EF91B5DA038AF611D9F1D896D518CDD65B1D766CBD835E30','hex')`
	_, err = db.Exec(queryUpdate)
	assert.NoError(t, err)
	const getQuery = `SELECT COUNT(*) FROM tx_pool WHERE effective_to_eth_addr = 'hez:0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf'`
	row = db.QueryRow(getQuery)
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 0, result)
}

func TestMigration0009(t *testing.T) {
	runMigrationTest(t, 9, migrationTest0009{})
}
