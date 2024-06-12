// Package iter implements an iterator package over values.
// The iterator yields values one at a time and returns nil when it is completed.
//
// This package is not meant to be a full feature rich iterator package,
// and will just support lazy iterator operations as much as needed to facilitate
// jsonpath ast evaluation lazily.
package iter

// Pair is the wrapper struct over a tuple of values.
type Pair[K any, V any] struct {
	Left  K
	Right V
}

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

// Flatmap applies maps a function, f, over an iterator and flattens its output
// as part of the output iterator.
//
// This is equivalent to a monadic bind operator, if we assume the iterator is a monad.
func Flatmap[A any](xs Iterator[A], f func(A) Iterator[A]) Iterator[A] {
	return Flatten(Map(xs, f))
}

// Flatten flattens a nested iterator.
//
// This is equivalent to a monadic join operator, if we assume the iterator is a monad.
func Flatten[A any](xs Iterator[Iterator[A]]) Iterator[A] {
	var ys Iterator[A]
	return func() *A {
		if ys == nil {
			x := xs()
			if x == nil {
				return nil
			}
			ys = *x
		}
		for {
			y := ys()
			if y != nil {
				return y
			}
			x := xs()
			if x == nil {
				return nil
			}
			ys = *x
		}
	}
}

// Zip combines two iterators into an output iterator that emits pairs of values from each
// of the iterators.
//
// The output iterator terminates when the shorter of the two input iterator terminates.
func Zip[A, B any](xs Iterator[A], ys Iterator[B]) Iterator[Pair[A, B]] {
	return func() *Pair[A, B] {
		x := xs()
		if x == nil {
			return nil
		}
		y := ys()
		if y == nil {
			return nil
		}
		return &Pair[A, B]{Left: *x, Right: *y}
	}
}

// Iterate returns an infinite iterator that keeps generating values based on an
// initial seed value, and subsequently applies f on the previous emited value.
func Iterate[A any](seed A, f func(A) A) Iterator[A] {
	x := seed
	return func() *A {
		result := x
		x = f(x)
		return &result
	}
}

// Enumerate returns an iterator with its associated index as a pair.
//
// This is mainly having indexed functionality while using iterators.
func Enumerate[A any](xs Iterator[A]) Iterator[Pair[int, A]] {
	indexes := Iterate(0, func(i int) int { return i + 1 })
	return Zip(indexes, xs)
}
