package db

import (
	"math/big"
	"testing"

	"github.com/hermeznetwork/hermez-node/log"
	"github.com/russross/meddler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type foo struct {
	V int
}

func init() {
	log.Init("debug", []string{"stdout"})
}
func TestSliceToSlicePtrs(t *testing.T) {
	n := 16
	a := make([]foo, n)
	for i := 0; i < n; i++ {
		a[i] = foo{V: i}
	}
	b := SliceToSlicePtrs(a).([]*foo)
	for i := 0; i < len(a); i++ {
		assert.Equal(t, a[i], *b[i])
	}
}

func TestSlicePtrsToSlice(t *testing.T) {
	n := 16
	a := make([]*foo, n)
	for i := 0; i < n; i++ {
		a[i] = &foo{V: i}
	}
	b := SlicePtrsToSlice(a).([]foo)
	for i := 0; i < len(a); i++ {
		assert.Equal(t, *a[i], b[i])
	}
}

func TestBigInt(t *testing.T) {
	db, err := InitTestSQLDB()
	require.NoError(t, err)
	defer func() {
		_, err := db.Exec("DROP TABLE IF EXISTS test_big_int;")
		require.NoError(t, err)
		err = db.Close()
		require.NoError(t, err)
	}()

	_, err = db.Exec("DROP TABLE IF EXISTS test_big_int;")
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE test_big_int (
		item_id SERIAL PRIMARY KEY,
		value1 DECIMAL(78, 0) NOT NULL,
		value2 DECIMAL(78, 0),
		value3 DECIMAL(78, 0)
	);`)
	require.NoError(t, err)

	type Entry struct {
		ItemID int      `meddler:"item_id"`
		Value1 *big.Int `meddler:"value1,bigint"`
		Value2 *big.Int `meddler:"value2,bigintnull"`
		Value3 *big.Int `meddler:"value3,bigintnull"`
	}

	entry := Entry{ItemID: 1, Value1: big.NewInt(1234567890), Value2: big.NewInt(9876543210), Value3: nil}
	err = meddler.Insert(db, "test_big_int", &entry)
	require.NoError(t, err)

	var dbEntry Entry
	err = meddler.QueryRow(db, &dbEntry, "SELECT * FROM test_big_int WHERE item_id = 1;")
	require.NoError(t, err)
	assert.Equal(t, entry, dbEntry)
}
