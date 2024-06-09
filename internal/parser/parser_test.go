package parser_test

import (
	"fmt"
	"testing"

	"github.com/marcfyk/go-jsonpath/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestRootIdent(t *testing.T) {
	p := parser.New("$")
	a, err := p.Parse()
	assert.Nil(t, err)
	assert.NotNil(t, a)
	assert.True(t, p.IsDone(), "cursor:%d", p.Index)
}

func TestNameSelectorInChildSegments(t *testing.T) {
	paths := []string{
		"$.o['j j']",
		"$.o['j j']['k.k']",
		"$.o[\"j j\"][\"k.k\"]",
		"$.o[\"'\"][\"@\"]",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(path)
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}

func TestWildcardSelectorInChildSegments(t *testing.T) {
	paths := []string{
		"$[*]",
		"$.o[*]",
		"$.o[*, *]",
		"$.a[*]",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(path)
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}

func TestIndexSelectorInChildSegments(t *testing.T) {
	paths := []string{
		"$[1]",
		"$[-2]",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(path)
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}

func TestArraySliceSelectorInChildSegment(t *testing.T) {
	paths := []string{
		"$[1:3]",
		"$[5:]",
		"$[1:5:2]",
		"$[5:1:-2]",
		"$[::-1]",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(path)
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}

func TestComparisonExpressions(t *testing.T) {
	paths := []string{
		"$.absent1 == $.absent2",
		"$.absent1 <= $.absent2",
		"$.absent == 'g'",
		"$.absent1 != $.absent2",
		"$.absent != 'g'",
		"1 <= 2",
		"1 > 2",
		"13 == '13'",
		"'a' <= 'b'",
		"'a' > 'b'",
		"$.obj == $.arr",
		"$.obj != $.arr",
		"$.obj == $.obj",
		"$.obj != $.obj",
		"$.arr == $.arr",
		"$.arr != $.arr",
		"$.obj == 17",
		"$.obj != 17",
		"$.obj <= $.arr",
		"$.obj < $.arr",
		"$.obj <= $.obj",
		"$.arr <= $.arr",
		"1 <= $.arr",
		"1 >= $.arr",
		"1 > $.arr",
		"1 < $.arr",
		"true <= true",
		"true > true",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(fmt.Sprintf("$[?%s]", path))
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}

func TestFilterSelectorInChildSelector(t *testing.T) {
	paths := []string{
		"$.a[?@.b == 'kilo']",
		"$.a[?(@.b == 'kilo')]",
		"$.a[?@>3.5]",
		"$.a[?@.b]",
		"$[?@.*]",
		"$[?@[?@.b]]",
		"$.o[?@<3, ?@<3]",
		"$.a[?@<2 || @.b == \"k\"]",
		"$.a[?match(@.b, \"[jk]\")]",
		"$.a[?search(@.b, \"[jk]\")]",
		"$.o[?@>1 && @<4]",
		"$.o[?@.u || @.x]",
		"$.a[?@.b == @.x]",
		"$.a[?@ == @]",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(path)
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}

func TestFunctionExtensions(t *testing.T) {
	paths := []string{
		"$[?length(@) < 3]",
		"$[?count(@.*) == 1]",
		"$[?count(count(@.*)) == 1]",
		"$[?match(@.timezone, 'Europe/.*')]",
		"$[?value(@..color) == \"red\"]",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			p := parser.New(path)
			a, err := p.Parse()
			assert.Nil(t, err)
			assert.NotNil(t, a)
			assert.True(t, p.IsDone(), "cursor:%d", p.Index)
		})
	}
}
