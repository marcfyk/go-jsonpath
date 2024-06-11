// Package iter implements an iterator package over values.
//
// The iterator yields values one at a time and returns nil when it is completed.
package iter

// Iterator is the main iterator type. An iterator over elements of type A will
// return *A where the dereferenced value is the iterated element.
// nil is returned when the iterator has been exhausted.
type Iterator[A any] func() *A

// Empty returns an empty iterator.
func Empty[A any]() Iterator[A] {
	return func() *A { return nil }
}

// Singleton returns an iterator with only 1 element.
func Singleton[A any](a A) Iterator[A] {
	ok := true
	return func() *A {
		if !ok {
			return nil
		}
		ok = false
		return &a
	}
}

// Map returns an iterator from another iterator that applies a function, f, to each element
// in the original iterator.
func Map[A, B any](xs Iterator[A], f func(A) B) Iterator[B] {
	return func() *B {
		if a := xs(); a != nil {
			b := f(*a)
			return &b
		}
		return nil
	}
}

// Filter returns an iterator from another iterator, only with elements, a, which evaluates to true
// when applied to a predicate, f.
func Filter[A any](xs Iterator[A], f func(A) bool) Iterator[A] {
	return func() *A {
		for a := xs(); a != nil; a = xs() {
			if f(*a) {
				return a
			}
		}
		return nil
	}
}

// Chain concatenates two iterators together without any eager evaluation.
func Chain[A any](xs, ys Iterator[A]) Iterator[A] {
	ok := true
	return func() *A {
		if ok {
			x := xs()
			if x != nil {
				return x
			}
			ok = false
			return ys()
		}
		return ys()
	}
}
