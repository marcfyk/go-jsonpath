package iter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromSliceAndToSlice(t *testing.T) {
	tests := [][]int{
		{},
		{2, 1, 3},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test, ToSlice(FromSlice(test)))
		})
	}
}
