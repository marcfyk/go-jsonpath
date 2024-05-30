package parser

import (
	"errors"
	"fmt"

	"github.com/marcfyk/go-jsonpath/internal/grammar"
)

func New(jsonpath string) Parser {
	if jsonpath == "" {
		return Parser{codepoints: []rune{}, index: 0, currCodepoint: nil}
	}
	codepoints := []rune(jsonpath)
	return Parser{
		codepoints:    codepoints,
		index:         0,
		currCodepoint: &codepoints[0],
	}
}

// ErrUnexpectedToken is the error type when the parser encounters a token
// it does not expect at its given state.
type ErrUnexpectedToken struct {
	// token is the actual unicode codepoint encountered by the parser.
	token *rune
	// index is the index of the unicode codepoint of token.
	index int
}

func (e ErrUnexpectedToken) Error() string {
	return fmt.Sprintf(
		"unexpected token:%v; found at index:%d",
		e.token, e.index)
}

// Parser is a recursive descent parser that scans jsonpath strings.
//
// If the Parser is given an empty string, currCodepoint = nil and index = 0.
type Parser struct {
	// codepoints is a slice of unicode codepoints from the given jsonpath string.
	codepoints []rune
	// currCodepoint is the current codepoint that the Parser is looking at.
	currCodepoint *rune
	// index is the zero-based index of the currCodepoint.
	index int
}

// Parse parses the codepoints in the Parser according to the jsonpath grammar rules.
func (p *Parser) Parse() error {
    return p.jsonpathQuery()
}

// errorUnsupportedToken returns an ErrUnsupportedToken error with the current state of
// the Parser's currCodepoint and index.
func (p *Parser) errorUnsupportedToken() ErrUnexpectedToken {
	return ErrUnexpectedToken{
		token: p.currCodepoint,
		index: p.index,
	}
}

// shift will move the currCodepoint and index to the next codepoint
// based on the sequence of codepoints.
func (p *Parser) shift() {
	p.index = min(p.index+1, len(p.codepoints))
	if p.index < len(p.codepoints) {
		p.currCodepoint = &p.codepoints[p.index]
	} else {
		p.currCodepoint = nil
	}
}

// matchBy will return if currCodepoint is true or false
// if it is not nil and based on the application of
// predicate, f, on currCodepoint.
func (p *Parser) matchBy(f func(rune) bool) bool {
	return p.currCodepoint != nil && f(*p.currCodepoint)
}

// match returns if the currCodepoint equals a given codepoint.
func (p *Parser) match(codepoint rune) bool {
	return p.matchBy(func(r rune) bool { return r == codepoint })
}

// expectBy attempts to match the currCodepoint by predicate, f.
// If the predicate evaluates to true, the parser shifts to the next codepoint.
// If the predicate evaluates to false, then ErrUnexpectedToken is returned.
func (p *Parser) expectBy(f func(rune) bool) error {
	if !p.matchBy(f) {
		return p.errorUnsupportedToken()
	}
	p.shift()
	return nil
}

// expect attempts to match the currCodepoint against a given codepoint.
// The behavior and side effects on whether the match is successful is the same as expectBy.
func (p *Parser) expect(codepoint rune) error {
	return p.expectBy(func(r rune) bool { return r == codepoint })
}

// jsonpathQuery will parse the given codepoints based on the jsonpath grammar rules.
func (p *Parser) jsonpathQuery() error {
	if err := p.rootIdent(); err != nil {
		return err
	}
	return p.segments()
}

func (p *Parser) rootIdent() error {
	return p.expect(grammar.Root)
}

func (p *Parser) blankSpace() {
	for p.expectBy(isBlankSpace) == nil {
	}
}

func (p *Parser) segments() error {
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.segment(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) segment() error {
	if p.childSegment() == nil {
		return nil
	}
	return p.descendantSegment()
}

func (p *Parser) childSegment() error {
	if err := p.bracketedSelection(); err == nil {
		return nil
	}
	if err := p.expect(grammar.Dot); err != nil {
		return err
	}
	if err := p.wildcardSelector(); err == nil {
		return nil
	}
	return p.memberNameShorthand()
}

func (p *Parser) bracketedSelection() error {
	if err := p.expect(grammar.BracketOpen); err != nil {
		return err
	}
	p.blankSpace()
	if err := p.selector(); err != nil {
		return err
	}
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.expect(grammar.Comma); err != nil {
			return err
		}
		p.blankSpace()
		if err := p.selector(); err != nil {
			return err
		}
	}
	p.blankSpace()
	return p.expect(grammar.BracketClose)
}

func (p *Parser) selector() error {
	if p.nameSelector() == nil {
		return nil
	}
	if p.wildcardSelector() == nil {
		return nil
	}
	if p.sliceSelector() == nil {
		return nil
	}
	if p.indexSelector() == nil {
		return nil
	}
	return p.filterSelector()
}

func (p *Parser) nameSelector() error {
	return p.stringLiteral()
}

func (p *Parser) stringLiteral() error {
	if p.expect(grammar.QuoteDouble) == nil {
		for p.doubleQuoted() == nil {
		}
		return p.expect(grammar.QuoteDouble)
	}
	if err := p.expect(grammar.QuoteSingle); err != nil {
		return nil
	}
	for p.singleQuoted() == nil {
	}
	return p.expect(grammar.QuoteSingle)
}

func (p *Parser) doubleQuoted() error {
	return p.quoted(grammar.QuoteDouble)
}

func (p *Parser) singleQuoted() error {
	return p.quoted(grammar.QuoteSingle)
}

func (p *Parser) quoted(quote rune) error {

	if p.expectBy(isUnescaped) == nil {
		return nil
	}
	unescapedQuote, err := getUnescapedQuote(quote)
	if err != nil {
		return err
	}
	if p.expect(unescapedQuote) == nil {
		return nil
	}
	if err := p.expect(grammar.Esc); err != nil {
		return err
	}
	if p.expect(quote) == nil {
		return nil
	}
	return p.escapable()
}

func (p *Parser) escapable() error {
	elements := [...]rune{grammar.BS, grammar.FF, grammar.LF, grammar.CR, grammar.HT, grammar.Slash, grammar.BackSlash}
	for _, r := range elements {
		if p.expect(r) == nil {
			return nil
		}
	}
	if err := p.expect(grammar.UnicodeEscape); err != nil {
		return err
	}
	return p.hexChar()
}

func (p *Parser) hexChar() error {
	if p.nonSurrogate() == nil {
		return nil
	}
	if err := p.highSurrogate(); err != nil {
		return err
	}
	if err := p.expect(grammar.BackSlash); err != nil {
		return err
	}
	if err := p.expect(grammar.UnicodeEscape); err != nil {
		return err
	}
	return p.lowSurrogate()
}

func (p *Parser) nonSurrogate() error {
	err := func() error {
		if p.expectBy(isDigit) == nil {
			return nil
		}
		for _, r := range [...]rune{'A', 'B', 'C', 'E'} {
			if p.expect(r) == nil {
				return nil
			}
		}
		if err := p.expect('F'); err != nil {
			return err
		}
		for range 3 {
			if err := p.expectBy(isHexDig); err != nil {
				return err
			}
		}
		return nil
	}()
	if err == nil {
		return nil
	}
	if err := p.expect('D'); err != nil {
		return err
	}
	if err := p.expectBy(isDigit0To7); err != nil {
		return err
	}
	for range 2 {
		if err := p.expectBy(isHexDig); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) highSurrogate() error {
	if err := p.expect('D'); err != nil {
		return err
	}
	err := func() error {
		for _, r := range [...]rune{'8', '9', 'A', 'B'} {
			if p.expect(r) == nil {
				return nil
			}
		}
		return p.errorUnsupportedToken()
	}()
	if err != nil {
		return err
	}
	for range 2 {
		if err := p.expectBy(isHexDig); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) lowSurrogate() error {
	if err := p.expect('D'); err != nil {
		return err
	}
	err := func() error {
		for _, r := range [...]rune{'C', 'D', 'E', 'F'} {
			if p.expect(r) == nil {
				return nil
			}
		}
		return p.errorUnsupportedToken()
	}()
	if err != nil {
		return err
	}
	for range 2 {
		if err := p.expectBy(isHexDig); err != nil {
			return err
		}
	}
	return nil

}

func (p *Parser) wildcardSelector() error {
	return p.expect(grammar.Wildcard)
}

func (p *Parser) sliceSelector() error {
	if p.start() == nil {
		p.blankSpace()
	}
	if err := p.expect(grammar.Colon); err != nil {
		return err
	}
	p.blankSpace()
	if p.end() == nil {
		p.blankSpace()
	}
	if p.expect(grammar.Colon) == nil {
		p.blankSpace()
		_ = p.step()
	}
	return nil
}

func (p *Parser) start() error {
	return p.int()
}

func (p *Parser) end() error {
	return p.int()
}

func (p *Parser) step() error {
	return p.int()
}

func (p *Parser) int() error {
	if p.expect('0') == nil {
		return nil
	}
	_ = p.expect(grammar.Minus)
	if err := p.expectBy(isDigit1); err != nil {
		return err
	}
	for p.expectBy(isDigit) == nil {
	}
	return nil
}

func (p *Parser) indexSelector() error {
	return p.int()
}

func (p *Parser) filterSelector() error {
	if err := p.expect(grammar.Question); err != nil {
		return err
	}
	p.blankSpace()
	return p.logicalExpr()
}

func (p *Parser) logicalExpr() error {
	return p.logicalOrExpr()
}

func (p *Parser) logicalOrExpr() error {
	if err := p.logicalAndExpr(); err != nil {
		return err
	}
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.expect(grammar.Pipe); err != nil {
			return err
		}
		if err := p.expect(grammar.Pipe); err != nil {
			return err
		}
		p.blankSpace()
		if err := p.logicalAndExpr(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) logicalAndExpr() error {
	if err := p.basicExpr(); err != nil {
		return err
	}
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.expect(grammar.Ampersand); err != nil {
			return err
		}
		if err := p.expect(grammar.Ampersand); err != nil {
			return err
		}
		p.blankSpace()
		if err := p.basicExpr(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) basicExpr() error {
	if p.parenExpr() == nil {
		return nil
	}
	if p.comparisonExpr() == nil {
		return nil
	}
	return p.testExpr()
}

func (p *Parser) parenExpr() error {
	_ = p.logicalNotOp()
	p.blankSpace()
	if err := p.expect(grammar.ParenthesisOpen); err != nil {
		return err
	}
	p.blankSpace()
	if err := p.logicalExpr(); err != nil {
		return err
	}
	p.blankSpace()
	return p.expect(grammar.ParenthesisClose)
}

func (p *Parser) logicalNotOp() error {
	return p.expect(grammar.Bang)
}

func (p *Parser) comparisonExpr() error {
	if err := p.comparable(); err != nil {
		return err
	}
	p.blankSpace()
	if err := p.comparisonOp(); err != nil {
		return err
	}
	p.blankSpace()
	return p.comparable()
}

func (p *Parser) comparable() error {
	if p.literal() == nil {
		return nil
	}
	if p.singularQuery() == nil {
		return nil
	}
	return p.functionExpr()
}

func (p *Parser) literal() error {
	if p.number() == nil {
		return nil
	}
	if p.stringLiteral() == nil {
		return nil
	}
	if p.true() == nil {
		return nil
	}
	if p.false() == nil {
		return nil
	}
	return p.null()
}

func (p *Parser) number() error {
	if err := p.int(); err != nil {
		if err := p.expect(grammar.Minus); err != nil {
			return err
		}
		if err := p.expect('0'); err != nil {
			return err
		}
	}
	if err := p.frac(); err != nil {
		_ = p.exp()
	}
	return nil
}

func (p *Parser) frac() error {
	if err := p.expect(grammar.Dot); err != nil {
		return err
	}
	if err := p.expectBy(isDigit1); err != nil {
		return err
	}
	for p.expectBy(isDigit1) == nil {
	}
	return nil
}

func (p *Parser) exp() error {
	if err := p.expect('e'); err != nil {
		return err
	}
	if err := p.expect(grammar.Minus); err != nil {
		_ = p.expect(grammar.Plus)
	}
	if err := p.expectBy(isDigit); err != nil {
		return err
	}
	for p.expectBy(isDigit) == nil {
	}
	return nil
}

func (p *Parser) true() error {
	for _, r := range [...]rune{'t', 'r', 'u', 'e'} {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) false() error {
	for _, r := range [...]rune{'f', 'a', 'l', 's', 'e'} {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) null() error {
	for _, r := range [...]rune{'n', 'u', 'l', 'l'} {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) singularQuery() error {
	if p.relSingularQuery() == nil {
		return nil
	}
	return p.absSingularQuery()
}

func (p *Parser) relSingularQuery() error {
	if err := p.currNodeIdent(); err != nil {
		return err
	}
	return p.singularQuerySegments()
}

func (p *Parser) currNodeIdent() error {
	return p.expect(grammar.At)
}

func (p *Parser) singularQuerySegments() error {
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.nameSegment(); err == nil {
			continue
		}
		if err := p.indexSegment(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) nameSegment() error {
	if err := p.expect(grammar.BracketOpen); err == nil {
		if err := p.nameSelector(); err != nil {
			return err
		}
		if err := p.expect(grammar.BracketClose); err != nil {
			return err
		}
		return nil
	}
	if err := p.expect(grammar.Dot); err != nil {
		return err
	}
	return p.memberNameShorthand()
}

func (p *Parser) memberNameShorthand() error {
	if err := p.expectBy(isNameFirst); err != nil {
		return err
	}
	for p.expectBy(isNameChar) == nil {
	}
	return nil
}

func (p *Parser) indexSegment() error {
	if err := p.expect(grammar.BracketOpen); err != nil {
		return err
	}
	if err := p.int(); err != nil {
		return err
	}
	return p.expect(grammar.BracketClose)
}

func (p *Parser) absSingularQuery() error {
	if err := p.rootIdent(); err != nil {
		return err
	}
	return p.singularQuerySegments()
}

func (p *Parser) functionExpr() error {
	if err := p.functionName(); err != nil {
		return err
	}
	if err := p.expect(grammar.ParenthesisOpen); err != nil {
		return err
	}
	p.blankSpace()
	if p.functionArgument() == nil {
		for {
			if !p.matchBy(isBlankSpace) {
				break
			}
			p.blankSpace()
			if err := p.expect(grammar.Comma); err != nil {
				return err
			}
			p.blankSpace()
			if err := p.functionArgument(); err != nil {
				return err
			}
		}
	}
	p.blankSpace()
	return p.expect(grammar.BracketClose)
}

func (p *Parser) functionName() error {
	if err := p.functionNameFirst(); err != nil {
		return err
	}
	for p.functionNameChar() == nil {
	}
	return nil
}

func (p *Parser) functionNameFirst() error {
	return p.expectBy(isAlphaLowercase)
}

func (p *Parser) functionNameChar() error {
	if p.functionNameFirst() == nil {
		return nil
	}
	if p.expect(grammar.Underscore) == nil {
		return nil
	}
	return p.expectBy(isDigit)
}

func (p *Parser) functionArgument() error {
	if p.literal() == nil {
		return nil
	}
	if p.filterSelector() == nil {
		return nil
	}
	if p.logicalExpr() == nil {
		return nil
	}
	return p.functionExpr()
}

func (p *Parser) filterQuery() error {
	if p.relQuery() == nil {
		return nil
	}
	return p.jsonpathQuery()
}

func (p *Parser) relQuery() error {
	if err := p.currNodeIdent(); err != nil {
		return err
	}
	return p.segments()
}

func (p *Parser) comparisonOp() error {
	if p.expect(grammar.Eq) == nil {
		return p.expect(grammar.Eq)
	}
	if p.expect(grammar.Bang) == nil {
		return p.expect(grammar.Eq)
	}
	if p.expect(grammar.Lt) == nil {
		_ = p.expect(grammar.Eq)
	}
	if err := p.expect(grammar.Gt); err != nil {
		return err
	}
	_ = p.expect(grammar.Eq)
	return nil
}

func (p *Parser) testExpr() error {
	if p.expect(grammar.Bang) == nil {
		p.blankSpace()
	}
	if p.filterQuery() == nil {
		return nil
	}
	return p.functionExpr()
}

func (p *Parser) descendantSegment() error {
	if err := p.expect(grammar.Dot); err != nil {
		return err
	}
	if err := p.expect(grammar.Dot); err != nil {
		return err
	}
	if p.bracketedSelection() == nil {
		return nil
	}
	if p.wildcardSelector() == nil {
		return nil
	}
	return p.memberNameShorthand()
}

func getUnescapedQuote(quote rune) (rune, error) {
	switch quote {
	case grammar.QuoteDouble:
		return grammar.QuoteSingle, nil
	case grammar.QuoteSingle:
		return grammar.QuoteDouble, nil
	default:
		return 0, errors.New("invalid quote")
	}
}

func isQuote(r rune) bool {
	return r == grammar.QuoteDouble || r == grammar.QuoteSingle
}

func isBlankSpace(r rune) bool {
	return r == grammar.Space ||
		r == grammar.HorizonalTab ||
		r == grammar.Newline ||
		r == grammar.CarriageReturn
}

func isUnescaped(r rune) bool {
	return ('\x20' <= r && r <= '\x21') ||
		('\x23' <= r && r <= '\x26') ||
		('\x28' <= r && r <= '\x5B') ||
		('\x5D' <= r && r <= '\uD7FF') ||
		(0xE000 <= r && r <= 0x10FFFF)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isDigit1(r rune) bool {
	return '1' <= r && r <= '9'
}

func isDigit0To7(r rune) bool {
	return '0' <= r && r <= '7'
}

func isAlphaUppercase(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func isAlphaLowercase(r rune) bool {
	return 'a' <= r && r <= 'z'
}

func isAlpha(r rune) bool {
	return isAlphaUppercase(r) || isAlphaLowercase(r)
}

func isNameFirst(r rune) bool {
	return isAlpha(r) ||
		r == grammar.Underscore ||
		('\x80' <= r && r <= '\uD7FF') ||
		('\uE000' <= r && r <= 0x10FFFF)
}

func isNameChar(r rune) bool {
	return isNameFirst(r) || isDigit(r)
}

func isHexDig(r rune) bool {
	return isDigit(r) || ('A' <= r && r <= 'F')
}
