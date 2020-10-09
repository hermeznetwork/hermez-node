package transakcio

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var debug = false

func TestParseBlockchainTxs(t *testing.T) {
	s := `
		Type: Blockchain
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

	parser := newParser(strings.NewReader(s))
	instructions, err := parser.parse()
	require.Nil(t, err)
	assert.Equal(t, 20, len(instructions.instructions))
	assert.Equal(t, 6, len(instructions.accounts))
	assert.Equal(t, 3, len(instructions.tokenIDs))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.instructions {
			fmt.Println(instruction.raw())
		}
	}

	assert.Equal(t, typeNewBatch, instructions.instructions[7].typ)
	assert.Equal(t, "Deposit(1)User0:20", instructions.instructions[11].raw())
	assert.Equal(t, "Type: DepositTransfer, From: A, To: B, LoadAmount: 15, Amount: 10, Fee: 1, TokenID: 1\n", instructions.instructions[8].String())
	assert.Equal(t, "Type: Transfer, From: User1, To: User0, Amount: 15, Fee: 1, TokenID: 3\n", instructions.instructions[14].String())
	assert.Equal(t, "Transfer(2)A-B:15(1)", instructions.instructions[10].raw())
	assert.Equal(t, "Type: Transfer, From: A, To: B, Amount: 15, Fee: 1, TokenID: 2\n", instructions.instructions[10].String())
	assert.Equal(t, "Exit(1)A:5", instructions.instructions[19].raw())
	assert.Equal(t, "Type: Exit, From: A, Amount: 5, TokenID: 1\n", instructions.instructions[19].String())
}

func TestParsePoolTxs(t *testing.T) {
	s := `
		Type: PoolL2
		PoolTransfer(1) A-B: 6 (1)
		PoolTransfer(2) A-B: 3 (3)
		PoolTransfer(1) B-D: 3 (1)
		PoolTransfer(1) C-D: 3 (1)
		PoolExit(1) A: 5
	`

	parser := newParser(strings.NewReader(s))
	instructions, err := parser.parse()
	require.Nil(t, err)
	assert.Equal(t, 5, len(instructions.instructions))
	assert.Equal(t, 4, len(instructions.accounts))
	assert.Equal(t, 2, len(instructions.tokenIDs))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.instructions {
			fmt.Println(instruction.raw())
		}
	}

	assert.Equal(t, "Transfer(1)A-B:6(1)", instructions.instructions[0].raw())
	assert.Equal(t, "Transfer(2)A-B:3(3)", instructions.instructions[1].raw())
	assert.Equal(t, "Transfer(1)B-D:3(1)", instructions.instructions[2].raw())
	assert.Equal(t, "Transfer(1)C-D:3(1)", instructions.instructions[3].raw())
	assert.Equal(t, "Exit(1)A:5", instructions.instructions[4].raw())
}

func TestParseErrors(t *testing.T) {
	s := `
		Type: Blockchain
		Deposit(1) A:: 10
	`
	parser := newParser(strings.NewReader(s))
	_, err := parser.parse()
	assert.Equal(t, "error parsing line 1: Deposit(1)A:: 10\n, err: strconv.Atoi: parsing \":\": invalid syntax", err.Error())

	s = `
		Type: Blockchain
		Deposit(1) A: 10 20
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 2: 20, err: Unexpected tx type: 20", err.Error())

	s = `
		Type: Blockchain
		Transfer(1) A: 10
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 1: Transfer(1)A:, err: Expected '-', found ':'", err.Error())

	s = `
		Type: Blockchain
		Transfer(1) A B: 10
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 1: Transfer(1)AB, err: Expected '-', found 'B'", err.Error())

	s = `
		Type: Blockchain
		Transfer(1) A-B: 10 (255)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Nil(t, err)
	s = `
		Type: Blockchain
		Transfer(1) A-B: 10 (256)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 1: Transfer(1)A-B:10(256)\n, err: Fee 256 can not be bigger than 255", err.Error())

	// check that the PoolTransfer & Transfer are only accepted in the
	// correct case case (PoolTxs/BlockchainTxs)
	s = `
		Type: PoolL2
		Transfer(1) A-B: 10 (1)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 1: Transfer, err: Unexpected 'Transfer' in a non Blockchain set", err.Error())
	s = `
		Type: Blockchain
		PoolTransfer(1) A-B: 10 (1)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 1: PoolTransfer, err: Unexpected 'PoolTransfer' in a non PoolL2 set", err.Error())

	s = `
		Type: Blockchain
		> btch
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 1: >, err: Unexpected '> btch', expected '> batch' or '> block'", err.Error())

	// check definition of set Type
	s = `PoolTransfer(1) A-B: 10 (1)`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 0: PoolTransfer, err: Set type not defined", err.Error())
	s = `Type: PoolL1`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 0: Type:, err: Invalid set type: 'PoolL1'. Valid set types: 'Blockchain', 'PoolL2'", err.Error())
	s = `Type: PoolL1
		Type: Blockchain`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "error parsing line 0: Type:, err: Invalid set type: 'PoolL1'. Valid set types: 'Blockchain', 'PoolL2'", err.Error())
}
