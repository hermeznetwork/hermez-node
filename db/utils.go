package db

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
	"github.com/ztrue/tracerr"
)

// InitSQLDB runs migrations and registers meddlers
func InitSQLDB(port int, host, user, password, name string) (*sqlx.DB, error) {
	// Init meddler
	initMeddler()
	meddler.Default = meddler.PostgreSQL
	// Stablish connection
	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		name,
	)
	db, err := sqlx.Connect("postgres", psqlconn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	// Run DB migrations
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("hermez-db-migrations", "./migrations"),
	}
	nMigrations, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	log.Info("successfully ran ", nMigrations, " migrations")
	return db, nil
}

// initMeddler registers tags to be used to read/write from SQL DBs using meddler
func initMeddler() {
	meddler.Register("bigint", BigIntMeddler{})
	meddler.Register("bigintnull", BigIntNullMeddler{})
}

// BulkInsert performs a bulk insert with a single statement into the specified table.  Example:
// `db.BulkInsert(myDB, "INSERT INTO block (eth_block_num, timestamp, hash) VALUES %s", blocks[:])`
// Note that all the columns must be specified in the query, and they must be
// in the same order as in the table.
// Note that the fields in the structs need to be defined in the same order as
// in the table columns.
func BulkInsert(db meddler.DB, q string, args interface{}) error {
	arrayValue := reflect.ValueOf(args)
	arrayLen := arrayValue.Len()
	valueStrings := make([]string, 0, arrayLen)
	var arglist = make([]interface{}, 0)
	for i := 0; i < arrayLen; i++ {
		arg := arrayValue.Index(i).Addr().Interface()
		elemArglist, err := meddler.Default.Values(arg, true)
		if err != nil {
			return tracerr.Wrap(err)
		}
		arglist = append(arglist, elemArglist...)
		value := "("
		for j := 0; j < len(elemArglist); j++ {
			value += fmt.Sprintf("$%d, ", i*len(elemArglist)+j+1)
		}
		value = value[:len(value)-2] + ")"
		valueStrings = append(valueStrings, value)
	}
	stmt := fmt.Sprintf(q, strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, arglist...)
	return tracerr.Wrap(err)
}

// BigIntMeddler encodes or decodes the field value to or from JSON
type BigIntMeddler struct{}

// PreRead is called before a Scan operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// give a pointer to a byte buffer to grab the raw data
	return new(string), nil
}

// PostRead is called after a Scan operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PostRead(fieldPtr, scanTarget interface{}) error {
	ptr := scanTarget.(*string)
	if ptr == nil {
		return tracerr.Wrap(fmt.Errorf("BigIntMeddler.PostRead: nil pointer"))
	}
	field := fieldPtr.(**big.Int)
	*field = new(big.Int).SetBytes([]byte(*ptr))
	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(*big.Int)

	return field.Bytes(), nil
}

// BigIntNullMeddler encodes or decodes the field value to or from JSON
type BigIntNullMeddler struct{}

// PreRead is called before a Scan operation for fields that have the BigIntNullMeddler
func (b BigIntNullMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	return &fieldAddr, nil
}

// PostRead is called after a Scan operation for fields that have the BigIntNullMeddler
func (b BigIntNullMeddler) PostRead(fieldPtr, scanTarget interface{}) error {
	field := fieldPtr.(**big.Int)
	ptrPtr := scanTarget.(*interface{})
	if *ptrPtr == nil {
		// null column, so set target to be zero value
		*field = nil
		return nil
	}
	// not null
	ptr := (*ptrPtr).([]byte)
	if ptr == nil {
		return tracerr.Wrap(fmt.Errorf("BigIntMeddler.PostRead: nil pointer"))
	}
	*field = new(big.Int).SetBytes(ptr)
	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the BigIntNullMeddler
func (b BigIntNullMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(*big.Int)
	if field == nil {
		return nil, nil
	}
	return field.Bytes(), nil
}

// SliceToSlicePtrs converts any []Foo to []*Foo
func SliceToSlicePtrs(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	vLen := v.Len()
	typ := v.Type().Elem()
	res := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(typ)), vLen, vLen)
	for i := 0; i < vLen; i++ {
		res.Index(i).Set(v.Index(i).Addr())
	}
	return res.Interface()
}

// SlicePtrsToSlice converts any []*Foo to []Foo
func SlicePtrsToSlice(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	vLen := v.Len()
	typ := v.Type().Elem().Elem()
	res := reflect.MakeSlice(reflect.SliceOf(typ), vLen, vLen)
	for i := 0; i < vLen; i++ {
		res.Index(i).Set(v.Index(i).Elem())
	}
	return res.Interface()
}

// Rollback an sql transaction, and log the error if it's not nil
func Rollback(txn *sqlx.Tx) {
	err := txn.Rollback()
	if err != nil {
		log.Errorw("Rollback", "err", err)
	}
}
