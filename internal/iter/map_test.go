package iter_test

import (
	"fmt"
	"testing"

	"github.com/marcfyk/go-jsonpath/internal/iter"
	"github.com/stretchr/testify/assert"
)

func TestFromMapAndToMap(t *testing.T) {
	tests := []map[int]int{
		{},
		{2: 11, 1: 33, 3: 22},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test, iter.ToMap(iter.FromMap(test)))
		})
	}
}

func TestToMapDuplicate(t *testing.T) {
	tests := []struct {
		x []iter.Pair[int, int]
		y map[int]int
	}{
		{
			x: []iter.Pair[int, int]{{Left: 1, Right: 2}, {Left: 1, Right: 3}},
			y: map[int]int{1: 3},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v %v", test.x, test.y)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.y, iter.ToMap(iter.FromSlice(test.x)))
		})
	}
}
