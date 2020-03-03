package pdftk

import (
	"sort"
	"strconv"
)

type InputFileMap map[string]string

func NewInputFileMap(inputFileNames ...string) InputFileMap {
	m := make(InputFileMap)
	for i, f := range inputFileNames {
		m[InputHandleNameFromInt(i)] = f
	}
	return m
}

func (m InputFileMap) parameterize() []string {
	// sort keys so that we can later properly handle dotted inputs
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// turn in key=value parameter list
	params := make([]string, len(m))
	for i, k := range keys {
		params[i] = k + "=" + m[k]
	}
	return params
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
