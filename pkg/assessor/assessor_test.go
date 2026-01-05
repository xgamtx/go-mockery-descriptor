package assessor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xgamtx/go-mockery-descriptor/pkg/assessor"
)

func TestElementsMatch(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name     string
		expected []int
		actual   []int
		wantRes  bool
	}
	tests := []testCase{
		{
			name:     "no elements",
			expected: nil,
			actual:   nil,
			wantRes:  true,
		},
		{
			name:     "one element",
			expected: []int{1},
			actual:   []int{1},
			wantRes:  true,
		},
		{
			name:     "one different element",
			expected: []int{1},
			actual:   []int{2},
			wantRes:  false,
		},
		{
			name:     "different order",
			expected: []int{1, 2},
			actual:   []int{2, 1},
			wantRes:  true,
		},
		{
			name:     "different size",
			expected: []int{1, 2, 3},
			actual:   []int{2, 1},
			wantRes:  false,
		},
		{
			name:     "with repeated element",
			expected: []int{1, 2, 2},
			actual:   []int{2, 1, 2},
			wantRes:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matcher := assessor.ElementsMatch(tt.expected)
			assert.Equal(t, tt.wantRes, matcher.Matches(tt.actual))
		})
	}
}

func TestOneOf(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name     string
		expected []int
		actual   int
		wantRes  bool
	}
	tests := []testCase{
		{
			name:     "no elements",
			expected: nil,
			actual:   0,
			wantRes:  false,
		},
		{
			name:     "empty slice",
			expected: []int{},
			actual:   0,
			wantRes:  false,
		},
		{
			name:     "found element",
			expected: []int{1},
			actual:   1,
			wantRes:  true,
		},
		{
			name:     "not found element",
			expected: []int{1},
			actual:   2,
			wantRes:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matcher := assessor.OneOf(tt.expected)
			assert.Equal(t, tt.wantRes, matcher.Matches(tt.actual))
		})
	}
}
