package ast

import (
	"fmt"

	"github.com/marcfyk/go-jsonpath/internal/iter"
)

// FromJSON returns an iterator over a JSON structure's values.
func FromJSON(value any) iter.Iterator[Node] {
	xs := iter.Singleton(Node{
		Location: "$",
		Value:    value,
	})
	return func() *Node {
		for {
			x := xs()
			if x == nil {
				return nil
			}
			rest := jsonValueChildren(*x)
			xs = iter.Chain(xs, rest)
			return x
		}
	}
}

func jsonValueChildren(node Node) iter.Iterator[Node] {
	switch node.Value.(type) {
	case []any:
		values := iter.FromSlice(node.Value.([]any))
		enumeratedValues := iter.Enumerate(values)
		return iter.Map(enumeratedValues, func(p iter.Pair[int, any]) Node {
			return Node{
				Location: Location(fmt.Sprintf("%s[%d]", node.Location, p.Left)),
				Value:    p.Right,
			}
		})
	case map[string]any:
		xs := iter.FromMap(node.Value.(map[string]any))
		return iter.Map(xs, func(p iter.Pair[string, any]) Node {
			return Node{
				Location: Location(fmt.Sprintf("%s[%s]", node.Location, p.Left)),
				Value:    p.Right,
			}
		})
	default:
		return iter.Empty[Node]()
	}
}
