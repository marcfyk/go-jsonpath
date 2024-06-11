package iter

// FromJSON returns an iterator over a JSON structure's values.
func FromJSON(value any) Iterator[any] {
	xs := Singleton(value)
	return func() *any {
		for {
			x := xs()
			if x == nil {
				return nil
			}
			rest := jsonValueChildren(*x)
			xs = Chain(xs, rest)
			if y := jsonValue(*x); y != nil {
				return y
			}
		}
	}
}

func jsonValue(value any) *any {
	switch v := value.(type) {
	case float64, string, nil:
		return &v
	default:
		return nil
	}
}

func jsonValueChildren(value any) Iterator[any] {
	switch value.(type) {
	case []any:
		return FromSlice(value.([]any))
	case map[string]any:
		xs := FromMap(value.(map[string]any))
		return Map(xs, func(x MapPair[string, any]) any { return x.Value })
	default:
		return Empty[any]()
	}
}
