package ast_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/marcfyk/go-jsonpath/internal/ast"
	"github.com/marcfyk/go-jsonpath/internal/iter"
	"github.com/stretchr/testify/assert"
)

func TestFromJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected []ast.Node
	}{
		{
			json:     "null",
			expected: []ast.Node{{Location: "$", Value: nil}},
		},
		{
			json:     "\"hello world!\"",
			expected: []ast.Node{{Location: "$", Value: "hello world!"}},
		},
		{
			json:     "123",
			expected: []ast.Node{{Location: "$", Value: float64(123)}},
		},
		{
			json:     "[]",
			expected: []ast.Node{{Location: "$", Value: []any{}}},
		},
		{
			json: "[3,1,2]",
			expected: []ast.Node{
				{Location: "$", Value: []any{float64(3), float64(1), float64(2)}},
				{Location: "$[0]", Value: float64(3)},
				{Location: "$[1]", Value: float64(1)},
				{Location: "$[2]", Value: float64(2)},
			},
		},
		{
			json:     "{}",
			expected: []ast.Node{{Location: "$", Value: map[string]any{}}},
		},
		{
			json: "{\"a\": 1}",
			expected: []ast.Node{
				{Location: "$", Value: map[string]any{"a": float64(1)}},
				{Location: "$[a]", Value: float64(1)},
			},
		},
		{
			json: "[1, [null, \"abc\"]]",
			expected: []ast.Node{
				{Location: "$", Value: []any{float64(1), []any{nil, "abc"}}},
				{Location: "$[0]", Value: float64(1)},
				{Location: "$[1]", Value: []any{nil, "abc"}},
				{Location: "$[1][0]", Value: nil},
				{Location: "$[1][1]", Value: "abc"},
			},
		},
		{
			json: "{\"a\": {\"b\": 1}}",
			expected: []ast.Node{
				{Location: "$", Value: map[string]any{"a": map[string]any{"b": float64(1)}}},
				{Location: "$[a]", Value: map[string]any{"b": float64(1)}},
				{Location: "$[a][b]", Value: float64(1)},
			},
		},
	}
	for _, test := range tests {
		var j any
		err := json.Unmarshal([]byte(test.json), &j)
		assert.Nil(t, err)
		name := fmt.Sprintf("%v = %v", test.json, test.expected)
		t.Run(name, func(t *testing.T) {
			x := ast.FromJSON(j)
			assert.Equal(t, test.expected, iter.ToSlice(x))
		})
	}
}
