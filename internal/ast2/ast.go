package ast2

import "github.com/marcfyk/go-jsonpath/internal/iter"

type Location string

type Node struct {
	Location Location
	Value    any
}

type Selector interface {
	Select(iter.Iterator[Node]) iter.Iterator[Node]
}

type LogicExpr interface {
	Evaluate(Node) bool
}

type Segment struct {
	Selectors []Selector
}

func (s Segment) Filter(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	return iter.Flatmap(nodes, func(n Node) iter.Iterator[Node] {
		selectors := iter.FromSlice(s.Selectors)
		return iter.Flatmap(selectors, func(s Selector) iter.Iterator[Node] {
			return s.Select(iter.Singleton(n))
		})
	})
}
