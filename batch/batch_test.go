package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMapValue(t *testing.T) {
	type testCase struct {
		name     string
		input    map[string]int
		key      string
		err      error
		expected int
	}

	tests := []testCase{
		{
			name:     "returns value",
			input:    map[string]int{"wibble": 0, "foo": 1},
			key:      "foo",
			expected: 1,
		},
		{
			name:  "error on empty map",
			input: nil,
			key:   "bar",
			err:   ErrNilMap,
		},
		{
			name:  "error on missing key",
			input: map[string]int{"wibble": 0, "foo": 1},
			key:   "bar",
			err:   ErrKeyNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetMapValue(tc.input, tc.key)
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestGetMapSlice(t *testing.T) {
	type testCase struct {
		name     string
		input    map[string][]int
		key      string
		err      error
		expected []int
	}

	tests := []testCase{
		{
			name:     "returns value",
			input:    map[string][]int{"wibble": {0}, "foo": {1, 2}},
			key:      "foo",
			expected: []int{1, 2},
		},
		{
			name:     "returns value for 0 length slice",
			input:    map[string][]int{"wibble": {0}, "foo": {}},
			key:      "foo",
			expected: []int{},
		},
		{
			name:     "returns value for nil slice",
			input:    map[string][]int{"wibble": {0}, "foo": nil},
			key:      "foo",
			expected: nil,
		},
		{
			name:  "error on empty map",
			input: nil,
			key:   "bar",
			err:   ErrNilMap,
		},
		{
			name:  "error on missing key",
			input: map[string][]int{"wibble": {0}, "foo": {1, 2}},
			key:   "bar",
			err:   ErrKeyNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetMapSlice(tc.input, tc.key)
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestTransformMapValues(t *testing.T) {
	type testCase struct {
		name     string
		input    map[string]int
		parserFn func(int) string
		expected map[string]string
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			parserFn: func(v int) string {
				return "should-not-run"
			},
			expected: nil,
		},
		{
			name:  "empty map returns empty map",
			input: map[string]int{},
			parserFn: func(v int) string {
				return "should-not-run"
			},
			expected: map[string]string{},
		},
		{
			name: "single item parsed",
			input: map[string]int{
				"one":   1,
				"two":   2,
				"three": 3,
			},
			parserFn: func(v int) string {
				if v%2 == 0 {
					return "even"
				}
				return "odd"
			},
			expected: map[string]string{
				"one":   "odd",
				"two":   "even",
				"three": "odd",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TransformMapValues(tc.input, tc.parserFn)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestTransformMapSlices(t *testing.T) {
	type testCase struct {
		name     string
		input    map[string][]int
		parserFn func([]int) []string
		expected map[string][]string
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			parserFn: func(v []int) []string {
				return []string{"should-not-run"}
			},
			expected: nil,
		},
		{
			name:  "empty map returns empty map",
			input: map[string][]int{},
			parserFn: func(v []int) []string {
				return []string{"should-not-run"}
			},
			expected: map[string][]string{},
		},
		{
			name: "parses slices for each key",
			input: map[string][]int{
				"evens": {2, 4},
				"odds":  {1, 3, 5},
			},
			parserFn: func(v []int) []string {
				out := make([]string, len(v))
				for i := range v {
					if v[i]%2 == 0 {
						out[i] = "even"
						continue
					}
					out[i] = "odd"
				}
				return out
			},
			expected: map[string][]string{
				"evens": {"even", "even"},
				"odds":  {"odd", "odd", "odd"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TransformMapSlices(tc.input, tc.parserFn)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestIndexBy(t *testing.T) {
	type testCase struct {
		name     string
		input    []int
		keyFn    func(item *int) string
		err      error
		expected map[string]int
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			keyFn: func(item *int) string {
				return "should-not-run"
			},
			expected: nil,
		},
		{
			name:  "empty slice returns empty map",
			input: []int{},
			keyFn: func(item *int) string {
				return "should-not-run"
			},
			expected: map[string]int{},
		},
		{
			name:  "maps items by key",
			input: []int{1, 2, 3},
			keyFn: func(item *int) string {
				return string(rune('a' + *item - 1))
			},
			expected: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			name:  "errors on duplicate key",
			input: []int{1, 3},
			keyFn: func(item *int) string {
				if *item%2 == 0 {
					return "even"
				}
				return "odd"
			},
			err: ErrDuplicateKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := IndexBy(tc.input, tc.keyFn)
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestIndexTo(t *testing.T) {
	type testCase struct {
		name     string
		input    []int
		keyFn    func(item *int) string
		parserFn func(item *int) bool
		err      error
		expected map[string]bool
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			keyFn: func(item *int) string {
				return "should-not-run"
			},
			parserFn: func(item *int) bool {
				return false
			},
			expected: nil,
		},
		{
			name:  "empty slice returns empty map",
			input: []int{},
			keyFn: func(item *int) string {
				return "should-not-run"
			},
			parserFn: func(item *int) bool {
				return false
			},
			expected: map[string]bool{},
		},
		{
			name:  "maps items by key with parsed output",
			input: []int{1, 2, 3},
			keyFn: func(item *int) string {
				return string(rune('a' + *item - 1))
			},
			parserFn: func(item *int) bool {
				return *item%2 == 0
			},
			expected: map[string]bool{
				"a": false,
				"b": true,
				"c": false,
			},
		},
		{
			name:  "errors on duplicate key",
			input: []int{1, 3},
			keyFn: func(item *int) string {
				return "odd"
			},
			parserFn: func(item *int) bool {
				return *item > 0
			},
			err: ErrDuplicateKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := IndexTo(tc.input, tc.keyFn, tc.parserFn)
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestGroupBy(t *testing.T) {
	type testCase struct {
		name     string
		input    []int
		keyFn    func(item *int) bool
		err      error
		expected map[bool][]int
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			keyFn: func(item *int) bool {
				return false
			},
			expected: nil,
		},
		{
			name:  "empty slice returns empty map",
			input: []int{},
			keyFn: func(item *int) bool {
				return false
			},
			expected: map[bool][]int{},
		},
		{
			name:  "groups items by key",
			input: []int{1, 2, 3, 4},
			keyFn: func(item *int) bool {
				return *item%2 == 0
			},
			expected: map[bool][]int{
				false: {1, 3},
				true:  {2, 4},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GroupBy(tc.input, tc.keyFn)
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestGroupTo(t *testing.T) {
	type testCase struct {
		name     string
		input    []int
		keyFn    func(item *int) bool
		parserFn func(item *int) string
		err      error
		expected map[bool][]string
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			keyFn: func(item *int) bool {
				return false
			},
			parserFn: func(item *int) string {
				return "should-not-run"
			},
			expected: nil,
		},
		{
			name:  "empty slice returns empty map",
			input: []int{},
			keyFn: func(item *int) bool {
				return false
			},
			parserFn: func(item *int) string {
				return "should-not-run"
			},
			expected: map[bool][]string{},
		},
		{
			name:  "groups parsed items by key",
			input: []int{1, 2, 3, 4},
			keyFn: func(item *int) bool {
				return *item%2 == 0
			},
			parserFn: func(item *int) string {
				if *item > 2 {
					return "big"
				}
				return "small"
			},
			expected: map[bool][]string{
				false: {"small", "big"},
				true:  {"small", "big"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GroupTo(tc.input, tc.keyFn, tc.parserFn)
			require.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestTransformSlice(t *testing.T) {
	type testCase struct {
		name     string
		input    []int
		parserFn func(item int) bool
		expected []bool
	}

	tests := []testCase{
		{
			name:  "nil input returns nil",
			input: nil,
			parserFn: func(item int) bool {
				return false
			},
			expected: nil,
		},
		{
			name:  "empty slice returns empty slice",
			input: []int{},
			parserFn: func(item int) bool {
				return false
			},
			expected: []bool{},
		},
		{
			name:  "parses each item",
			input: []int{1, 2, 3, 4},
			parserFn: func(item int) bool {
				return item%2 == 0
			},
			expected: []bool{false, true, false, true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TransformSlice(tc.input, tc.parserFn)
			assert.Equal(t, tc.expected, got)
		})
	}
}
