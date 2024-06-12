package iter_test

import (
	"fmt"
	"testing"

	"github.com/marcfyk/go-jsonpath/internal/iter"
	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	assert.Empty(t, iter.ToSlice(iter.Empty[int]()))
}

func TestSingleton(t *testing.T) {
	assert.Equal(t, []int{1}, iter.ToSlice(iter.Singleton(1)))
}

func TestMap(t *testing.T) {
	addOne := func(x int) int { return x + 1 }
	tests := []struct {
		xs       []int
		f        func(int) int
		expected []int
	}{
		{
			xs:       make([]int, 0),
			f:        addOne,
			expected: make([]int, 0),
		},
		{
			xs:       []int{2, 0, 3},
			f:        addOne,
			expected: []int{3, 1, 4},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%+v", test)
		t.Run(name, func(t *testing.T) {
			xs := iter.FromSlice(test.xs)
			ys := iter.Map(xs, addOne)
			zs := iter.ToSlice(ys)
			assert.Equal(t, test.expected, zs)
		})
	}
}

func TestFilter(t *testing.T) {
	isEven := func(x int) bool { return x%2 == 0 }
	tests := []struct {
		xs       []int
		f        func(int) bool
		expected []int
	}{
		{
			xs:       make([]int, 0),
			f:        isEven,
			expected: make([]int, 0),
		},
		{
			xs:       []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			f:        isEven,
			expected: []int{2, 4, 6, 8, 10},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			xs := iter.FromSlice(test.expected)
			ys := iter.Filter(xs, isEven)
			zs := iter.ToSlice(ys)
			assert.Equal(t, test.expected, zs)
		})
	}
}

func TestChain(t *testing.T) {
	tests := []struct {
		xs       []int
		ys       []int
		expected []int
	}{
		{
			xs:       make([]int, 0),
			ys:       make([]int, 0),
			expected: make([]int, 0),
		},
		{
			xs:       []int{2, 1, 3},
			ys:       make([]int, 0),
			expected: []int{2, 1, 3},
		},
		{
			xs:       make([]int, 0),
			ys:       []int{2, 1, 3},
			expected: []int{2, 1, 3},
		},
		{
			xs:       []int{2, 1, 3},
			ys:       []int{0, 3, 2, 1, 1},
			expected: []int{2, 1, 3, 0, 3, 2, 1, 1},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v + %v = %v", test.xs, test.ys, test.expected)
		t.Run(name, func(t *testing.T) {
			actual := iter.ToSlice(iter.Chain(iter.FromSlice(test.xs), iter.FromSlice(test.ys)))
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFlatMap(t *testing.T) {
	double := func(n int) iter.Iterator[int] { return iter.FromSlice([]int{n, n}) }
	tests := []struct {
		xs       []int
		f        func(int) iter.Iterator[int]
		expected []int
	}{
		{
			xs:       make([]int, 0),
			f:        double,
			expected: make([]int, 0),
		},
		{
			xs:       []int{2, 1, 3},
			f:        double,
			expected: []int{2, 2, 1, 1, 3, 3},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v", test.xs)
		t.Run(name, func(t *testing.T) {
			ys := iter.FromSlice(test.xs)
			zs := iter.Flatmap(ys, test.f)
			assert.Equal(t, test.expected, iter.ToSlice(zs))
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		xs       [][]int
		expected []int
	}{
		{
			xs:       make([][]int, 0),
			expected: make([]int, 0),
		},
		{
			xs: [][]int{
				{},
				{2, 1},
				{},
				{3},
				{},
			},
			expected: []int{2, 1, 3},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v", test.xs)
		t.Run(name, func(t *testing.T) {
			xs := iter.Map(iter.FromSlice(test.xs), iter.FromSlice)
			assert.Equal(t, test.expected, iter.ToSlice(iter.Flatten(xs)))
		})
	}
}

func TestZip(t *testing.T) {
	tests := []struct {
		xs       []int
		ys       []int
		expected []iter.Pair[int, int]
	}{
		{
			xs:       make([]int, 0),
			ys:       make([]int, 0),
			expected: make([]iter.Pair[int, int], 0),
		},
		{
			xs:       make([]int, 0),
			ys:       []int{2, 1, 3},
			expected: make([]iter.Pair[int, int], 0),
		},
		{
			xs:       []int{2, 1, 3},
			ys:       make([]int, 0),
			expected: make([]iter.Pair[int, int], 0),
		},
		{
			xs:       []int{2, 1, 3, 4, 5},
			ys:       []int{5, 4, 3, 7, 8},
			expected: []iter.Pair[int, int]{{2, 5}, {1, 4}, {3, 3}, {4, 7}, {5, 8}},
		},
		{
			xs:       []int{2, 1, 3},
			ys:       []int{5, 4, 3, 7, 8},
			expected: []iter.Pair[int, int]{{2, 5}, {1, 4}, {3, 3}},
		},

		{
			xs:       []int{2, 1, 3, 4, 5},
			ys:       []int{5, 4, 3},
			expected: []iter.Pair[int, int]{{2, 5}, {1, 4}, {3, 3}},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("zip %v %v = %v", test.xs, test.ys, test.expected)
		t.Run(name, func(t *testing.T) {
			zipped := iter.Zip(iter.FromSlice(test.xs), iter.FromSlice(test.ys))
			assert.Equal(t, test.expected, iter.ToSlice(zipped))
		})
	}
}

func TestIterate(t *testing.T) {
	addOne := func(i int) int { return i + 1 }
	take := 5
	tests := []struct {
		seed     int
		f        func(int) int
		expected []int
	}{
		{
			seed:     0,
			f:        addOne,
			expected: []int{0, 1, 2, 3, 4},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("seed = %v", test.seed)
		t.Run(name, func(t *testing.T) {
			xs := iter.Iterate(test.seed, test.f)
			actual := make([]int, take)
			for i := 0; i < take; i++ {
				x := xs()
				assert.NotNil(t, x)
				actual[i] = *x
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEnumerate(t *testing.T) {
	tests := []struct {
		xs       []int
		expected []iter.Pair[int, int]
	}{
		{
			xs:       make([]int, 0),
			expected: make([]iter.Pair[int, int], 0),
		},
		{
			xs:       []int{2, 1, 3},
			expected: []iter.Pair[int, int]{{0, 2}, {1, 1}, {2, 3}},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("enumerate %v = %v", test.xs, test.expected)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, iter.ToSlice(iter.Enumerate(iter.FromSlice(test.xs))))
		})
	}
}
