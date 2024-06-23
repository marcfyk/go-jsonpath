package ast2_test

import (
	"fmt"
	"testing"

	"github.com/marcfyk/go-jsonpath/internal/ast"
	"github.com/marcfyk/go-jsonpath/internal/iter"
	"github.com/stretchr/testify/assert"
)

func ptr[A any](a A) *A { return &a }

func TestSelectorNameEvaluate(t *testing.T) {
	tests := []struct {
		selector ast.SelectorName
		input    []ast.Node
		expected []ast.Node
	}{
		{
			selector: ast.SelectorName{Name: "a"},
			input:    make([]ast.Node, 0),
			expected: make([]ast.Node, 0),
		},
		{
			selector: ast.SelectorName{Name: "a"},
			input: []ast.Node{
				{Location: "$", Value: nil},
				{Location: "$", Value: float64(1)},
				{Location: "$", Value: "a"},
				{Location: "$", Value: []any{float64(1), "a", nil}},
				{Location: "$", Value: map[string]any{"a": float64(1), "b": nil, "c": "abc"}},
				{Location: "$", Value: map[string]any{"a": "b"}},
				{Location: "$", Value: map[string]any{"a": nil}},
				{Location: "$", Value: map[string]any{"b": "abc"}},
				{Location: "$", Value: map[string]any{"a": []any{float64(1), "a", nil}}},
			},
			expected: []ast.Node{
				{Location: "$[\"a\"]", Value: float64(1)},
				{Location: "$[\"a\"]", Value: "b"},
				{Location: "$[\"a\"]", Value: nil},
				{Location: "$[\"a\"]", Value: []any{float64(1), "a", nil}},
			},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v | %s = %v", test.input, test.selector, test.expected)
		t.Run(name, func(t *testing.T) {
			nodes := iter.FromSlice(test.input)
			output := test.selector.Evaluate(nodes)
			assert.Equal(t, test.expected, iter.ToSlice(output))
		})
	}
}

func TestSelectorWildcardEvaluate(t *testing.T) {
	tests := []struct {
		input    []ast.Node
		expected []ast.Node
	}{
		{
			input:    make([]ast.Node, 0),
			expected: make([]ast.Node, 0),
		},
		{
			input: []ast.Node{
				{Location: "$", Value: nil},
				{Location: "$", Value: float64(1)},
				{Location: "$", Value: "a"},
				{Location: "$", Value: []any{float64(1), "a", nil}},
				{Location: "$", Value: map[string]any{"a": float64(1)}},
				{Location: "$", Value: map[string]any{"a": []any{float64(1), "a", nil}}},
			},
			expected: []ast.Node{
				{Location: "$[0]", Value: float64(1)},
				{Location: "$[1]", Value: "a"},
				{Location: "$[2]", Value: nil},
				{Location: "$[\"a\"]", Value: float64(1)},
				{Location: "$[\"a\"]", Value: []any{float64(1), "a", nil}},
			},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v | [wildcard selector] = %v", test.input, test.expected)
		t.Run(name, func(t *testing.T) {
			nodes := iter.FromSlice(test.input)
			output := ast.SelectorWildcard{}.Evaluate(nodes)
			assert.Equal(t, test.expected, iter.ToSlice(output))
		})
	}
}

func TestSelectorSliceEvaluate(t *testing.T) {
	tests := []struct {
		selector ast.SelectorSlice
		input    []ast.Node
		expected []ast.Node
	}{
		{
			selector: ast.SelectorSlice{Start: nil, End: nil, Step: 0},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: make([]ast.Node, 0),
		},
		{
			selector: ast.SelectorSlice{Start: nil, End: nil, Step: 1},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{
				{Location: "$[0]", Value: "a"},
				{Location: "$[1]", Value: "b"},
				{Location: "$[2]", Value: "c"},
				{Location: "$[3]", Value: "d"},
				{Location: "$[4]", Value: "e"},
				{Location: "$[5]", Value: "f"},
				{Location: "$[6]", Value: "g"},
			},
		},
		{
			selector: ast.SelectorSlice{Start: nil, End: nil, Step: -1},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{
				{Location: "$[6]", Value: "g"},
				{Location: "$[5]", Value: "f"},
				{Location: "$[4]", Value: "e"},
				{Location: "$[3]", Value: "d"},
				{Location: "$[2]", Value: "c"},
				{Location: "$[1]", Value: "b"},
				{Location: "$[0]", Value: "a"},
			},
		},
		{
			selector: ast.SelectorSlice{Start: ptr(1), End: ptr(3), Step: 1},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{
				{Location: "$[1]", Value: "b"},
				{Location: "$[2]", Value: "c"},
			},
		},
		{
			selector: ast.SelectorSlice{Start: ptr(5), End: nil, Step: 1},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{
				{Location: "$[5]", Value: "f"},
				{Location: "$[6]", Value: "g"},
			},
		},
		{
			selector: ast.SelectorSlice{Start: ptr(1), End: ptr(5), Step: 2},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{
				{Location: "$[1]", Value: "b"},
				{Location: "$[3]", Value: "d"},
			},
		},
		{
			selector: ast.SelectorSlice{Start: ptr(5), End: ptr(1), Step: -2},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{
				{Location: "$[5]", Value: "f"},
				{Location: "$[3]", Value: "d"},
			},
		},
		{
			selector: ast.SelectorSlice{Start: ptr(1), End: ptr(5), Step: -1},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{},
		},
		{
			selector: ast.SelectorSlice{Start: ptr(5), End: ptr(5), Step: 1},
			input:    []ast.Node{{Location: "$", Value: []any{"a", "b", "c", "d", "e", "f", "g"}}},
			expected: []ast.Node{},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%s | %+v = %+v", test.input, test.selector, test.expected)
		t.Run(name, func(t *testing.T) {
			nodes := iter.FromSlice(test.input)
			output := test.selector.Evaluate(nodes)
			assert.Equal(t, test.expected, iter.ToSlice(output))
		})
	}
}

func TestSelectorIndexEvaluate(t *testing.T) {
	tests := []struct {
		selector ast.SelectorIndex
		input    []ast.Node
		expected []ast.Node
	}{
		{
			selector: ast.SelectorIndex{Index: 1},
			input:    make([]ast.Node, 0),
			expected: make([]ast.Node, 0),
		},
		{
			selector: ast.SelectorIndex{Index: 1},
			input: []ast.Node{
				{Location: "$", Value: nil},
				{Location: "$", Value: float64(1)},
				{Location: "$", Value: "a"},
				{Location: "$", Value: []any{float64(1), "a", nil}},
				{Location: "$", Value: map[string]any{"a": nil}},
				{Location: "$", Value: []any{1, []any{float64(1), "a", nil}}},
			},
			expected: []ast.Node{
				{Location: "$[1]", Value: "a"},
				{Location: "$[1]", Value: []any{float64(1), "a", nil}},
			},
		},
		{
			selector: ast.SelectorIndex{Index: -2},
			input: []ast.Node{
				{Location: "$", Value: []any{float64(1), "a", nil}},
				{Location: "$", Value: []any{float64(1), []any{float64(1), "a", nil}}},
			},
			expected: []ast.Node{
				{Location: "$[1]", Value: "a"},
				{Location: "$[0]", Value: float64(1)},
			},
		},
		{
			selector: ast.SelectorIndex{Index: 200},
			input: []ast.Node{
				{Location: "$", Value: []any{float64(1), "a", nil}},
				{Location: "$", Value: []any{float64(1), []any{float64(1), "a", nil}}},
			},
			expected: []ast.Node{},
		},
		{
			selector: ast.SelectorIndex{Index: -200},
			input: []ast.Node{
				{Location: "$", Value: []any{float64(1), "a", nil}},
				{Location: "$", Value: []any{float64(1), []any{float64(1), "a", nil}}},
			},
			expected: []ast.Node{},
		},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v | %v = %v", test.input, test.selector, test.expected)
		t.Run(name, func(t *testing.T) {
			nodes := iter.FromSlice(test.input)
			output := test.selector.Evaluate(nodes)
			assert.Equal(t, test.expected, iter.ToSlice(output))
		})
	}
}
