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
var ecomment = fmt.Errorf("comment in parseline")
var enewbatch = fmt.Errorf("newbatch")
var TypeNewBatch common.TxType = "TxTypeNewBatch"

const (
	ILLEGAL token = iota
	WS
	EOF

	IDENT // val
)

type Instruction struct {
	Literal string
	From    string
	To      string
	Amount  uint64
	Fee     uint8
	TokenID common.TokenID
	Type    common.TxType // D: Deposit, T: Transfer, E: ForceExit
}

type Instructions struct {
	Instructions []*Instruction
	Accounts     []string
	TokenIDs     []common.TokenID
}

func (i Instruction) String() string {
	buf := bytes.NewBufferString("")
	switch i.Type {
	case common.TxTypeCreateAccountDeposit:
		fmt.Fprintf(buf, "Type: Create&Deposit, ")
	case common.TxTypeTransfer:
		fmt.Fprintf(buf, "Type: Transfer, ")
	case common.TxTypeForceExit:
		fmt.Fprintf(buf, "Type: ForceExit, ")
	default:
	}
	fmt.Fprintf(buf, "From: %s, ", i.From)
	if i.Type == common.TxTypeTransfer {
		fmt.Fprintf(buf, "To: %s, ", i.To)
	}
	fmt.Fprintf(buf, "Amount: %d, ", i.Amount)
	if i.Type == common.TxTypeTransfer {
		fmt.Fprintf(buf, "Fee: %d, ", i.Fee)
	}
	fmt.Fprintf(buf, "TokenID: %d,\n", i.TokenID)
	return buf.String()
}

func (i Instruction) Raw() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s", i.From)
	if i.Type == common.TxTypeTransfer {
		fmt.Fprintf(buf, "-%s", i.To)
	}
	fmt.Fprintf(buf, " (%d)", i.TokenID)
	if i.Type == common.TxTypeForceExit {
		fmt.Fprintf(buf, "E")
	}
	fmt.Fprintf(buf, ":")
	fmt.Fprintf(buf, " %d", i.Amount)
	if i.Type == common.TxTypeTransfer {
		fmt.Fprintf(buf, " %d", i.Fee)
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

// NewScanner creates a new scanner with the given io.Reader
func NewScanner(r io.Reader) *scanner {
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
	return &Parser{s: NewScanner(r)}
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
	// line can be Deposit:
	//         A (1): 10
	// or Transfer:
	//         A-B (1): 6
	// or Withdraw:
	//         A (1) E: 4
	// or NextBatch:
	// 	> and here the comment

	c := &Instruction{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return nil, errof
	}
	c.Literal += lit
	if lit == "/" {
		_, _ = p.s.r.ReadString('\n')
		return nil, ecomment
	} else if lit == ">" {
		_, _ = p.s.r.ReadString('\n')
		return nil, enewbatch
	}
	c.From = lit

	_, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	if lit == "-" {
		// transfer
		_, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		c.To = lit
		c.Type = common.TxTypeTransfer
		_, lit = p.scanIgnoreWhitespace() // expect (
		c.Literal += lit
		if lit != "(" {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, fmt.Errorf("Expected '(', found '%s'", lit)
		}
	} else {
		c.Type = common.TxTypeCreateAccountDeposit
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

	_, lit = p.scanIgnoreWhitespace() // expect ':' or 'E' (Exit type)
	c.Literal += lit
	if lit == "E" {
		c.Type = common.TxTypeForceExit
		_, lit = p.scanIgnoreWhitespace() // expect ':'
		c.Literal += lit
	}
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

	if c.Type == common.TxTypeTransfer {
		tok, lit = p.scanIgnoreWhitespace()
		c.Literal += lit
		fee, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, err
		}
		if fee > common.MAXFEEPLAN-1 {
			line, _ := p.s.r.ReadString('\n')
			c.Literal += line
			return c, fmt.Errorf("Fee %d can not be bigger than 255", fee)
		}
		c.Fee = uint8(fee)
	}

	if tok == EOF {
		return nil, errof
	}
	return c, nil
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
		if err == ecomment {
			i++
			continue
		}
		if err == enewbatch {
			i++
			inst := &Instruction{Type: TypeNewBatch}
			instructions.Instructions = append(instructions.Instructions, inst)
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
