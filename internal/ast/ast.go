// Package ast implements an abstract syntax tree that serves as an IR of a jsonpath query.
package ast

import (
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/marcfyk/go-jsonpath/internal/iter"
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

// Expr is an expression that maps []Node -> []Node.
//
// Expr takes in 0..n nodes and outputs 0..n nodes.
type Expr interface {
	Evaluate(iter.Iterator[Node]) iter.Iterator[Node]
}

// ExprLogical is an expression that maps []Node -> bool.
//
// ExprLogical takes in a list of nodes and returns a boolean value,
// behaving like a predicate.
type ExprLogical interface {
	EvaluateLogical(iter.Iterator[Node]) bool
}

// ExprSingle is an expression that evaluates []Node -> Node.
//
// ExprSingle is an expression that takes in a list of nodes but returns only 1 node.
type ExprSingle interface {
	EvaluateSingle(iter.Iterator[Node]) Node
}

type QueryJSONPath struct {
	Segments []Expr
}

func (q QueryJSONPath) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	output := nodes
	for _, s := range q.Segments {
		output = s.Evaluate(output)
	}
	return output
}

type SegmentChild struct {
	Selectors []Expr
}

func (s SegmentChild) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	output := nodes
	for _, s := range s.Selectors {
		output = s.Evaluate(output)
	}
	return output
}

type SegmentDescendant struct {
	Selectors []Expr
}

func (s SegmentDescendant) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	output := nodes
	output = iter.Flatmap(output, func(n Node) iter.Iterator[Node] {
		return FromJSON(n.Value)
	})
	for _, s := range s.Selectors {
		output = s.Evaluate(output)
	}
	return output
}

type SelectorName struct {
	Name string
}

func (s SelectorName) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	return iter.Flatmap(nodes, func(n Node) iter.Iterator[Node] {
		switch v := n.Value.(type) {
		case map[string]any:
			value, ok := v[s.Name]
			if !ok {
				return iter.Empty[Node]()
			}
			location := Location(fmt.Sprintf("%s[\"%s\"]", n.Location, s.Name))
			node := Node{Location: location, Value: value}
			return iter.Singleton(node)
		default:
			return iter.Empty[Node]()
		}
	})
}

type SelectorWildcard struct{}

func (s SelectorWildcard) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	return iter.Flatmap(nodes, func(n Node) iter.Iterator[Node] {
		switch v := n.Value.(type) {
		case map[string]any:
			pairs := iter.FromMap(v)
			return iter.Map(pairs, func(p iter.Pair[string, any]) Node {
				location := Location(fmt.Sprintf("%s[\"%s\"]", n.Location, p.Left))
				return Node{Location: location, Value: p.Right}
			})
		case []any:
			indexed := iter.Enumerate(iter.FromSlice(v))
			return iter.Map(indexed, func(p iter.Pair[int, any]) Node {
				location := Location(fmt.Sprintf("%s[%d]", n.Location, p.Left))
				return Node{Location: location, Value: p.Right}
			})
		default:
			return iter.Empty[Node]()
		}
	})
}

type SelectorSlice struct {
	Start *int
	End   *int
	Step  int
}

func (s SelectorSlice) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	normalize := func(i, size int) int {
		if i < 0 {
			return size + i
		}
		return i
	}
	bounds := func(start, end, step, size int) (lower, upper int) {
		nStart := normalize(start, size)
		nEnd := normalize(end, size)
		if step < 0 {
			upper = min(max(nStart, -1), size-1)
			lower = min(max(nEnd, -1), size-1)
		} else {
			lower = min(max(nStart, 0), size)
			upper = min(max(nEnd, 0), size)
		}
		return lower, upper
	}
	return iter.Flatmap(nodes, func(n Node) iter.Iterator[Node] {
		switch v := n.Value.(type) {
		case []any:
			if s.Step == 0 {
				return iter.Empty[Node]()
			}
			start, end := 0, len(v)
			if s.Step < 0 {
				start, end = len(v)-1, -len(v)-1
			}
			if s.Start != nil {
				start = *s.Start
			}
			if s.End != nil {
				end = *s.End
			}
			lower, upper := bounds(start, end, s.Step, len(v))
			if upper < lower {
				return iter.Empty[Node]()
			}
			span := max(upper, lower) - min(upper, lower)
			absStep := max(s.Step, -s.Step)
			size := span / absStep
			if span%absStep > 0 {
				size++
			}
			result := make([]Node, size)
			index := 0
			if s.Step < 0 {
				for i := upper; i > lower; i += s.Step {
					location := Location(fmt.Sprintf("%s[%d]", n.Location, i))
					result[index] = Node{Location: location, Value: v[i]}
					index++
				}
			} else {
				for i := lower; i < upper; i += s.Step {
					location := Location(fmt.Sprintf("%s[%d]", n.Location, i))
					result[index] = Node{Location: location, Value: v[i]}
					index++
				}
			}
			return iter.FromSlice(result)
		default:
			return iter.Empty[Node]()
		}
	})
}

type SelectorIndex struct {
	Index int
}

func (s SelectorIndex) Evaluate(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	return iter.Flatmap(nodes, func(n Node) iter.Iterator[Node] {
		switch v := n.Value.(type) {
		case []any:
			index := s.Index
			if index < 0 {
				index += len(v)
			}
			if index < 0 || index >= len(v) {
				return iter.Empty[Node]()
			}
			location := Location(fmt.Sprintf("%s[%d]", n.Location, index))
			node := Node{Location: location, Value: v[index]}
			return iter.Singleton(node)
		default:
			return iter.Empty[Node]()
		}
	})
}

func (s SelectorIndex) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type SelectorFilter struct {
	Expr ExprLogical
}

func (s SelectorFilter) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type ExprLogicalOr struct {
	Exprs []Expr
}

func (e ExprLogicalOr) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type ExprLogicalAnd struct {
	Exprs []Expr
}

func (e ExprLogicalAnd) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type ExprLogicalNot struct {
	Expr Expr
}

func (e ExprLogicalNot) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type ExprParen struct {
	Expr Expr
}

func (e ExprParen) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
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

func (e ExprComparison) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type Literal struct {
	Value Value
}

func (l Literal) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

func (l Literal) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type QuerySingularRel struct {
	Segments []ExprSingle
}

func (q QuerySingularRel) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

func (q QuerySingularRel) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type QuerySingularAbs struct {
	Segments []ExprSingle
}

func (q QuerySingularAbs) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

func (q QuerySingularAbs) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type SegmentName struct {
	Name string
}

func (s SegmentName) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type SegmentIndex struct {
	Index int
}

func (s SegmentIndex) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type QueryRel struct {
	Segments []Expr
}

func (q QueryRel) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

type FuncLength struct{}

func (f FuncLength) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncLength) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

func (f FuncLength) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type FuncCount struct{}

func (f FuncCount) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncCount) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

func (f FuncCount) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type FuncMatch struct {
	Regex *regexp.Regexp
}

func (f FuncMatch) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncMatch) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

func (f FuncMatch) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type FuncSearch struct {
	Regex *regexp.Regexp
}

func (f FuncSearch) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncSearch) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

func (f FuncSearch) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}

type FuncValue struct {
	Expr Expr
}

func (f FuncValue) EvaluateFunc(Value) Value {
	panic("unimplemented")
}

func (f FuncValue) Evaluate(iter.Iterator[Node]) iter.Iterator[Node] {
	panic("unimplemented")
}

func (f FuncValue) EvaluateSingle(iter.Iterator[Node]) Node {
	panic("unimplemented")
}
