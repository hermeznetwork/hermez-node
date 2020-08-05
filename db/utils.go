package db

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/russross/meddler"
)

// InitMeddler registers tags to be used to read/write from SQL DBs using meddler
func InitMeddler() {
	meddler.Register("bigint", BigIntMeddler{})
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
