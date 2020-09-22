package db

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
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
		return nil, err
	}
	// Run DB migrations
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("hermez-db-migrations", "./migrations"),
	}
	nMigrations, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, err
	}
	log.Info("Successfuly runt ", nMigrations, " migrations")
	return db, nil
}

// initMeddler registers tags to be used to read/write from SQL DBs using meddler
func initMeddler() {
	meddler.Register("bigint", BigIntMeddler{})
	meddler.Register("bigintnull", BigIntNullMeddler{})
}

// BulkInsert performs a bulk insert with a single statement into the specified table.  Example:
// `db.BulkInsert(myDB, "INSERT INTO block (eth_block_num, timestamp, hash) VALUES %s", blocks[:])`
// Note that all the columns must be specified in the query, and they must be in the same order as in the table.
func BulkInsert(db meddler.DB, q string, args interface{}) error {
	arrayValue := reflect.ValueOf(args)
	arrayLen := arrayValue.Len()
	valueStrings := make([]string, 0, arrayLen)
	var arglist = make([]interface{}, 0)
	for i := 0; i < arrayLen; i++ {
		arg := arrayValue.Index(i).Addr().Interface()
		elemArglist, err := meddler.Default.Values(arg, true)
		if err != nil {
			return err
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
	return err
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
		return fmt.Errorf("BigIntMeddler.PostRead: nil pointer")
	}

	data, err := base64.StdEncoding.DecodeString(*ptr)
	if err != nil {
		return fmt.Errorf("big.Int decode error: %v", err)
	}
	field := fieldPtr.(**big.Int)
	*field = new(big.Int).SetBytes(data)

	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the BigIntMeddler
func (b BigIntMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(*big.Int)

	str := base64.StdEncoding.EncodeToString(field.Bytes())

	return str, nil
}

// BigIntNullMeddler encodes or decodes the field value to or from JSON
type BigIntNullMeddler struct{}

// PreRead is called before a Scan operation for fields that have the BigIntNullMeddler
func (b BigIntNullMeddler) PreRead(fieldAddr interface{}) (scanTarget interface{}, err error) {
	return &fieldAddr, nil
}

// PostRead is called after a Scan operation for fields that have the BigIntNullMeddler
func (b BigIntNullMeddler) PostRead(fieldPtr, scanTarget interface{}) error {
	sv := reflect.ValueOf(scanTarget)
	if sv.Elem().IsNil() {
		// null column, so set target to be zero value
		fv := reflect.ValueOf(fieldPtr)
		fv.Elem().Set(reflect.Zero(fv.Elem().Type()))
		return nil
	}
	// not null
	encoded := new([]byte)
	refEnc := reflect.ValueOf(encoded)
	refEnc.Elem().Set(sv.Elem().Elem())
	data, err := base64.StdEncoding.DecodeString(string(*encoded))
	if err != nil {
		return fmt.Errorf("big.Int decode error: %v", err)
	}
	field := fieldPtr.(**big.Int)
	*field = new(big.Int).SetBytes(data)
	return nil
}

// PreWrite is called before an Insert or Update operation for fields that have the BigIntNullMeddler
func (b BigIntNullMeddler) PreWrite(fieldPtr interface{}) (saveValue interface{}, err error) {
	field := fieldPtr.(*big.Int)
	if field == nil {
		return nil, nil
	}
	return base64.StdEncoding.EncodeToString(field.Bytes()), nil
}
