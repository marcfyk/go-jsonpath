package iter

// FromMap creates an iterator over elements in a map.
// This operation is O(n) as it is not possible to lazily evaluate
// values in a map unless we use a go routine.
func FromMap[K comparable, V any](m map[K]V) Iterator[Pair[K, V]] {
	pairs := make([]Pair[K, V], len(m))
	i := 0
	for k, v := range m {
		pairs[i] = Pair[K, V]{Left: k, Right: v}
		i++
	}
	return FromSlice(pairs)
}

// ToMap collects an iterator into a map.
// Duplicate keys will be overwritten with the latest key, value pair from the iterator.
func ToMap[K comparable, V any](kvs Iterator[Pair[K, V]]) map[K]V {
	m := make(map[K]V)
	for kv := kvs(); kv != nil; kv = kvs() {
		m[kv.Left] = kv.Right
	}
	return m
}
