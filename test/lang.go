package test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/hermeznetwork/hermez-node/common"
)

var eof = rune(0)
var errof = fmt.Errorf("eof in parseline")
var errComment = fmt.Errorf("comment in parseline")
var newEvent = fmt.Errorf("newEvent")

// TypeNewBatch is used for testing purposes only, and represents the
// common.TxType of a new batch
var TypeNewBatch common.TxType = "TxTypeNewBatch"

// TypeNewBlock is used for testing purposes only, and represents the
// common.TxType of a new ethereum block
var TypeNewBlock common.TxType = "TxTypeNewBlock"

//nolint
const (
	ILLEGAL token = iota
	WS
	EOF

	IDENT // val
)

// Instruction is the data structure that represents one line of code
type Instruction struct {
	Literal    string
	From       string
	To         string
	Amount     uint64
	LoadAmount uint64
	Fee        uint8
	TokenID    common.TokenID
	Type       common.TxType // D: Deposit, T: Transfer, E: ForceExit
}

// Instructions contains the full Set of Instructions representing a full code
type Instructions struct {
	Instructions []*Instruction
	Accounts     []string
	TokenIDs     []common.TokenID
}

func (i Instruction) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Type: %s, ", i.Type)
	fmt.Fprintf(buf, "From: %s, ", i.From)
	if i.Type == common.TxTypeTransfer ||
		i.Type == common.TxTypeDepositTransfer ||
		i.Type == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "To: %s, ", i.To)
	}
	if i.Type == common.TxTypeDepositTransfer ||
		i.Type == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "LoadAmount: %d, ", i.LoadAmount)
	}
	fmt.Fprintf(buf, "Amount: %d, ", i.Amount)
	if i.Type == common.TxTypeTransfer ||
		i.Type == common.TxTypeDepositTransfer ||
		i.Type == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "Fee: %d, ", i.Fee)
	}
	fmt.Fprintf(buf, "TokenID: %d\n", i.TokenID)
	return buf.String()
}

// Raw returns a string with the raw representation of the Instruction
func (i Instruction) Raw() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s", i.Type)
	fmt.Fprintf(buf, "(%d)", i.TokenID)
	fmt.Fprintf(buf, "%s", i.From)
	if i.Type == common.TxTypeTransfer ||
		i.Type == common.TxTypeDepositTransfer ||
		i.Type == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "-%s", i.To)
	}
	fmt.Fprintf(buf, ":")
	if i.Type == common.TxTypeDepositTransfer ||
		i.Type == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "%d,", i.LoadAmount)
	}
	fmt.Fprintf(buf, "%d", i.Amount)
	if i.Type == common.TxTypeTransfer {
		fmt.Fprintf(buf, "(%d)", i.Fee)
	}
	return buf.String()
}

type token int

type scanner struct {
	r *bufio.Reader
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isComment(ch rune) bool {
	return ch == '/'
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// newScanner creates a new scanner with the given io.Reader
func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *scanner) unread() {
	_ = s.r.UnreadRune()
}

// scan returns the token and literal string of the current value
func (s *scanner) scan() (tok token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		// space
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) || isDigit(ch) {
		// letter/digit
		s.unread()
		return s.scanIndent()
	} else if isComment(ch) {
		// comment
		s.unread()
		return s.scanIndent()
	}

	if ch == eof {
		return EOF, ""
	}

	return ILLEGAL, string(ch)
}

func (s *scanner) scanWhitespace() (token token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	return WS, buf.String()
}

func (s *scanner) scanIndent() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	if len(buf.String()) == 1 {
		return token(rune(buf.String()[0])), buf.String()
	}
	return IDENT, buf.String()
}

// Parser defines the parser
type Parser struct {
	s   *scanner
	buf struct {
		tok token
		lit string
		n   int
	}
}

// NewParser creates a new parser from a io.Reader
func NewParser(r io.Reader) *Parser {
	return &Parser{s: newScanner(r)}
}

func (p *Parser) scan() (tok token, lit string) {
	// if there is a token in the buffer return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.scan()

	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *Parser) scanIgnoreWhitespace() (tok token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// parseLine parses the current line
func (p *Parser) parseLine() (*Instruction, error) {
	c := &Instruction{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return nil, errof
	}
	c.Literal += lit
	if lit == "/" {
		_, _ = p.s.r.ReadString('\n')
		return nil, errComment
	} else if lit == ">" {
		_, lit = p.scanIgnoreWhitespace()
		if lit == "batch" {
			_, _ = p.s.r.ReadString('\n')
			return &Instruction{Type: TypeNewBatch}, newEvent
		} else if lit == "block" {
			_, _ = p.s.r.ReadString('\n')
			return &Instruction{Type: TypeNewBlock}, newEvent
		} else {
			return c, fmt.Errorf("Unexpected '> %s', expected '> batch' or '> block'", lit)
		}
	}
	transfering := false
	switch lit {
	case "Deposit":
		c.Type = common.TxTypeDeposit
	case "Exit", "PoolExit":
		c.Type = common.TxTypeExit
	case "Transfer", "PoolTransfer":
		c.Type = common.TxTypeTransfer
		transfering = true
	case "CreateAccountDeposit":
		c.Type = common.TxTypeCreateAccountDeposit
	case "CreateAccountDepositTransfer":
		c.Type = common.TxTypeCreateAccountDepositTransfer
		transfering = true
	case "DepositTransfer":
		c.Type = common.TxTypeDepositTransfer
		transfering = true
	case "ForceTransfer":
		c.Type = common.TxTypeForceTransfer
	case "ForceExit":
		c.Type = common.TxTypeForceExit
	default:
		return c, fmt.Errorf("Unexpected tx type: %s", lit)
	}

	if err := p.expectChar(c, "("); err != nil {
		return c, err
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	tidI, err := strconv.Atoi(lit)
	if err != nil {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, err
	}
	c.TokenID = common.TokenID(tidI)
	if err := p.expectChar(c, ")"); err != nil {
		return c, err
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	c.From = lit
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	if lit == "-" {
		if !transfering {
			return c, fmt.Errorf("To defined, but not type {Transfer, CreateAccountDepositTransfer, DepositTransfer}")
		}
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		c.To = lit
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
	}
	if lit != ":" {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, fmt.Errorf("Expected ':', found '%s'", lit)
	}
	if c.Type == common.TxTypeDepositTransfer ||
		c.Type == common.TxTypeCreateAccountDepositTransfer {
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		loadAmount, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, err
		}
		c.LoadAmount = uint64(loadAmount)
		if err := p.expectChar(c, ","); err != nil {
			return c, err
		}
	}
	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	amount, err := strconv.Atoi(lit)
	if err != nil {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, err
	}
	c.Amount = uint64(amount)
	if transfering {
		if err := p.expectChar(c, "("); err != nil {
			return c, err
		}
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		fee, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, err
		}
		if fee > common.MaxFeePlan-1 {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, fmt.Errorf("Fee %d can not be bigger than 255", fee)
		}
		c.Fee = uint8(fee)

		if err := p.expectChar(c, ")"); err != nil {
			return c, err
		}
	}

	if tok == EOF {
		return nil, errof
	}
	return c, nil
}

func (p *Parser) expectChar(c *Instruction, ch string) error {
	_, lit := p.scanIgnoreWhitespace()
	c.Literal += lit
	if lit != ch {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return fmt.Errorf("Expected '%s', found '%s'", ch, lit)
	}
	return nil
}

func idxTokenIDToString(idx string, tid common.TokenID) string {
	return idx + strconv.Itoa(int(tid))
}

// Parse parses through reader
func (p *Parser) Parse() (Instructions, error) {
	var instructions Instructions
	i := 0
	accounts := make(map[string]bool)
	tokenids := make(map[common.TokenID]bool)
	for {
		instruction, err := p.parseLine()
		if err == errof {
			break
		}
		if err == errComment {
			i++
			continue
		}
		if err == newEvent {
			i++
			instructions.Instructions = append(instructions.Instructions, instruction)
			continue
		}
		if err != nil {
			return instructions, fmt.Errorf("error parsing line %d: %s, err: %s", i, instruction.Literal, err.Error())
		}
		instructions.Instructions = append(instructions.Instructions, instruction)
		accounts[idxTokenIDToString(instruction.From, instruction.TokenID)] = true
		if instruction.Type == common.TxTypeTransfer { // type: Transfer
			accounts[idxTokenIDToString(instruction.To, instruction.TokenID)] = true
		}
		tokenids[instruction.TokenID] = true
		i++
	}
	for a := range accounts {
		instructions.Accounts = append(instructions.Accounts, a)
	}
	sort.Strings(instructions.Accounts)
	for tid := range tokenids {
		instructions.TokenIDs = append(instructions.TokenIDs, tid)
	}
	return instructions, nil
}
