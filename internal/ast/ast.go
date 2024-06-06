package ast

import "regexp"

// Value is the leaf values of a JSON structure.
//
// numbers | text strings | null | true | false | JSON objects     | arrays
//
// float64 | string       | nil  | true | false | map[string]Value | []Value
type Value any

// Location is the position of a Value in a JSON structure.
type Location string

// Node contains the Value of a JSON along with its Location.
type Node struct {
	Location Location
	Value    Value
}

// QueryArgument is the JSON applied to a JSONPath query.
type QueryArgument interface{}

type QueryJSONPath struct {
	Segments []Expr
}

func (q QueryJSONPath) Apply(arg QueryArgument) []Node {
	panic("unimplemented")
}

// Expr is an expression that maps []Node -> []Node.
//
// Expr takes in 0..n nodes and outputs 0..n nodes.
type Expr interface {
	Evaluate([]Node) []Node
}

type SegmentChild struct {
	Selectors []Expr
}

func (s SegmentChild) Evaluate(input []Node) []Node {
	panic("unimplemented")
}

type SegmentDescendant struct {
	Selectors []Expr
}

func (s SegmentDescendant) Evaluate(input []Node) []Node {
	panic("unimplemented")
}

type SelectorName struct {
	Name string
}

func (s SelectorName) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type SelectorWildcard struct{}

func (s SelectorWildcard) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type SelectorSlice struct {
	Start  int
	End    int
	Offset int
}

func (s SelectorSlice) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type SelectorIndex struct {
	Index int
}

func (s SelectorIndex) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type SelectorFilter struct {
	Expr ExprLogical
}

func (s SelectorFilter) Evaluate([]Node) []Node {
	panic("unimplemented")
}

// ExprLogical is an expression that maps []Node -> bool.
//
// ExprLogical takes in a list of nodes and returns a boolean value,
// behaving like a predicate.
type ExprLogical interface {
	Evaluate([]Node) bool
}

type ExprLogicalOr struct {
	Exprs []ExprLogicalAnd
}

func (e ExprLogicalOr) Evaluate([]Node) bool {
	panic("unimplemented")
}

type ExprLogicalAnd struct {
	Exprs []Expr
}

func (e ExprLogicalAnd) Evaluate([]Node) bool {
	panic("unimplemented")
}

type ExprParen struct {
	Expr ExprLogical
}

func (e ExprParen) Evaluate([]Node) bool {
	panic("unimplemented")
}

type ExprComparison struct {
	Left  ExprSingle
	Right ExprSingle
	F     func(Value, Value) bool
}

func (e ExprComparison) Evaluate([]Node) bool {
	panic("unimplemented")
}

// ExprSingle is an expression that evaluates []Node -> Node.
//
// ExprSingle is an expression that takes in a list of nodes but returns only 1 node.
type ExprSingle interface {
	Evaluate([]Node) Node
}

type LiteralNumber struct {
	Value float64
}

func (l LiteralNumber) Evaluate([]Node) Node {
	panic("unimplemented")
}

type LiteralString struct {
	Value string
}

func (l LiteralString) Evaluate([]Node) Node {
	panic("unimplemented")
}

type LiteralTrue struct{}

func (l LiteralTrue) Evaluate([]Node) Node {
	panic("unimplemented")
}

type LiteralFalse struct{}

func (l LiteralFalse) Evaluate([]Node) Node {
	panic("unimplemented")
}

type LiteralNull struct{}

func (l LiteralNull) Evaluate([]Node) Node {
	panic("unimplemented")
}

type QuerySingularRel struct {
	Segments []ExprSingle
}

func (q QuerySingularRel) Evaluate([]Node) Node {
	panic("unimplemented")
}

type QuerySingularAbs struct {
	Segments []ExprSingle
}

func (q QuerySingularAbs) Evaluate([]Node) Node {
	panic("unimplemented")
}

type SegmentName struct {
	Name string
}

func (s SegmentName) Evaluate([]Node) Node {
	panic("unimplemented")
}

type SegmentIndex struct {
	Index string
}

func (s SegmentIndex) Evaluate([]Node) Node {
	panic("unimplemented")
}

type QueryRel struct {
	Segments []Expr
}

func (q QueryRel) Evaluate([]Node) []Node {
	panic("unimplemented")
}

// ExprFunc is an expression that maps some Value -> Value.
//
// ExprFunc takes in some value -> value depending on the
// type of function it is.
type ExprFunc interface {
	Evaluate(Value) Value
}

type FuncLength struct{}

func (f FuncLength) Evaluate(Value) Value {
	panic("unimplemented")
}

type FuncCount struct{}

func (f FuncCount) Evaluate(Value) Value {
	panic("unimplemented")
}

type FuncMatch struct {
	Regex regexp.Regexp
}

func (f FuncMatch) Evaluate(Value) Value {
	panic("unimplemented")
}

type FuncSearch struct {
	Regex regexp.Regexp
}

func (f FuncSearch) Evaluate(Value) Value {
	panic("unimplemented")
}

type FuncValue struct{}

func (f FuncValue) Evaluate(Value) Value {
	panic("unimplemented")
}
