// Package parser implements the parsing component of jsonpath expressions.
//
// The parser component should take in jsonpath expressions or tokens output from a lexer,
// and output an IR of the jsonpath expression.
package parser

import (
	"errors"
	"fmt"
	"math"
	"regexp"

	"github.com/marcfyk/go-jsonpath/internal/ast"
	"github.com/marcfyk/go-jsonpath/internal/parser/grammar"
)

func New(jsonpath string) Parser {
	if jsonpath == "" {
		return Parser{Codepoints: []rune{}, Index: 0}
	}
	codepoints := []rune(jsonpath)
	return Parser{
		Codepoints: codepoints,
		Index:      0,
	}
}

// ErrUnexpectedCodepoint is the error type when the parser encounters a codepoint
// it does not expect at its given state.
type ErrUnexpectedCodepoint struct {
	// codepoint is the actual unicode codepoint encountered by the parser.
	codepoint *rune
	// index is the index of the unicode codepoint.
	index int
}

func (e ErrUnexpectedCodepoint) Error() string {
	return fmt.Sprintf(
		"unexpected codepoint:%v; found at index:%d",
		e.codepoint, e.index)
}

// Parser is a recursive descent parser that scans jsonpath strings.
type Parser struct {
	// Codepoints is a slice of unicode Codepoints from the given jsonpath string.
	Codepoints []rune
	// Index is the zero-based Index of the current codepoint.
	Index int
}

// IsDone returns if the parser has consumed all codepoints in its buffer.
func (p *Parser) IsDone() bool {
	return p.Index == len(p.Codepoints)
}

// Parse parses the codepoints in the Parser according to the jsonpath grammar rules.
func (p *Parser) Parse() (ast.Expr, error) {
	return p.queryJSONPath()
}

// errorUnsupportedCodepoint returns an ErrUnexpectedCodepoint error with the current state of
// the Parser's index.
func (p *Parser) errorUnsupportedCodepoint() ErrUnexpectedCodepoint {
	var c *rune
	if 0 <= p.Index && p.Index < len(p.Codepoints) {
		c = &p.Codepoints[p.Index]
	}
	return ErrUnexpectedCodepoint{
		codepoint: c,
		index:     p.Index,
	}
}

// shift will move the index to the next codepoint
// based on the sequence of codepoints.
func (p *Parser) shift() {
	p.Index = min(p.Index+1, len(p.Codepoints))
}

// matchBy will return if current codepoint is true or false
// based on the application of predicate, f, on the current codepoint.
// If the index is out of range, it is always false.
func (p *Parser) matchBy(f func(rune) bool) bool {

	return 0 <= p.Index && p.Index < len(p.Codepoints) && f(p.Codepoints[p.Index])
}

// match returns if the current codepoint equals a given codepoint.
func (p *Parser) match(codepoint rune) bool {
	return p.matchBy(func(r rune) bool { return r == codepoint })
}

// expectBy attempts to match the current codepoint by predicate, f.
// If the predicate evaluates to true, the parser shifts to the next codepoint.
// If the predicate evaluates to false, then ErrUnexpectedToken is returned.
func (p *Parser) expectBy(f func(rune) bool) error {
	if !p.matchBy(f) {
		return p.errorUnsupportedCodepoint()
	}
	p.shift()
	return nil
}

// expect attempts to match the current codepoint against a given codepoint.
// The behavior and side effects on whether the match is successful is the same as expectBy.
func (p *Parser) expect(codepoint rune) error {
	return p.expectBy(func(r rune) bool { return r == codepoint })
}

// queryJSONPath will parse the given codepoints based on the jsonpath grammar rules.
func (p *Parser) queryJSONPath() (ast.Expr, error) {
	if err := p.identRootNode(); err != nil {
		return nil, err
	}
	segments, err := p.segments()
	if err != nil {
		return nil, err
	}
	return ast.QueryJSONPath{
		Segments: segments,
	}, nil
}

func (p *Parser) identRootNode() error {
	return p.expect(grammar.Dollar)
}

func (p *Parser) blankSpace() {
	for p.expectBy(isBlankSpace) == nil {
	}
}

func (p *Parser) segments() ([]ast.Expr, error) {
	segments := make([]ast.Expr, 0)
	for {
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
		}
		initial := p.Index
		s, err := p.segment()
		if err != nil {
			p.Index = initial
			break
		}
		segments = append(segments, s)
	}
	return segments, nil
}

func (p *Parser) segment() (ast.Expr, error) {
	initial := p.Index
	if s, err := p.segmentChild(); err == nil {
		return s, nil
	}
	p.Index = initial
	return p.segmentDescendant()
}

func (p *Parser) segmentChild() (ast.Expr, error) {
	if b, err := p.bracketedSelection(); err == nil {
		return ast.SegmentChild{
			Selectors: b,
		}, nil
	} else {
		if err := p.expect(grammar.Dot); err != nil {
			return nil, err
		}
		if w, err := p.selectorWildcard(); err == nil {
			return w, nil
		} else if n, err := p.memberNameShorthand(); err == nil {
			return ast.SegmentChild{
				Selectors: []ast.Expr{ast.SelectorName{Name: n}},
			}, nil
		} else {
			return nil, p.errorUnsupportedCodepoint()
		}
	}
}

func (p *Parser) bracketedSelection() ([]ast.Expr, error) {
	selectors := make([]ast.Expr, 0)
	if err := p.expect(grammar.BracketOpen); err != nil {
		return nil, err
	}
	p.blankSpace()
	s, err := p.selector()
	if err != nil {
		return nil, err
	}
	selectors = append(selectors, s)
	for {
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
		}
		if err := p.expect(grammar.Comma); err != nil {
			break
		}
		p.blankSpace()
		s, err := p.selector()
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, s)
	}
	p.blankSpace()
	if err := p.expect(grammar.BracketClose); err != nil {
		return nil, err
	}
	return selectors, nil
}

func (p *Parser) selector() (ast.Expr, error) {
	switch {
	case p.matchBy(isQuote):
		return p.selectorName()
	case p.match(grammar.Asterisk):
		return p.selectorWildcard()
	case p.matchBy(func(r rune) bool { return isDigit(r) || r == '-' || r == ':' }):
		initial := p.Index
		if s, err := p.selectorSlice(); err == nil {
			return s, nil
		}
		p.Index = initial
		return p.selectorIndex()
	default:
		return p.selectorFilter()
	}
}

func (p *Parser) selectorName() (ast.SelectorName, error) {
	s, err := p.literalString()
	if err != nil {
		return ast.SelectorName{}, err
	}
	return ast.SelectorName{Name: s}, nil
}

func (p *Parser) literalString() (string, error) {
	if p.expect(grammar.QuoteDouble) == nil {
		start := p.Index
		for p.quotedDouble() == nil {
		}
		if err := p.expect(grammar.QuoteDouble); err != nil {
			return "", err
		}
		end := p.Index - 1
		return string(p.Codepoints[start:end]), nil
	} else if p.expect(grammar.QuoteSingle) == nil {
		start := p.Index
		for p.quotedSingle() == nil {
		}
		if err := p.expect(grammar.QuoteSingle); err != nil {
			return "", err
		}
		end := p.Index - 1
		return string(p.Codepoints[start:end]), nil
	} else {
		return "", p.errorUnsupportedCodepoint()
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
			return p.errorUnsupportedCodepoint()
		}
	} else {
		return p.errorUnsupportedCodepoint()
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
			return p.errorUnsupportedCodepoint()
		}
	} else {
		return p.errorUnsupportedCodepoint()
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

func (p *Parser) selectorWildcard() (ast.Expr, error) {
	if err := p.expect(grammar.Asterisk); err != nil {
		return nil, err
	}
	return ast.SelectorWildcard{}, nil
}

func (p *Parser) selectorSlice() (ast.Expr, error) {
	var start *int
	if s, err := p.start(); err == nil {
		start = &s
		p.blankSpace()
	}
	if err := p.expect(grammar.Colon); err != nil {
		return nil, err
	}
	p.blankSpace()
	var end *int
	if e, err := p.end(); err == nil {
		end = &e
		p.blankSpace()
	}
	step := 1
	if p.expect(grammar.Colon) == nil {
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
		}
		if s, err := p.step(); err == nil {
			step = s
		}
	}
	return ast.SelectorSlice{
		Start: start,
		End:   end,
		Step:  step,
	}, nil
}

func (p *Parser) start() (int, error) {
	return p.int()
}

func (p *Parser) end() (int, error) {
	return p.int()
}

func (p *Parser) step() (int, error) {
	return p.int()
}

func (p *Parser) int() (int, error) {
	if p.expect('0') == nil {
		return 0, nil
	} else {
		isNegative := false
		if p.expect(grammar.Minus) == nil {
			isNegative = true
		}
		if err := p.expectBy(isDigit1); err != nil {
			return 0, err
		}
		n := int(p.Codepoints[p.Index-1] - '0')
		for p.expectBy(isDigit) == nil {
			n = n*10 + int(p.Codepoints[p.Index-1]-'0')
		}
		if isNegative {
			n = -n
		}
		return n, nil
	}
}

func (p *Parser) selectorIndex() (ast.Expr, error) {
	n, err := p.int()
	if err != nil {
		return nil, err
	}
	return ast.SelectorIndex{Index: n}, nil
}

func (p *Parser) selectorFilter() (ast.Expr, error) {
	if err := p.expect(grammar.Question); err != nil {
		return nil, err
	}
	p.blankSpace()
	return p.logicalExpr()
}

func (p *Parser) logicalExpr() (ast.Expr, error) {
	return p.logicalExprOr()
}

func (p *Parser) logicalExprOr() (ast.Expr, error) {
	exprs := make([]ast.Expr, 0)
	e, err := p.logicalExprAnd()
	if err != nil {
		return nil, err
	}
	exprs = append(exprs, e)
	for {
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
		}
		initial := p.Index
		if err := p.or(); err != nil {
			p.Index = initial
			break
		}
		p.blankSpace()
		e, err := p.logicalExprAnd()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, e)
	}
	return ast.ExprLogicalOr{
		Exprs: exprs,
	}, nil
}

func (p *Parser) logicalExprAnd() (ast.Expr, error) {
	exprs := make([]ast.Expr, 0)
	e, err := p.basicExpr()
	if err != nil {
		return nil, err
	}
	exprs = append(exprs, e)
	for {
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
		}
		initial := p.Index
		if err := p.and(); err != nil {
			p.Index = initial
			break
		}
		p.blankSpace()
		e, err := p.basicExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, e)
	}
	return ast.ExprLogicalAnd{
		Exprs: exprs,
	}, nil
}

func (p *Parser) basicExpr() (ast.Expr, error) {
	initial := p.Index
	e, err := p.parenExpr()
	if err == nil {
		return e, nil
	}
	p.Index = initial
	e, err = p.comparisonExpr()
	if err == nil {
		return e, nil
	}
	p.Index = initial
	return p.testExpr()
}

func (p *Parser) parenExpr() (ast.Expr, error) {
	isNegated := false
	if p.not() == nil {
		isNegated = true
	}
	p.blankSpace()
	if err := p.expect(grammar.ParenthesisOpen); err != nil {
		return nil, err
	}
	p.blankSpace()
	e, err := p.logicalExpr()
	if err != nil {
		return nil, err
	}
	p.blankSpace()
	if err := p.expect(grammar.ParenthesisClose); err != nil {
		return nil, err
	}
	var expr ast.Expr = ast.ExprParen{Expr: e}
	if isNegated {
		expr = ast.ExprLogicalNot{Expr: expr}
	}
	return expr, nil
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

func (p *Parser) comparisonExpr() (ast.Expr, error) {
	left, err := p.comparable()
	if err != nil {
		return nil, err
	}
	p.blankSpace()
	f, err := p.comparisonOp()
	if err != nil {
		return nil, err
	}
	p.blankSpace()
	right, err := p.comparable()
	if err != nil {
		return nil, err
	}
	return ast.ExprComparison{
		Left:  left,
		Right: right,
		F:     f,
	}, nil
}

func (p *Parser) comparable() (ast.ExprSingle, error) {
	if v, err := p.literal(); err == nil {
		return ast.Literal{Value: v}, nil
	} else if e, err := p.querySingular(); err == nil {
		return e, nil
	} else if f, err := p.functionExpr(); err == nil {
		f, ok := f.(ast.ExprSingle)
		if !ok {
			return nil, errors.New("function does not implemented ast.ExprSingle")
		}
		return f, nil
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) literal() (ast.Value, error) {
	if n, err := p.literalNumber(); err == nil {
		return n, nil
	} else if s, err := p.literalString(); err == nil {
		return s, nil
	} else if p.literalTrue() == nil {
		return true, nil
	} else if p.literalFalse() == nil {
		return false, nil
	} else if p.literalNull() == nil {
		return nil, nil
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) literalNumber() (float64, error) {
	f := float64(0)
	if n, err := p.int(); err == nil {
		f = float64(n)
	} else if err := p.negativeZero(); err == nil {
		f = -0
	} else {
		return 0, p.errorUnsupportedCodepoint()
	}
	if frac, err := p.frac(); err == nil {
		f = f + (float64(frac) / 10)
	}
	if exp, err := p.exp(); err != nil {
		f = f * math.Pow10(exp)
	}
	return f, nil
}

func (p *Parser) negativeZero() error {
	if err := p.expect(grammar.Minus); err != nil {
		return err
	}
	return p.expect('0')
}

func (p *Parser) frac() (int, error) {
	if err := p.expect(grammar.Dot); err != nil {
		return 0, err
	}
	if err := p.expectBy(isDigit); err != nil {
		return 0, err
	}
	n := int(p.Codepoints[p.Index-1] - '0')
	for p.expectBy(isDigit) == nil {
		n = n*10 + int(p.Codepoints[p.Index-1]-'0')
	}
	return n, nil
}

func (p *Parser) exp() (int, error) {
	if err := p.expect('e'); err != nil {
		return 0, err
	}
	isNegative := false
	if p.expect(grammar.Minus) == nil {
		isNegative = true
	} else if p.expect(grammar.Plus) == nil {
	}
	if err := p.expectBy(isDigit); err != nil {
		return 0, err
	}
	n := 0
	for p.expectBy(isDigit) == nil {
		n = n*10 + int(p.Codepoints[p.Index]-'0')
	}
	if isNegative {
		n = -n
	}
	return n, nil
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

func (p *Parser) querySingular() (ast.ExprSingle, error) {
	if e, err := p.querySingularRel(); err == nil {
		return e, err
	} else if e, err := p.querySingularAbs(); err == nil {
		return e, nil
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) querySingularRel() (ast.ExprSingle, error) {
	if err := p.identCurrNode(); err != nil {
		return nil, err
	}
	s, err := p.querySingularSegments()
	if err != nil {
		return nil, err
	}
	return ast.QuerySingularRel{
		Segments: s,
	}, nil
}

func (p *Parser) identCurrNode() error {
	return p.expect(grammar.At)
}

func (p *Parser) querySingularSegments() ([]ast.ExprSingle, error) {
	segments := make([]ast.ExprSingle, 0)
	for {
		if p.matchBy(isBlankSpace) {
			p.blankSpace()
		}
		initial := p.Index
		if e, err := p.segmentName(); err == nil {
			segments = append(segments, e)
			continue
		}
		p.Index = initial
		if e, err := p.segmentIndex(); err == nil {
			segments = append(segments, e)
			continue
		}
		p.Index = initial
		break
	}
	return segments, nil
}

func (p *Parser) segmentName() (ast.ExprSingle, error) {
	if p.expect(grammar.BracketOpen) == nil {
		s, err := p.selectorName()
		if err != nil {
			return nil, err
		}
		if err := p.expect(grammar.BracketClose); err != nil {
			return nil, err
		}
		return ast.SegmentName{
			Name: s.Name,
		}, nil
	} else if p.expect(grammar.Dot) == nil {
		name, err := p.memberNameShorthand()
		if err != nil {
			return nil, err
		}
		return ast.SegmentName{
			Name: name,
		}, nil
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) memberNameShorthand() (string, error) {
	start := p.Index
	if err := p.expectBy(isNameFirst); err != nil {
		return "", err
	}
	for p.expectBy(isNameChar) == nil {
	}
	end := p.Index
	return string(p.Codepoints[start:end]), nil
}

func (p *Parser) segmentIndex() (ast.ExprSingle, error) {
	if err := p.expect(grammar.BracketOpen); err != nil {
		return nil, err
	}
	n, err := p.int()
	if err != nil {
		return nil, err
	}
	if err := p.expect(grammar.BracketClose); err != nil {
		return nil, err
	}
	return ast.SegmentIndex{
		Index: n,
	}, nil
}

func (p *Parser) querySingularAbs() (ast.ExprSingle, error) {
	if err := p.identRootNode(); err != nil {
		return nil, err
	}
	s, err := p.querySingularSegments()
	if err != nil {
		return nil, err
	}
	return ast.QuerySingularAbs{
		Segments: s,
	}, nil
}

func (p *Parser) functionExpr() (ast.Expr, error) {
	name, err := p.functionName()
	if err != nil {
		return nil, err
	}
	if !isSupportedFunc(name) {
		return nil, ErrUnsupportedFunction{Name: name}
	}
	if err := p.expect(grammar.ParenthesisOpen); err != nil {
		return nil, err
	}
	p.blankSpace()
	args := make([]ast.Expr, 0)
	if a, err := p.functionArgument(); err == nil {
		args = append(args, a)
		for {
			if p.matchBy(isBlankSpace) {
				p.blankSpace()
			}
			if err := p.expect(grammar.Comma); err != nil {
				break
			}
			p.blankSpace()
			a, err := p.functionArgument()
			if err != nil {
				return nil, err
			}
			args = append(args, a)
		}
	}
	p.blankSpace()
	if err := p.expect(grammar.ParenthesisClose); err != nil {
		return nil, err
	}
	return generateFunc(name, args)
}

func (p *Parser) functionName() (string, error) {
	start := p.Index
	if err := p.functionNameFirst(); err != nil {
		return "", err
	}
	for p.functionNameChar() == nil {
	}
	end := p.Index
	return string(p.Codepoints[start:end]), nil
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
		return p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) functionArgument() (ast.Expr, error) {
	if l, err := p.literal(); err == nil {
		return ast.Literal{Value: l}, nil
	}
	if e, err := p.selectorFilter(); err == nil {
		return e, nil
	}
	if e, err := p.logicalExpr(); err == nil {
		return e, nil
	}
	e, err := p.functionExpr()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (p *Parser) queryFilter() (ast.Expr, error) {
	if e, err := p.queryRel(); err == nil {
		return e, nil
	} else if e, err := p.queryJSONPath(); err == nil {
		return e, nil
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) queryRel() (ast.Expr, error) {
	if err := p.identCurrNode(); err != nil {
		return nil, err
	}
	s, err := p.segments()
	if err != nil {
		return nil, err
	}
	return ast.QueryRel{
		Segments: s,
	}, nil
}

func (p *Parser) comparisonOp() (func(ast.Value, ast.Value) bool, error) {
	if p.expect(grammar.Eq) == nil {
		if err := p.expect(grammar.Eq); err != nil {
			return nil, err
		}
		return ast.EQ, nil
	} else if p.expect(grammar.Bang) == nil {
		if err := p.expect(grammar.Eq); err != nil {
			return nil, err
		}
		return ast.NE, nil
	} else if p.expect(grammar.Lt) == nil {
		if err := p.expect(grammar.Eq); err == nil {
			return ast.LTE, nil
		} else {
			return ast.LT, nil
		}
	} else if p.expect(grammar.Gt) == nil {
		if err := p.expect(grammar.Eq); err == nil {
			return ast.GTE, nil
		} else {
			return ast.GT, nil
		}
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func (p *Parser) testExpr() (ast.Expr, error) {
	isNegated := false
	if p.expect(grammar.Bang) == nil {
		isNegated = true
		p.blankSpace()
	}
	if e, err := p.queryFilter(); err == nil {
		if isNegated {
			e = ast.ExprLogicalNot{Expr: e}
		}
		return e, nil
	}
	f, err := p.functionExpr()
	if err != nil {
		return nil, err
	}
	if isNegated {
		f = ast.ExprLogicalNot{Expr: f}
	}
	return f, nil
}

func (p *Parser) segmentDescendant() (ast.Expr, error) {
	for _, r := range grammar.DescendantPrefix {
		if err := p.expect(r); err != nil {
			return nil, err
		}
	}
	if s, err := p.bracketedSelection(); err == nil {
		return ast.SegmentDescendant{
			Selectors: s,
		}, nil
	} else if s, err := p.selectorWildcard(); err == nil {
		return ast.SegmentDescendant{
			Selectors: []ast.Expr{s},
		}, nil
	} else if n, err := p.memberNameShorthand(); err == nil {
		return ast.SegmentDescendant{
			Selectors: []ast.Expr{ast.SelectorName{
				Name: n,
			}},
		}, nil
	} else {
		return nil, p.errorUnsupportedCodepoint()
	}
}

func isBlankSpace(r rune) bool {
	switch r {
	case grammar.Space,
		grammar.HorizonalTab,
		grammar.Newline,
		grammar.CarriageReturn:
		return true
	default:
		return false
	}
}

func isQuote(r rune) bool {
	switch r {
	case grammar.QuoteDouble, grammar.QuoteSingle:
		return true
	default:
		return false
	}
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

func isSupportedFunc(name string) bool {
	switch name {
	case grammar.FuncLength, grammar.FuncCount, grammar.FuncMatch, grammar.FuncSearch, grammar.FuncValue:
		return true
	default:
		return false
	}
}

// ErrWrongArgTypeFunction is the error type when the argument is a different
// type than expected.
type ErrWrongArgTypeFunction struct {
	// Name is the name of the function.
	Name string
	// Index is the zeroth-based index of the argument supplied to the function.
	Index int
	// ExpectedType is the type required.
	ExpectedType interface{}
	// ActualType is the actual type supplied in the query.
	ActualType interface{}
}

func (e ErrWrongArgTypeFunction) Error() string {
	return fmt.Sprintf(
		"invalid type at %s($%d); expected:%T; actual:%T",
		e.Name, e.Index, e.ExpectedType, e.ActualType)
}

// ErrWrongArgsCountFunction is the error type when the argument found is
// not correct.
type ErrWrongArgsCountFunction struct {
	// Name is the name of the function.
	Name string
	// Expected is the correct number of arguments for this function.
	Expected int
	// Actual is the actual number of arguments supplied to the function in the query.
	Actual int
}

func (e ErrWrongArgsCountFunction) Error() string {
	return fmt.Sprintf(
		"wrong arguments supplied to function:%s; expected:%d; actual:%d",
		e.Name, e.Expected, e.Actual)
}

// ErrUnsupportedFunction is the error type when the function
// is not specified in the RFC.
type ErrUnsupportedFunction struct {
	// Name is the function's name in the query.
	Name string
}

func (e ErrUnsupportedFunction) Error() string {
	return fmt.Sprintf("unsupported function:%v", e.Name)
}

func generateFunc(name string, args []ast.Expr) (ast.Expr, error) {
	switch name {
	case grammar.FuncLength:
		if len(args) != 1 {
			return nil, ErrWrongArgsCountFunction{Name: name, Expected: 1, Actual: len(args)}
		}
		return ast.FuncLength{Arg: args[0]}, nil
	case grammar.FuncCount:
		if len(args) != 1 {
			return nil, ErrWrongArgsCountFunction{Name: name, Expected: 1, Actual: len(args)}
		}
		return ast.FuncCount{Arg: args[0]}, nil
	case grammar.FuncMatch:
		if len(args) != 2 {
			return nil, ErrWrongArgsCountFunction{Name: name, Expected: 1, Actual: len(args)}
		}
		literal, ok := args[1].(ast.Literal)
		if !ok {
			return nil, ErrWrongArgTypeFunction{Name: name, Index: 1, ExpectedType: ast.Literal{Value: ""}, ActualType: args[1]}
		}
		s, ok := literal.Value.(string)
		if !ok {
			return nil, ErrWrongArgTypeFunction{Name: name, Index: 1, ExpectedType: "", ActualType: args[1]}
		}
		rg, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		return ast.FuncMatch{Regex: rg}, nil
	case grammar.FuncSearch:
		if len(args) != 2 {
			return nil, ErrWrongArgsCountFunction{Name: name, Expected: 1, Actual: len(args)}
		}
		literal, ok := args[1].(ast.Literal)
		if !ok {
			return nil, ErrWrongArgTypeFunction{Name: name, Index: 1, ExpectedType: ast.Literal{Value: ""}, ActualType: args[1]}
		}
		s, ok := literal.Value.(string)
		if !ok {
			return nil, ErrWrongArgTypeFunction{Name: name, Index: 1, ExpectedType: "", ActualType: args[1]}
		}
		rg, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		return ast.FuncSearch{Regex: rg}, nil
	case grammar.FuncValue:
		if len(args) != 1 {
			return nil, ErrWrongArgsCountFunction{Name: name, Expected: 1, Actual: len(args)}
		}
		return ast.FuncValue{Expr: args[0]}, nil
	default:
		return nil, ErrUnsupportedFunction{Name: name}
	}
}
