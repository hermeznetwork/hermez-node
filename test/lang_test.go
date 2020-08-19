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
		A (1): 10
		A (2): 20
		B (1): 5
		A-B (1): 6
		B-C (1): 3
		C-A (1): 3
		A-B (2): 15
		User0   (1): 20
		User1 (3) : 20
		User0-User1 (1): 15
		User1-User0 (3): 15
	`
	parser := NewParser(strings.NewReader(s))
	instructions, err := parser.Parse()
	assert.Nil(t, err)
	assert.Equal(t, 11, len(instructions.Instructions))
	assert.Equal(t, 5, len(instructions.Accounts))
	assert.Equal(t, 3, len(instructions.TokenIDs))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.Instructions {
			fmt.Println(instruction.Raw())
		}
	}

	assert.Equal(t, "User0 (1): 20", instructions.Instructions[7].Raw())
	assert.Equal(t, "Type: Deposit, From: User0, Amount: 20, TokenID: 1,\n", instructions.Instructions[7].String())
	assert.Equal(t, "User0-User1 (1): 15", instructions.Instructions[9].Raw())
	assert.Equal(t, "Type: Transfer, From: User0, To: User1, Amount: 15, TokenID: 1,\n", instructions.Instructions[9].String())
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
}
