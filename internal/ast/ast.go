// Package ast implements an abstract syntax tree that serves as an IR of a jsonpath query.
package ast

import (
	"maps"
	"regexp"
	"slices"
)

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

type QueryJSONPath struct {
	Segments []Expr
}

func (q QueryJSONPath) Evaluate([]Node) []Node {
	panic("unimplemented")
}

// Expr is an expression that maps []Node -> []Node.
//
// Expr takes in 0..n nodes and outputs 0..n nodes.
type Expr interface {
	Evaluate([]Node) []Node
}

// ExprLogical is an expression that maps []Node -> bool.
//
// ExprLogical takes in a list of nodes and returns a boolean value,
// behaving like a predicate.
type ExprLogical interface {
	EvaluateLogical([]Node) bool
}

// ExprSingle is an expression that evaluates []Node -> Node.
//
// ExprSingle is an expression that takes in a list of nodes but returns only 1 node.
type ExprSingle interface {
	EvaluateSingle([]Node) Node
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

func (s SelectorName) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type SelectorWildcard struct{}

func (s SelectorWildcard) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type SelectorSlice struct {
	Start int
	End   int
	Step  int
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

func (s SelectorIndex) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type SelectorFilter struct {
	Expr ExprLogical
}

func (s SelectorFilter) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type ExprLogicalOr struct {
	Exprs []Expr
}

func (e ExprLogicalOr) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type ExprLogicalAnd struct {
	Exprs []Expr
}

func (e ExprLogicalAnd) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type ExprLogicalNot struct {
	Expr Expr
}

func (e ExprLogicalNot) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type ExprParen struct {
	Expr Expr
}

func (e ExprParen) Evaluate([]Node) []Node {
	panic("unimplemented")
}

var (
	EQ = func(v1, v2 Value) bool {
		switch v1 := v1.(type) {
		case float64:
			v2, ok := v2.(float64)
			return ok && v1 == v2
		case string:
			v2, ok := v2.(string)
			return ok && v1 == v2
		case bool:
			v2, ok := v2.(bool)
			return ok && v1 == v2
		case []any:
			v2, ok := v2.([]any)
			return ok && slices.Equal(v1, v2)
		case map[string]any:
			v2, ok := v2.(map[string]any)
			return ok && maps.Equal(v1, v2)
		case nil:
			return v2 == nil
		default:
			return false
		}
	}
	NE = func(v1, v2 Value) bool {
		return !EQ(v1, v2)
	}

	LT = func(v1, v2 Value) bool {
		switch v1 := v1.(type) {
		case float64:
			v2, ok := v2.(float64)
			return ok && v1 < v2
		case string:
			v2, ok := v2.(string)
			return ok && v1 < v2
		default:
			return false
		}
	}

	LTE = func(v1, v2 Value) bool {
		return EQ(v1, v2) || LT(v1, v2)
	}

	GT = func(v1, v2 Value) bool {
		return !LTE(v1, v2)
	}

	GTE = func(v1, v2 Value) bool {
		return !LT(v1, v2)
	}
)

type ExprComparison struct {
	Left  ExprSingle
	Right ExprSingle
	F     func(Value, Value) bool
}

func (e ExprComparison) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type Literal struct {
	Value Value
}

func (l Literal) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

func (l Literal) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type QuerySingularRel struct {
	Segments []ExprSingle
}

func (q QuerySingularRel) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

func (q QuerySingularRel) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type QuerySingularAbs struct {
	Segments []ExprSingle
}

func (q QuerySingularAbs) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

func (q QuerySingularAbs) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type SegmentName struct {
	Name string
}

func (s SegmentName) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type SegmentIndex struct {
	Index int
}

func (s SegmentIndex) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type QueryRel struct {
	Segments []Expr
}

func (q QueryRel) Evaluate([]Node) []Node {
	panic("unimplemented")
}

type FuncLength struct{}

func (f FuncLength) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncLength) Evaluate([]Node) []Node {
	panic("unimplemented")
}

func (f FuncLength) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type FuncCount struct{}

func (f FuncCount) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncCount) Evaluate([]Node) []Node {
	panic("unimplemented")
}

func (f FuncCount) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type FuncMatch struct {
	Regex *regexp.Regexp
}

func (f FuncMatch) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncMatch) Evaluate([]Node) []Node {
	panic("unimplemented")
}

func (f FuncMatch) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type FuncSearch struct {
	Regex *regexp.Regexp
}

func (f FuncSearch) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncSearch) Evaluate([]Node) []Node {
	panic("unimplemented")
}

func (f FuncSearch) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}

type FuncValue struct {
	Expr Expr
}

func (f FuncValue) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncValue) Evaluate([]Node) []Node {
	panic("unimplemented")
}

func (f FuncValue) EvaluateSingle([]Node) Node {
	panic("unimplemented")
}
