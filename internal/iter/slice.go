package iter

// FromSlice constructs an iterator over elements of a slice.
// This operation is O(1).
func FromSlice[A any](xs []A) Iterator[A] {
	index := 0
	return func() *A {
		if index >= len(xs) {
			return nil
		}
		a := xs[index]
		index++
		return &a
	}
}

// ToSlice collects elements in an iterator into a slice.
// This operation is O(n).
func ToSlice[A any](xs Iterator[A]) []A {
	ys := make([]A, 0)
	for a := xs(); a != nil; a = xs() {
		ys = append(ys, *a)
	}
	return ys
}
