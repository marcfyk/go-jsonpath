package ast2

import (
	"fmt"

	"github.com/marcfyk/go-jsonpath/internal/iter"
)

type SelectorName struct {
	Name string
}

func (s SelectorName) Select(nodes iter.Iterator[Node]) iter.Iterator[Node] {
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

func (s SelectorWildcard) Select(nodes iter.Iterator[Node]) iter.Iterator[Node] {
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

func (s SelectorSlice) Selector(nodes iter.Iterator[Node]) iter.Iterator[Node] {
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

func (s SelectorIndex) Select(nodes iter.Iterator[Node]) iter.Iterator[Node] {
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

type SelectorFilter struct {
	LogicExpr LogicExpr
}

func (s SelectorFilter) Select(nodes iter.Iterator[Node]) iter.Iterator[Node] {
	return iter.Filter(nodes, func(n Node) bool {
		return s.LogicExpr.Evaluate(n)
	})
}
