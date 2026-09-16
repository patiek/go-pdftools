package pdftk

import (
	"cmp"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// Input PDFs for Cat keyed by handle name (uppercase A-Z), which PageRange
// refers to. Readers are read at most once and never closed.
type InputMap map[string]io.Reader

// Assign handles A, B, C, ... to inputs in order.
func NewInputMap(inputs ...io.Reader) InputMap {
	m := make(InputMap, len(inputs))
	for i, r := range inputs {
		m[InputHandleNameFromInt(i)] = r
	}
	return m
}

// Handles sorted by length and then lexicographically, so a map built by
// NewInputMap keeps its argument order.
func (m InputMap) handles() []string {
	handles := slices.Collect(maps.Keys(m))
	slices.SortFunc(handles, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(a), len(b)), strings.Compare(a, b))
	})
	return handles
}

// Get valid input handle name (A-Z characters only) by converting num to A-Z base 26
func InputHandleNameFromInt(num int) string {
	// convert to base 26
	s := strconv.FormatInt(int64(num), 26)

	// transform the 0-9,a-p base 26 into uppercase A-Z characters
	r := make([]byte, len(s))
	for i := range s {
		switch {
		case s[i] >= '0' && s[i] <= '9':
			// 0 --> A, 1 --> B, ...
			// e.g. 0 has ascii 48, so 48 + 17 = 65 which is 'A'
			r[i] = s[i] + 17
		case s[i] >= 'a' && s[i] <= 'p':
			// a --> K, b --> L, ...
			// e.g. a has ascii 97, so 97 - 22 = 75 which is 'K'
			r[i] = s[i] - 22
		}
	}
	return string(r)
}

// Get int representation of input handle name by converting A-Z base 26 to base 10
func InputHandleNameToInt(s string) (int, error) {
	// transform the A-Z characters into 0-9,a-p base 26
	// e.g. CZZ --> 2pp
	r := make([]byte, len(s))
	for i := range s {
		switch {
		case s[i] >= 'A' && s[i] <= 'J':
			// A --> 0, B --> 1, ..., J -> 9
			// e.g. A has ascii 65, so 65 - 17 = 48 which is '0'
			r[i] = s[i] - 17
		case s[i] >= 'K' && s[i] <= 'Z':
			// K --> a, L --> b, ...
			// e.g. K has ascii 75, so 75 + 22 = 97 which is 'a'
			r[i] = s[i] + 22
		}
	}

	num, err := strconv.ParseInt(string(r), 26, 0)
	return int(num), err
}
