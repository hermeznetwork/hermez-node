package til

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

var eof = rune(0)
var errof = fmt.Errorf("eof in parseline")
var commentLine = fmt.Errorf("comment in parseline") //nolint:golint
var newEventLine = fmt.Errorf("newEventLine")        //nolint:golint
var setTypeLine = fmt.Errorf("setTypeLine")          //nolint:golint

// setType defines the type of the set
type setType string

// setTypeBlockchain defines the type 'Blockchain' of the set
var setTypeBlockchain = setType("Blockchain")

// setTypePoolL2 defines the type 'PoolL2' of the set
var setTypePoolL2 = setType("PoolL2")

// typeNewBatch is used for testing purposes only, and represents the
// common.TxType of a new batch
var typeNewBatch common.TxType = "InstrTypeNewBatch"

// typeNewBatchL1 is used for testing purposes only, and represents the
// common.TxType of a new batch
var typeNewBatchL1 common.TxType = "InstrTypeNewBatchL1"

// typeNewBlock is used for testing purposes only, and represents the
// common.TxType of a new ethereum block
var typeNewBlock common.TxType = "InstrTypeNewBlock"

// typeAddToken is used for testing purposes only, and represents the
// common.TxType of a new Token regsitration
// It has 'nolint:gosec' as the string 'Token' triggers gosec as a potential leaked Token (which is not the case)
var typeAddToken common.TxType = "InstrTypeAddToken" //nolint:gosec

var txTypeCreateAccountDepositCoordinator common.TxType = "TypeCreateAccountDepositCoordinator"

//nolint
const (
	ILLEGAL token = iota
	WS
	EOF

	IDENT // val
)

// instruction is the data structure that represents one line of code
type instruction struct {
	lineNum    int
	literal    string
	from       string
	to         string
	amount     uint64
	loadAmount uint64
	fee        uint8
	tokenID    common.TokenID
	typ        common.TxType // D: Deposit, T: Transfer, E: ForceExit
}

// parsedSet contains the full Set of Instructions representing a full code
type parsedSet struct {
	typ          setType
	instructions []instruction
	users        []string
}

func (i instruction) String() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Type: %s, ", i.typ)
	fmt.Fprintf(buf, "From: %s, ", i.from)
	if i.typ == common.TxTypeTransfer ||
		i.typ == common.TxTypeDepositTransfer ||
		i.typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "To: %s, ", i.to)
	}

	if i.typ == common.TxTypeDeposit ||
		i.typ == common.TxTypeDepositTransfer ||
		i.typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "LoadAmount: %d, ", i.loadAmount)
	}
	if i.typ != common.TxTypeDeposit {
		fmt.Fprintf(buf, "Amount: %d, ", i.amount)
	}
	if i.typ == common.TxTypeTransfer ||
		i.typ == common.TxTypeDepositTransfer ||
		i.typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "Fee: %d, ", i.fee)
	}
	fmt.Fprintf(buf, "TokenID: %d\n", i.tokenID)
	return buf.String()
}

// Raw returns a string with the raw representation of the Instruction
func (i instruction) raw() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "%s", i.typ)
	fmt.Fprintf(buf, "(%d)", i.tokenID)
	fmt.Fprintf(buf, "%s", i.from)
	if i.typ == common.TxTypeTransfer ||
		i.typ == common.TxTypeDepositTransfer ||
		i.typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "-%s", i.to)
	}
	fmt.Fprintf(buf, ":")
	if i.typ == common.TxTypeDeposit ||
		i.typ == common.TxTypeDepositTransfer ||
		i.typ == common.TxTypeCreateAccountDepositTransfer {
		fmt.Fprintf(buf, "%d", i.loadAmount)
	}
	if i.typ != common.TxTypeDeposit {
		fmt.Fprintf(buf, "%d", i.amount)
	}
	if i.typ == common.TxTypeTransfer {
		fmt.Fprintf(buf, "(%d)", i.fee)
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

// parser defines the parser
type parser struct {
	s   *scanner
	buf struct {
		tok token
		lit string
		n   int
	}
}

// newParser creates a new parser from a io.Reader
func newParser(r io.Reader) *parser {
	return &parser{s: newScanner(r)}
}

func (p *parser) scan() (tok token, lit string) {
	// if there is a token in the buffer return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.scan()

	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *parser) scanIgnoreWhitespace() (tok token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// parseLine parses the current line
func (p *parser) parseLine(setType setType) (*instruction, error) {
	c := &instruction{}
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return nil, errof
	}
	c.literal += lit
	if lit == "/" {
		_, _ = p.s.r.ReadString('\n')
		return nil, commentLine
	} else if lit == ">" {
		if setType == setTypePoolL2 {
			return c, fmt.Errorf("Unexpected '>' at PoolL2Txs set")
		}
		_, lit = p.scanIgnoreWhitespace()
		if lit == "batch" {
			_, _ = p.s.r.ReadString('\n')
			return &instruction{typ: typeNewBatch}, newEventLine
		} else if lit == "batchL1" {
			_, _ = p.s.r.ReadString('\n')
			return &instruction{typ: typeNewBatchL1}, newEventLine
		} else if lit == "block" {
			_, _ = p.s.r.ReadString('\n')
			return &instruction{typ: typeNewBlock}, newEventLine
		} else {
			return c, fmt.Errorf("Unexpected '> %s', expected '> batch' or '> block'", lit)
		}
	} else if lit == "Type" {
		if err := p.expectChar(c, ":"); err != nil {
			return c, err
		}
		_, lit = p.scanIgnoreWhitespace()
		if lit == "Blockchain" {
			return &instruction{typ: "Blockchain"}, setTypeLine
		} else if lit == "PoolL2" {
			return &instruction{typ: "PoolL2"}, setTypeLine
		} else {
			return c, fmt.Errorf("Invalid set type: '%s'. Valid set types: 'Blockchain', 'PoolL2'", lit)
		}
	} else if lit == "AddToken" {
		if err := p.expectChar(c, "("); err != nil {
			return c, err
		}
		_, lit = p.scanIgnoreWhitespace()
		c.literal += lit
		tidI, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.literal += line
			return c, err
		}
		c.tokenID = common.TokenID(tidI)
		if err := p.expectChar(c, ")"); err != nil {
			return c, err
		}
		c.typ = typeAddToken
		line, _ := p.s.r.ReadString('\n')
		c.literal += line
		return c, newEventLine
	}

	if setType == "" {
		return c, fmt.Errorf("Set type not defined")
	}
	transferring := false

	if setType == setTypeBlockchain {
		switch lit {
		case "Deposit":
			c.typ = common.TxTypeDeposit
		case "Exit":
			c.typ = common.TxTypeExit
		case "Transfer":
			c.typ = common.TxTypeTransfer
			transferring = true
		case "CreateAccountDeposit":
			c.typ = common.TxTypeCreateAccountDeposit
		case "CreateAccountDepositTransfer":
			c.typ = common.TxTypeCreateAccountDepositTransfer
			transferring = true
		case "CreateAccountDepositCoordinator":
			c.typ = txTypeCreateAccountDepositCoordinator
			// transferring is false, as the Coordinator tx transfer will be 0
		case "DepositTransfer":
			c.typ = common.TxTypeDepositTransfer
			transferring = true
		case "ForceTransfer":
			c.typ = common.TxTypeForceTransfer
		case "ForceExit":
			c.typ = common.TxTypeForceExit
		default:
			return c, fmt.Errorf("Unexpected Blockchain tx type: %s", lit)
		}
	} else if setType == setTypePoolL2 {
		switch lit {
		case "PoolTransfer":
			c.typ = common.TxTypeTransfer
			transferring = true
		case "PoolTransferToEthAddr":
			c.typ = common.TxTypeTransferToEthAddr
			transferring = true
		case "PoolTransferToBJJ":
			c.typ = common.TxTypeTransferToBJJ
			transferring = true
		case "PoolExit":
			c.typ = common.TxTypeExit
		default:
			return c, fmt.Errorf("Unexpected PoolL2 tx type: %s", lit)
		}
	} else {
		return c, fmt.Errorf("Invalid set type: '%s'. Valid set types: 'Blockchain', 'PoolL2'", setType)
	}

	if err := p.expectChar(c, "("); err != nil {
		return c, err
	}
	_, lit = p.scanIgnoreWhitespace()
	c.literal += lit
	tidI, err := strconv.Atoi(lit)
	if err != nil {
		line, _ := p.s.r.ReadString('\n')
		c.literal += line
		return c, err
	}
	c.tokenID = common.TokenID(tidI)
	if err := p.expectChar(c, ")"); err != nil {
		return c, err
	}
	_, lit = p.scanIgnoreWhitespace()
	c.literal += lit
	c.from = lit
	if c.typ == txTypeCreateAccountDepositCoordinator {
		line, _ := p.s.r.ReadString('\n')
		c.literal += line
		return c, nil
	}
	_, lit = p.scanIgnoreWhitespace()
	c.literal += lit
	if transferring {
		if lit != "-" {
			return c, fmt.Errorf("Expected '-', found '%s'", lit)
		}
		_, lit = p.scanIgnoreWhitespace()
		c.literal += lit
		c.to = lit
		_, lit = p.scanIgnoreWhitespace()
		c.literal += lit
	}
	if lit != ":" {
		line, _ := p.s.r.ReadString('\n')
		c.literal += line
		return c, fmt.Errorf("Expected ':', found '%s'", lit)
	}
	if c.typ == common.TxTypeDepositTransfer ||
		c.typ == common.TxTypeCreateAccountDepositTransfer {
		// deposit case
		_, lit = p.scanIgnoreWhitespace()
		c.literal += lit
		loadAmount, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.literal += line
			return c, err
		}
		c.loadAmount = uint64(loadAmount)
		if err := p.expectChar(c, ","); err != nil {
			return c, err
		}
	}
	_, lit = p.scanIgnoreWhitespace()
	c.literal += lit
	amount, err := strconv.Atoi(lit)
	if err != nil {
		line, _ := p.s.r.ReadString('\n')
		c.literal += line
		return c, err
	}
	if c.typ == common.TxTypeDeposit ||
		c.typ == common.TxTypeCreateAccountDeposit {
		c.loadAmount = uint64(amount)
	} else {
		c.amount = uint64(amount)
	}
	if transferring {
		if err := p.expectChar(c, "("); err != nil {
			return c, err
		}
		_, lit = p.scanIgnoreWhitespace()
		c.literal += lit
		fee, err := strconv.Atoi(lit)
		if err != nil {
			line, _ := p.s.r.ReadString('\n')
			c.literal += line
			return c, err
		}
		if fee > common.MaxFeePlan-1 {
			line, _ := p.s.r.ReadString('\n')
			c.literal += line
			return c, fmt.Errorf("Fee %d can not be bigger than 255", fee)
		}
		c.fee = uint8(fee)

		if err := p.expectChar(c, ")"); err != nil {
			return c, err
		}
	}

	if tok == EOF {
		return nil, errof
	}
	return c, nil
}

func (p *parser) expectChar(c *instruction, ch string) error {
	_, lit := p.scanIgnoreWhitespace()
	c.literal += lit
	if lit != ch {
		line, _ := p.s.r.ReadString('\n')
		c.literal += line
		return fmt.Errorf("Expected '%s', found '%s'", ch, lit)
	}
	return nil
}

func idxTokenIDToString(idx string, tid common.TokenID) string {
	return idx + strconv.Itoa(int(tid))
}

// parse parses through reader
func (p *parser) parse() (*parsedSet, error) {
	ps := &parsedSet{}
	i := 0 // lines will start counting at line 1
	users := make(map[string]bool)
	for {
		i++
		instruction, err := p.parseLine(ps.typ)
		if err == errof {
			break
		}
		if err == setTypeLine {
			if ps.typ != "" {
				return ps, fmt.Errorf("Line %d: Instruction of 'Type: %s' when there is already a previous instruction 'Type: %s' defined", i, instruction.typ, ps.typ)
			}
			if instruction.typ == "PoolL2" {
				ps.typ = setTypePoolL2
			} else if instruction.typ == "Blockchain" {
				ps.typ = setTypeBlockchain
			} else {
				log.Fatalf("Line %d: Invalid set type: '%s'. Valid set types: 'Blockchain', 'PoolL2'", i, instruction.typ)
			}
			continue
		}
		if err == commentLine {
			continue
		}
		instruction.lineNum = i
		if err == newEventLine {
			if instruction.typ == typeAddToken && instruction.tokenID == common.TokenID(0) {
				return ps, fmt.Errorf("Line %d: AddToken can not register TokenID 0", i)
			}
			ps.instructions = append(ps.instructions, *instruction)
			continue
		}
		if err != nil {
			return ps, fmt.Errorf("Line %d: %s, err: %s", i, instruction.literal, err.Error())
		}
		if ps.typ == "" {
			return ps, fmt.Errorf("Line %d: Set type not defined", i)
		}
		ps.instructions = append(ps.instructions, *instruction)
		users[instruction.from] = true
		if instruction.typ == common.TxTypeTransfer || instruction.typ == common.TxTypeTransferToEthAddr || instruction.typ == common.TxTypeTransferToBJJ { // type: Transfer
			users[instruction.to] = true
		}
	}
	for u := range users {
		ps.users = append(ps.users, u)
	}
	sort.Strings(ps.users)
	return ps, nil
}
