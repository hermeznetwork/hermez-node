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

const (
	ILLEGAL Token = iota
	WS
	EOF

	IDENT // val
)

type Instruction struct {
	Literal string
	From    string
	To      string
	Amount  uint64
	TokenID common.TokenID
	Type    int // 0: Deposit, 1: Transfer
}

type Instructions struct {
	Instructions []*Instruction
	Accounts     []string
	TokenIDs     []common.TokenID
}

func (i Instruction) String() string {
	buf := bytes.NewBufferString("")
	switch i.Type {
	case 0:
		fmt.Fprintf(buf, "Type: Deposit, ")
	case 1:
		fmt.Fprintf(buf, "Type: Transfer, ")
	default:
	}
	fmt.Fprintf(buf, "From: %s, ", i.From)
	if i.Type == 1 {
		fmt.Fprintf(buf, "To: %s, ", i.To)
	}
	fmt.Fprintf(buf, "Amount: %d, ", i.Amount)
	fmt.Fprintf(buf, "TokenID: %d,\n", i.TokenID)
	return buf.String()
}

func (i Instruction) Raw() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s", i.From)
	if i.Type == 1 {
		fmt.Fprintf(buf, "-%s", i.To)
	}
	fmt.Fprintf(buf, " (%d):", i.TokenID)
	fmt.Fprintf(buf, " %d", i.Amount)
	return buf.String()
}

type Token int

type Scanner struct {
	r *bufio.Reader
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// NewScanner creates a new Scanner with the given io.Reader
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

// scan returns the Token and literal string of the current value
func (s *Scanner) scan() (tok Token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		// space
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) || isDigit(ch) {
		// letter/digit
		s.unread()
		return s.scanIndent()
	}

	if ch == eof {
		return EOF, ""
	}

	return ILLEGAL, string(ch)
}

func (s *Scanner) scanWhitespace() (token Token, lit string) {
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

func (s *Scanner) scanIndent() (tok Token, lit string) {
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
		return Token(rune(buf.String()[0])), buf.String()
	}
	return IDENT, buf.String()
}

// Parser defines the parser
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token
		lit string
		n   int
	}
}

// NewParser creates a new parser from a io.Reader
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) scan() (tok Token, lit string) {
	// if there is a token in the buffer return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.scan()

	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// parseLine parses the current line
func (p *Parser) parseLine() (*Instruction, error) {
	/*
		line can be Deposit:
			A (1): 10
		or Transfer:
			A-B (1): 6
	*/
	c := &Instruction{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return nil, errof
	}
	c.Literal += lit
	c.From = lit

	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	if lit == "-" {
		// transfer
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		c.To = lit
		c.Type = 1
		_, lit = p.scanIgnoreWhitespace() // expect (
		c.Literal += lit
		if lit != "(" {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, fmt.Errorf("Expected '(', found '%s'", lit)
		}
	} else {
		c.Type = 0
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
	_, lit = p.scanIgnoreWhitespace() // expect )
	c.Literal += lit
	if lit != ")" {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, fmt.Errorf("Expected ')', found '%s'", lit)
	}

	_, lit = p.scanIgnoreWhitespace() // expect :
	c.Literal += lit
	if lit != ":" {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, fmt.Errorf("Expected ':', found '%s'", lit)
	}
	tok, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	amount, err := strconv.Atoi(lit)
	if err != nil {
		line, _ := p.s.r.ReadString('\n')
		c.Literal += line
		return c, err
	}
	c.Amount = uint64(amount)

	if tok == EOF {
		return nil, errof
	}
	return c, nil
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
		if err != nil {
			return instructions, fmt.Errorf("error parsing line %d: %s, err: %s", i, instruction.Literal, err.Error())
		}
		instructions.Instructions = append(instructions.Instructions, instruction)
		accounts[instruction.From] = true
		if instruction.Type == 1 { // type: Transfer
			accounts[instruction.To] = true
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
