package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var debug = false

func TestParseBlockchainTxs(t *testing.T) {
	s := `
		// deposits
		Deposit(1) A: 10
		Deposit(2) A: 20
		Deposit(1) B: 5
		CreateAccountDeposit(1) C: 5
		CreateAccountDepositTransfer(1) D-A: 15, 10 (3)

		// L2 transactions
		Transfer(1) A-B: 6 (1)
		Transfer(1) B-D: 3 (1)

		// set new batch
		> batch

		DepositTransfer(1) A-B: 15, 10 (1)
		Transfer(1) C-A : 3 (1)
		Transfer(2) A-B: 15 (1)

		Deposit(1) User0: 20
		Deposit(3) User1: 20
		Transfer(1) User0-User1: 15 (1)
		Transfer(3) User1-User0: 15 (1)

		> batch

		Transfer(1) User1-User0: 1 (1)

		> batch
		> block

		// Exits
		Exit(1) A: 5
	`

	parser := NewParser(strings.NewReader(s))
	instructions, err := parser.Parse()
	assert.Nil(t, err)
	assert.Equal(t, 20, len(instructions.Instructions))
	assert.Equal(t, 10, len(instructions.Accounts))
	assert.Equal(t, 3, len(instructions.TokenIDs))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.Instructions {
			fmt.Println(instruction.Raw())
		}
	}

	assert.Equal(t, TypeNewBatch, instructions.Instructions[7].Type)
	assert.Equal(t, "Deposit(1)User0:20", instructions.Instructions[11].Raw())
	assert.Equal(t, "Type: DepositTransfer, From: A, To: B, LoadAmount: 15, Amount: 10, Fee: 1, TokenID: 1\n", instructions.Instructions[8].String())
	assert.Equal(t, "Type: Transfer, From: User1, To: User0, Amount: 15, Fee: 1, TokenID: 3\n", instructions.Instructions[14].String())
	assert.Equal(t, "Transfer(2)A-B:15(1)", instructions.Instructions[10].Raw())
	assert.Equal(t, "Type: Transfer, From: A, To: B, Amount: 15, Fee: 1, TokenID: 2\n", instructions.Instructions[10].String())
	assert.Equal(t, "Exit(1)A:5", instructions.Instructions[19].Raw())
	assert.Equal(t, "Type: Exit, From: A, Amount: 5, TokenID: 1\n", instructions.Instructions[19].String())
}

func TestParsePoolTxs(t *testing.T) {
	s := `
		PoolTransfer(1) A-B: 6 (1)
		PoolTransfer(2) A-B: 3 (3)
		PoolTransfer(1) B-D: 3 (1)
		PoolTransfer(1) C-D: 3 (1)
		Exit(1) A: 5
	`

	parser := NewParser(strings.NewReader(s))
	instructions, err := parser.Parse()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(instructions.Instructions))
	assert.Equal(t, 6, len(instructions.Accounts))
	assert.Equal(t, 2, len(instructions.TokenIDs))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.Instructions {
			fmt.Println(instruction.Raw())
		}
	}

	assert.Equal(t, "Transfer(1)A-B:6(1)", instructions.Instructions[0].Raw())
	assert.Equal(t, "Transfer(2)A-B:3(3)", instructions.Instructions[1].Raw())
	assert.Equal(t, "Transfer(1)B-D:3(1)", instructions.Instructions[2].Raw())
	assert.Equal(t, "Transfer(1)C-D:3(1)", instructions.Instructions[3].Raw())
	assert.Equal(t, "Exit(1)A:5", instructions.Instructions[4].Raw())
}

func TestParseErrors(t *testing.T) {
	s := "Deposit(1) A:: 10"
	parser := NewParser(strings.NewReader(s))
	_, err := parser.Parse()
	assert.Equal(t, "error parsing line 0: Deposit(1)A:: 10, err: strconv.Atoi: parsing \":\": invalid syntax", err.Error())

	s = "Deposit(1) A: 10 20"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 1: 20, err: Unexpected tx type: 20", err.Error())

	s = "Transfer(1) A B: 10"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 0: Transfer(1)AB: 10, err: Expected ':', found 'B'", err.Error())

	s = "Transfer(1) A-B: 10 (255)"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Nil(t, err)
	s = "Transfer(1) A-B: 10 (256)"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 0: Transfer(1)A-B:10(256), err: Fee 256 can not be bigger than 255", err.Error())

	s = "> btch"
	parser = NewParser(strings.NewReader(s))
	_, err = parser.Parse()
	assert.Equal(t, "error parsing line 0: >, err: Unexpected '> btch', expected '> batch' or '> block'", err.Error())
}
