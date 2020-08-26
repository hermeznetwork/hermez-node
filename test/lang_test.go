package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var debug = false

func TestParse(t *testing.T) {
	s := `
		// deposits
		A (1): 10
		A (2): 20
		B (1): 5

		// L2 transactions
		A-B (1): 6 1
		B-C (1): 3 1

		// set new batch, label does not affect
		> batch1

		C-A (1): 3 1
		A-B (2): 15 1

		User0   (1): 20
		User1 (3) : 20
		User0-User1 (1): 15 1
		User1-User0 (3): 15 1

		// Exits
		A (1) E: 5
	`
	parser := NewParser(strings.NewReader(s))
	instructions, err := parser.Parse()
	assert.Nil(t, err)
	assert.Equal(t, 13, len(instructions.Instructions))
	// assert.Equal(t, 5, len(instructions.Accounts))
	fmt.Println(instructions.Accounts)
	assert.Equal(t, 3, len(instructions.TokenIDs))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.Instructions {
			fmt.Println(instruction.Raw())
		}
	}

	assert.Equal(t, TypeNewBatch, instructions.Instructions[5].Type)
	assert.Equal(t, "User0 (1): 20", instructions.Instructions[8].Raw())
	assert.Equal(t, "Type: Create&Deposit, From: User0, Amount: 20, TokenID: 1,\n", instructions.Instructions[8].String())
	assert.Equal(t, "User0-User1 (1): 15 1", instructions.Instructions[10].Raw())
	assert.Equal(t, "Type: Transfer, From: User0, To: User1, Amount: 15, Fee: 1, TokenID: 1,\n", instructions.Instructions[10].String())
	assert.Equal(t, "A (1)E: 5", instructions.Instructions[12].Raw())
	assert.Equal(t, "Type: ForceExit, From: A, Amount: 5, TokenID: 1,\n", instructions.Instructions[12].String())
}

func TestParseErrors(t *testing.T) {
	s := "A (1):: 10"
	parser := NewParser(strings.NewReader(s))
	_, err := parser.Parse()
	assert.Equal(t, "error parsing line 0: A(1):: 10, err: strconv.Atoi: parsing \":\": invalid syntax", err.Error())

	s = "A (1): 10 20"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 1: 20, err: strconv.Atoi: parsing \"\": invalid syntax", err.Error())

	s = "A B (1): 10"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 0: AB(1): 10, err: strconv.Atoi: parsing \"(\": invalid syntax", err.Error())

	s = "A-B (1): 10 255"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Nil(t, err)
	s = "A-B (1): 10 256"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 0: A-B(1):10256, err: Fee 256 can not be bigger than 255", err.Error())
}
