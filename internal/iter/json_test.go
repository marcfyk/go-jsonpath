package iter_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/marcfyk/go-jsonpath/internal/iter"
	"github.com/stretchr/testify/assert"
)

func TestFromJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected []any
	}{
		{
			json:     "null",
			expected: []any{nil},
		},
		{
			json:     "\"hello world!\"",
			expected: []any{"hello world!"},
		},
		{
			json:     "123",
			expected: []any{float64(123)},
		},
		{
			json:     "[]",
			expected: []any{},
		},
		{
			json:     "[3,1,2]",
			expected: []any{float64(3), float64(1), float64(2)},
		},
		{
			json:     "{}",
			expected: []any{},
		},
		{
			json:     "{\"a\": 1}",
			expected: []any{float64(1)},
		},
		{
			json:     "[1, [null, \"abc\"]]",
			expected: []any{float64(1), nil, "abc"},
		},
		{
			json:     "{\"a\": {\"b\": 1}}",
			expected: []any{float64(1)},
		},
	}
	for _, test := range tests {
		var j any
		err := json.Unmarshal([]byte(test.json), &j)
		assert.Nil(t, err)
		name := fmt.Sprintf("%v = %v", test.json, test.expected)
		t.Run(name, func(t *testing.T) {
			x := iter.FromJSON(j)
			assert.Equal(t, test.expected, iter.ToSlice(x))
		})
	}
}
