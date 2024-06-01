package parser

import (
	"fmt"

	"github.com/marcfyk/go-jsonpath/internal/parser/grammar"
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
	return p.queryJSONPath()
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

// queryJSONPath will parse the given codepoints based on the jsonpath grammar rules.
func (p *Parser) queryJSONPath() error {
	if err := p.identRootNode(); err != nil {
		return err
	}
	return p.segments()
}

func (p *Parser) identRootNode() error {
	return p.expect(grammar.Dollar)
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
	if p.segmentChild() == nil {
		return nil
	} else if p.segmentDescendant() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) segmentChild() error {
	if err := p.bracketedSelection(); err == nil {
		return nil
	} else {
		if err := p.expect(grammar.Dot); err != nil {
			return err
		}
		if p.selectorWildcard() == nil {
			return nil
		} else if p.memberNameShorthand() == nil {
			return nil
		} else {
			return p.errorUnsupportedToken()
		}
	}
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
	if p.selectorName() == nil {
		return nil
	} else if p.selectorWildcard() == nil {
		return nil
	} else if p.selectorSlice() == nil {
		return nil
	} else if p.selectorIndex() == nil {
		return nil
	} else if p.selectorFilter() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) selectorName() error {
	return p.literalString()
}

func (p *Parser) literalString() error {
	if p.expect(grammar.QuoteDouble) == nil {
		for p.quotedDouble() == nil {
		}
		return p.expect(grammar.QuoteDouble)
	} else if p.expect(grammar.QuoteSingle) == nil {
		for p.quotedSingle() == nil {
		}
		return p.expect(grammar.QuoteSingle)
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) quotedDouble() error {
	if p.expectBy(isUnescaped) == nil {
		return nil
	} else if p.expect(grammar.QuoteSingle) == nil {
		return nil
	} else if p.expect(grammar.Esc) == nil {
		if p.expect(grammar.QuoteDouble) == nil {
			return nil
		} else if p.escapable() == nil {
			return nil
		} else {
			return p.errorUnsupportedToken()
		}
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) quotedSingle() error {
	if p.expectBy(isUnescaped) == nil {
		return nil
	} else if p.expect(grammar.QuoteDouble) == nil {
		return nil
	} else if p.expect(grammar.Esc) == nil {
		if p.expect(grammar.QuoteSingle) == nil {
			return nil
		} else if p.escapable() == nil {
			return nil
		} else {
			return p.errorUnsupportedToken()
		}
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) escapable() error {
	for _, r := range grammar.Escapable {
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
	} else {
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
}

func (p *Parser) nonSurrogate() error {

	if p.expect('D') == nil {
		if err := p.expectBy(isDigit0To7); err != nil {
			return err
		}
		for range 2 {
			if err := p.expectBy(isHexDig); err != nil {
				return err
			}
		}
		return nil
	} else {
		predicate := func(r rune) bool {
			return isDigit(r) ||
				r == 'A' ||
				r == 'B' ||
				r == 'C' ||
				r == 'E' ||
				r == 'F'
		}
		if err := p.expectBy(predicate); err != nil {
			return err
		}
		for range 3 {
			if err := p.expectBy(isHexDig); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *Parser) highSurrogate() error {
	if err := p.expect('D'); err != nil {
		return err
	}
	predicate := func(r rune) bool {
		return r == '8' ||
			r == '9' ||
			r == 'A' ||
			r == 'B'
	}
	if err := p.expectBy(predicate); err != nil {
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
	predicate := func(r rune) bool {
		return r == 'C' ||
			r == 'D' ||
			r == 'E' ||
			r == 'F'
	}
	if err := p.expectBy(predicate); err != nil {
		return err
	}
	for range 2 {
		if err := p.hexChar(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) selectorWildcard() error {
	return p.expect(grammar.Wildcard)
}

func (p *Parser) selectorSlice() error {
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
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
			_ = p.step()
		}
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
	} else {
		_ = p.expect(grammar.Minus)
		if err := p.expectBy(isDigit1); err != nil {
			return err
		}
		for p.expectBy(isDigit) == nil {
		}
		return nil
	}
}

func (p *Parser) selectorIndex() error {
	return p.int()
}

func (p *Parser) selectorFilter() error {
	if err := p.expect(grammar.Question); err != nil {
		return err
	}
	p.blankSpace()
	return p.logicalExpr()
}

func (p *Parser) logicalExpr() error {
	return p.logicalExprOr()
}

func (p *Parser) logicalExprOr() error {
	if err := p.logicalExprAnd(); err != nil {
		return err
	}
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.or(); err != nil {
			return err
		}
		p.blankSpace()
		if err := p.logicalExprAnd(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) logicalExprAnd() error {
	if err := p.basicExpr(); err != nil {
		return err
	}
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if err := p.and(); err != nil {
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
	} else if p.comparisonExpr() == nil {
		return nil
	} else if p.testExpr() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) parenExpr() error {
	_ = p.not()
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

func (p *Parser) not() error {
	return p.expect(grammar.Bang)
}

func (p *Parser) and() error {
	for _, r := range grammar.And {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) or() error {
	for _, r := range grammar.Or {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
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
	} else if p.querySingular() == nil {
		return nil
	} else if p.functionExpr() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) literal() error {
	if p.literalNumber() == nil {
		return nil
	} else if p.literalString() == nil {
		return nil
	} else if p.literalTrue() == nil {
		return nil
	} else if p.literalFalse() == nil {
		return nil
	} else if p.literalNull() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) literalNumber() error {
	if err := p.int(); err == nil {
	} else if err := p.negativeZero(); err == nil {
	} else {
		return p.errorUnsupportedToken()
	}
	_ = p.frac()
	_ = p.exp()
	return nil
}

func (p *Parser) negativeZero() error {
	if err := p.expect(grammar.Minus); err != nil {
		return err
	}
	return p.expect('0')
}

func (p *Parser) frac() error {
	if err := p.expect(grammar.Dot); err != nil {
		return err
	}
	if err := p.expectBy(isDigit); err != nil {
		return err
	}
	for p.expectBy(isDigit) == nil {
	}
	return nil
}

func (p *Parser) exp() error {
	if err := p.expect('e'); err != nil {
		return err
	}
	if p.expect(grammar.Minus) == nil {
	} else if p.expect(grammar.Plus) == nil {
	}
	if err := p.expectBy(isDigit); err != nil {
		return err
	}
	for p.expectBy(isDigit) == nil {
	}
	return nil
}

func (p *Parser) literalTrue() error {
	for _, r := range grammar.True {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) literalFalse() error {
	for _, r := range grammar.False {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) literalNull() error {
	for _, r := range grammar.Null {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) querySingular() error {
	if p.querySingularRel() == nil {
		return nil
	} else if p.querySingularAbs() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) querySingularRel() error {
	if err := p.identCurrNode(); err != nil {
		return err
	}
	return p.querySingularSegments()
}

func (p *Parser) identCurrNode() error {
	return p.expect(grammar.At)
}

func (p *Parser) querySingularSegments() error {
	for {
		if !p.matchBy(isBlankSpace) {
			break
		}
		p.blankSpace()
		if p.segmentName() == nil {
			continue
		} else if p.segmentIndex() == nil {
			continue
		} else {
			return p.errorUnsupportedToken()
		}
	}
	return nil
}

func (p *Parser) segmentName() error {
	if p.expect(grammar.BracketOpen) == nil {
		if err := p.selectorName(); err != nil {
			return err
		}
		return p.expect(grammar.BracketClose)
	} else if p.expect(grammar.Dot) == nil {
		return p.memberNameShorthand()
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) memberNameShorthand() error {
	if err := p.expectBy(isNameFirst); err != nil {
		return err
	}
	for p.expectBy(isNameChar) == nil {
	}
	return nil
}

func (p *Parser) segmentIndex() error {
	if err := p.expect(grammar.BracketOpen); err != nil {
		return err
	}
	if err := p.int(); err != nil {
		return err
	}
	return p.expect(grammar.BracketClose)
}

func (p *Parser) querySingularAbs() error {
	if err := p.identRootNode(); err != nil {
		return err
	}
	return p.querySingularSegments()
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
	} else if p.expect(grammar.Underscore) == nil {
		return nil
	} else if p.expectBy(isDigit) == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) functionArgument() error {
	if p.literal() == nil {
		return nil
	} else if p.selectorFilter() == nil {
		return nil
	} else if p.logicalExpr() == nil {
		return nil
	} else if p.functionExpr() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) queryFilter() error {
	if p.queryRel() == nil {
		return nil
	} else if p.queryJSONPath() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) queryRel() error {
	if err := p.identCurrNode(); err != nil {
		return err
	}
	return p.segments()
}

func (p *Parser) comparisonOp() error {
	if p.expect(grammar.Eq) == nil {
		return p.expect(grammar.Eq)
	} else if p.expect(grammar.Bang) == nil {
		return p.expect(grammar.Eq)
	} else if p.expect(grammar.Lt) == nil {
		_ = p.expect(grammar.Eq)
		return nil
	} else if p.expect(grammar.Gt) == nil {
		_ = p.expect(grammar.Eq)
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) testExpr() error {
	if p.expect(grammar.Bang) == nil {
		p.blankSpace()
	}
	if p.queryFilter() == nil {
		return nil
	} else if p.functionExpr() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
}

func (p *Parser) segmentDescendant() error {
	for _, r := range grammar.DescendantPrefix {
		if err := p.expect(r); err != nil {
			return err
		}
	}
	if p.bracketedSelection() == nil {
		return nil
	} else if p.selectorWildcard() == nil {
		return nil
	} else if p.memberNameShorthand() == nil {
		return nil
	} else {
		return p.errorUnsupportedToken()
	}
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
