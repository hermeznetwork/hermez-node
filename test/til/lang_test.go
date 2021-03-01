package til

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

		// token registrations
		AddToken(1)
		AddToken(2)

		// deposits
		Deposit(1) A: 10
		Deposit(2) A: 20
		Deposit(1) B: 5
		CreateAccountDeposit(1) C: 5
		CreateAccountDepositTransfer(1) D-A: 15, 10
		CreateAccountCoordinator(1) E

		// L2 transactions
		Transfer(1) A-B: 6 (1)
		Transfer(1) B-D: 3 (1)
		Transfer(1) A-E: 1 (1)

		// set new batch
		> batch
		AddToken(3)

		DepositTransfer(1) A-B: 15, 10
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
		Exit(1) A: 5 (1)
	`

	parser := newParser(strings.NewReader(s))
	instructions, err := parser.parse()
	require.NoError(t, err)
	assert.Equal(t, 25, len(instructions.instructions))
	assert.Equal(t, 7, len(instructions.users))

	if debug {
		fmt.Println(instructions)
		for _, instruction := range instructions.instructions {
			fmt.Println(instruction.raw())
		}
	}

	assert.Equal(t, TxTypeCreateAccountDepositCoordinator, instructions.instructions[7].Typ)
	assert.Equal(t, TypeNewBatch, instructions.instructions[11].Typ)
	assert.Equal(t, "Deposit(1)User0:20", instructions.instructions[16].raw())
	assert.Equal(t,
		"Type: DepositTransfer, From: A, To: B, DepositAmount: 15, Amount: 10, Fee: 0, TokenID: 1\n",
		instructions.instructions[13].String())
	assert.Equal(t,
		"Type: Transfer, From: User1, To: User0, Amount: 15, Fee: 1, TokenID: 3\n",
		instructions.instructions[19].String())
	assert.Equal(t, "Transfer(2)A-B:15(1)", instructions.instructions[15].raw())
	assert.Equal(t,
		"Type: Transfer, From: A, To: B, Amount: 15, Fee: 1, TokenID: 2\n",
		instructions.instructions[15].String())
	assert.Equal(t, "Exit(1)A:5", instructions.instructions[24].raw())
	assert.Equal(t, "Type: Exit, From: A, Amount: 5, TokenID: 1\n",
		instructions.instructions[24].String())
}

func TestParsePoolTxs(t *testing.T) {
	s := `
		Type: PoolL2
		PoolTransfer(1) A-B: 6 (1)
		PoolTransfer(2) A-B: 3 (3)
		PoolTransfer(1) B-D: 3 (1)
		PoolTransfer(1) C-D: 3 (1)
		PoolExit(1) A: 5 (1)
	`

	parser := newParser(strings.NewReader(s))
	instructions, err := parser.parse()
	require.NoError(t, err)
	assert.Equal(t, 5, len(instructions.instructions))
	assert.Equal(t, 4, len(instructions.users))

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
	assert.Equal(t, "Line 2: Deposit(1)A:: 10\n, err: Can not parse number for Amount: :", err.Error())

	s = `
		Type: Blockchain
		AddToken(1)
		Deposit(1) A: 10 20
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "Line 4: 20, err: Unexpected Blockchain tx type: 20", err.Error())

	s = `
		Type: Blockchain
		Transfer(1) A: 10
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "Line 2: Transfer(1)A:, err: Expected '-', found ':'", err.Error())

	s = `
		Type: Blockchain
		Transfer(1) A B: 10
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "Line 2: Transfer(1)AB, err: Expected '-', found 'B'", err.Error())

	s = `
		Type: Blockchain
		AddToken(1)
		Transfer(1) A-B: 10 (255)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.NoError(t, err)
	s = `
		Type: Blockchain
		Transfer(1) A-B: 10 (256)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"Line 2: Transfer(1)A-B:10(256)\n, err: Fee 256 can not be bigger than 255",
		err.Error())

	// check that the PoolTransfer & Transfer are only accepted in the
	// correct case case (PoolTxs/BlockchainTxs)
	s = `
		Type: PoolL2
		Transfer(1) A-B: 10 (1)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "Line 2: Transfer, err: Unexpected PoolL2 tx type: Transfer", err.Error())
	s = `
		Type: Blockchain
		PoolTransfer(1) A-B: 10 (1)
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"Line 2: PoolTransfer, err: Unexpected Blockchain tx type: PoolTransfer",
		err.Error())

	s = `
		Type: Blockchain
		> btch
	`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"Line 2: >, err: Unexpected '> btch', expected '> batch' or '> block'",
		err.Error())

	// check definition of set Type
	s = `PoolTransfer(1) A-B: 10 (1)`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "Line 1: PoolTransfer, err: Set type not defined", err.Error())
	s = `Type: PoolL1`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"Line 1: Type:, err: Invalid set type: 'PoolL1'. Valid set types: 'Blockchain', 'PoolL2'",
		err.Error())
	s = `Type: PoolL1
		Type: Blockchain`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"Line 1: Type:, err: Invalid set type: 'PoolL1'. Valid set types: 'Blockchain', 'PoolL2'",
		err.Error())
	s = `Type: PoolL2
		Type: Blockchain`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t,
		"Line 2: Instruction of 'Type: Blockchain' when there is already a previous "+
			"instruction 'Type: PoolL2' defined", err.Error())

	s = `Type: Blockchain
		AddToken(1)
		AddToken(0)
		`
	parser = newParser(strings.NewReader(s))
	_, err = parser.parse()
	assert.Equal(t, "Line 3: AddToken can not register TokenID 0", err.Error())
}
