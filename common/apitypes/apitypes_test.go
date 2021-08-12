package apitypes

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/iden3/go-iden3-crypto/babyjub"

	// nolint sqlite driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/meddler"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// Register meddler
	meddler.Default = meddler.SQLite
	meddler.Register("bigint", dbUtils.BigIntMeddler{})
	meddler.Register("bigintnull", dbUtils.BigIntNullMeddler{})
	// Create temporary sqlite DB
	dir, err := ioutil.TempDir("", "db")
	if err != nil {
		panic(err)
	}
	db, err = sql.Open("sqlite3", dir+"sqlite.db")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir) //nolint
	schema := `CREATE TABLE test (i BLOB);`
	if _, err := db.Exec(schema); err != nil {
		panic(err)
	}
	// Run tests
	result := m.Run()
	os.Exit(result)
}

func TestBigIntStrScannerValuer(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type bigInMeddlerStruct struct {
		I *big.Int `meddler:"i,bigint"` // note the bigint that instructs meddler to use BigIntMeddler
	}
	type bigIntStrStruct struct {
		I BigIntStr `meddler:"i"` // note that no meddler is specified, and Scan/Value will be used
	}
	type bigInMeddlerStructNil struct {
		I *big.Int `meddler:"i,bigintnull"` // note the bigint that instructs meddler to use BigIntNullMeddler
	}
	type bigIntStrStructNil struct {
		I *BigIntStr `meddler:"i"` // note that no meddler is specified, and Scan/Value will be used
	}

	// Not nil case
	// Insert into DB using meddler
	const x = int64(12345)
	fromMeddler := bigInMeddlerStruct{
		I: big.NewInt(x),
	}
	err = meddler.Insert(db, "test", &fromMeddler)
	assert.NoError(t, err)
	// Read from DB using BigIntStr
	toBigIntStr := bigIntStrStruct{}
	err = meddler.QueryRow(db, &toBigIntStr, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, fromMeddler.I.String(), string(toBigIntStr.I))
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using BigIntStr
	fromBigIntStr := bigIntStrStruct{
		I: "54321",
	}
	err = meddler.Insert(db, "test", &fromBigIntStr)
	assert.NoError(t, err)
	// Read from DB using meddler
	toMeddler := bigInMeddlerStruct{}
	err = meddler.QueryRow(db, &toMeddler, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, string(fromBigIntStr.I), toMeddler.I.String())

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using meddler
	fromMeddlerNil := bigInMeddlerStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromMeddlerNil)
	assert.NoError(t, err)
	// Read from DB using BigIntStr
	foo := BigIntStr("foo")
	toBigIntStrNil := bigIntStrStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toBigIntStrNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toBigIntStrNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using BigIntStr
	fromBigIntStrNil := bigIntStrStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromBigIntStrNil)
	assert.NoError(t, err)
	// Read from DB using meddler
	toMeddlerNil := bigInMeddlerStructNil{
		I: big.NewInt(x), // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toMeddlerNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toMeddlerNil.I)
}

func TestStrBigInt(t *testing.T) {
	type testStrBigInt struct {
		I StrBigInt
	}
	from := []byte(`{"I":"4"}`)
	to := &testStrBigInt{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, big.NewInt(4), (*big.Int)(&to.I))
}

func TestStrHezEthAddr(t *testing.T) {
	type testStrHezEthAddr struct {
		I StrHezEthAddr
	}
	withoutHez := "0xaa942cfcd25ad4d90a62358b0dd84f33b398262a"
	from := []byte(`{"I":"hez:` + withoutHez + `"}`)
	var addr ethCommon.Address
	if err := addr.UnmarshalText([]byte(withoutHez)); err != nil {
		panic(err)
	}
	to := &testStrHezEthAddr{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, addr, ethCommon.Address(to.I))
}

func TestStrHezBJJ(t *testing.T) {
	type testStrHezBJJ struct {
		I StrHezBJJ
	}
	priv := babyjub.NewRandPrivKey()
	hezBjj := NewHezBJJ(priv.Public().Compress())
	from := []byte(`{"I":"` + hezBjj + `"}`)
	to := &testStrHezBJJ{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, priv.Public().Compress(), (babyjub.PublicKeyComp)(to.I))
}

func TestStrHezIdx(t *testing.T) {
	type testStrHezIdx struct {
		I StrHezIdx
	}
	from := []byte(`{"I":"hez:foo:4"}`)
	to := &testStrHezIdx{}
	assert.NoError(t, json.Unmarshal(from, to))
	assert.Equal(t, common.Idx(4), common.Idx(to.I))
}

func TestHezEthAddr(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type ethAddrStruct struct {
		I ethCommon.Address `meddler:"i"`
	}
	type hezEthAddrStruct struct {
		I HezEthAddr `meddler:"i"`
	}
	type ethAddrStructNil struct {
		I *ethCommon.Address `meddler:"i"`
	}
	type hezEthAddrStructNil struct {
		I *HezEthAddr `meddler:"i"`
	}

	// Not nil case
	// Insert into DB using ethCommon.Address Scan/Value
	fromEth := ethAddrStruct{
		I: ethCommon.BigToAddress(big.NewInt(73737373)),
	}
	err = meddler.Insert(db, "test", &fromEth)
	assert.NoError(t, err)
	// Read from DB using HezEthAddr Scan/Value
	toHezEth := hezEthAddrStruct{}
	err = meddler.QueryRow(db, &toHezEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, NewHezEthAddr(fromEth.I), toHezEth.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using HezEthAddr Scan/Value
	fromHezEth := hezEthAddrStruct{
		I: NewHezEthAddr(ethCommon.BigToAddress(big.NewInt(3786872586))),
	}
	err = meddler.Insert(db, "test", &fromHezEth)
	assert.NoError(t, err)
	// Read from DB using ethCommon.Address Scan/Value
	toEth := ethAddrStruct{}
	err = meddler.QueryRow(db, &toEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, fromHezEth.I, NewHezEthAddr(toEth.I))

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using ethCommon.Address Scan/Value
	fromEthNil := ethAddrStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromEthNil)
	assert.NoError(t, err)
	// Read from DB using HezEthAddr Scan/Value
	foo := HezEthAddr("foo")
	toHezEthNil := hezEthAddrStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toHezEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toHezEthNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using HezEthAddr Scan/Value
	fromHezEthNil := hezEthAddrStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromHezEthNil)
	assert.NoError(t, err)
	// Read from DB using ethCommon.Address Scan/Value
	fooAddr := ethCommon.BigToAddress(big.NewInt(1))
	toEthNil := ethAddrStructNil{
		I: &fooAddr, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toEthNil.I)
}

func TestHezBJJ(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type bjjStruct struct {
		I babyjub.PublicKeyComp `meddler:"i"`
	}
	type hezBJJStruct struct {
		I HezBJJ `meddler:"i"`
	}
	type bjjStructNil struct {
		I *babyjub.PublicKeyComp `meddler:"i"`
	}
	type hezBJJStructNil struct {
		I *HezBJJ `meddler:"i"`
	}

	// Not nil case
	// Insert into DB using *babyjub.PublicKeyComp Scan/Value
	priv := babyjub.NewRandPrivKey()
	fromBJJ := bjjStruct{
		I: priv.Public().Compress(),
	}
	err = meddler.Insert(db, "test", &fromBJJ)
	assert.NoError(t, err)
	// Read from DB using HezBJJ Scan/Value
	toHezBJJ := hezBJJStruct{}
	err = meddler.QueryRow(db, &toHezBJJ, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, NewHezBJJ(fromBJJ.I), toHezBJJ.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using HezBJJ Scan/Value
	fromHezBJJ := hezBJJStruct{
		I: NewHezBJJ(priv.Public().Compress()),
	}
	err = meddler.Insert(db, "test", &fromHezBJJ)
	assert.NoError(t, err)
	// Read from DB using *babyjub.PublicKeyComp Scan/Value
	toBJJ := bjjStruct{}
	err = meddler.QueryRow(db, &toBJJ, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, fromHezBJJ.I, NewHezBJJ(toBJJ.I))

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using *babyjub.PublicKeyComp Scan/Value
	fromBJJNil := bjjStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromBJJNil)
	assert.NoError(t, err)
	// Read from DB using HezBJJ Scan/Value
	foo := HezBJJ("foo")
	toHezBJJNil := hezBJJStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toHezBJJNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toHezBJJNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using HezBJJ Scan/Value
	fromHezBJJNil := hezBJJStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromHezBJJNil)
	assert.NoError(t, err)
	// Read from DB using *babyjub.PublicKeyComp Scan/Value
	bjjComp := priv.Public().Compress()
	toBJJNil := bjjStructNil{
		I: &bjjComp, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toBJJNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toBJJNil.I)
}

func TestEthSignature(t *testing.T) {
	// Clean DB
	_, err := db.Exec("delete from test")
	assert.NoError(t, err)
	// Example structs
	type ethSignStruct struct {
		I []byte `meddler:"i"`
	}
	type hezEthSignStruct struct {
		I EthSignature `meddler:"i"`
	}
	type hezEthSignStructNil struct {
		I *EthSignature `meddler:"i"`
	}

	// Not nil case
	// Insert into DB using []byte Scan/Value
	s := "someRandomFooForYou"
	fromEth := ethSignStruct{
		I: []byte(s),
	}
	err = meddler.Insert(db, "test", &fromEth)
	assert.NoError(t, err)
	// Read from DB using EthSignature Scan/Value
	toHezEth := hezEthSignStruct{}
	err = meddler.QueryRow(db, &toHezEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, NewEthSignature(fromEth.I), &toHezEth.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using EthSignature Scan/Value
	fromHezEth := hezEthSignStruct{
		I: *NewEthSignature([]byte(s)),
	}
	err = meddler.Insert(db, "test", &fromHezEth)
	assert.NoError(t, err)
	// Read from DB using []byte Scan/Value
	toEth := ethSignStruct{}
	err = meddler.QueryRow(db, &toEth, "select * from test")
	assert.NoError(t, err)
	assert.Equal(t, &fromHezEth.I, NewEthSignature(toEth.I))

	// Nil case
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using []byte Scan/Value
	fromEthNil := ethSignStruct{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromEthNil)
	assert.NoError(t, err)
	// Read from DB using EthSignature Scan/Value
	foo := EthSignature("foo")
	toHezEthNil := hezEthSignStructNil{
		I: &foo, // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toHezEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toHezEthNil.I)
	// Clean DB
	_, err = db.Exec("delete from test")
	assert.NoError(t, err)
	// Insert into DB using EthSignature Scan/Value
	fromHezEthNil := hezEthSignStructNil{
		I: nil,
	}
	err = meddler.Insert(db, "test", &fromHezEthNil)
	assert.NoError(t, err)
	// Read from DB using []byte Scan/Value
	toEthNil := ethSignStruct{
		I: []byte(s), // check that this will be set to nil, not because of not being initialized
	}
	err = meddler.QueryRow(db, &toEthNil, "select * from test")
	assert.NoError(t, err)
	assert.Nil(t, toEthNil.I)
}
