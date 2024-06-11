package iter

// MapPair is the wrapper struct over key value pairs in a map.
type MapPair[K comparable, V any] struct {
	Key   K
	Value V
}

// FromMap creates an iterator over elements in a map.
// This operation is O(n) as it is not possible to lazily evaluate
// values in a map unless we use a go routine.
func FromMap[K comparable, V any](m map[K]V) Iterator[MapPair[K, V]] {
	pairs := make([]MapPair[K, V], len(m))
	i := 0
	for k, v := range m {
		pairs[i] = MapPair[K, V]{Key: k, Value: v}
		i++
	}
	return FromSlice(pairs)
}

// ToMap collects an iterator into a map.
// Duplicate keys will be overwritten with the latest key, value pair from the iterator.
func ToMap[K comparable, V any](kvs Iterator[MapPair[K, V]]) map[K]V {
	m := make(map[K]V)
	for kv := kvs(); kv != nil; kv = kvs() {
		m[kv.Key] = kv.Value
	}
	return m
}
