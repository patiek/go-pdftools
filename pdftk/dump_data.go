package pdftk

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const numberOfPagesKey = "NumberOfPages"

// Find the NumberOfPages line in dump_data output and parse its value.
// Info and bookmark strings are printed unescaped, so a value containing a
// newline could forge a NumberOfPages line; refuse output with more than one.
func parseNumberOfPages(s string) (int, error) {
	n, found := 0, false
	for line := range strings.Lines(s) {
		v, ok := strings.CutPrefix(line, numberOfPagesKey+":")
		if !ok {
			continue
		}
		if found {
			return 0, errors.New("pdftk error: multiple " + numberOfPagesKey + " lines in dump_data output")
		}
		v = strings.TrimSpace(v)
		var err error
		if n, err = strconv.Atoi(v); err != nil {
			return 0, fmt.Errorf("pdftk error: invalid %s value %q", numberOfPagesKey, v)
		}
		found = true
	}
	if !found {
		return 0, errors.New("pdftk error: " + numberOfPagesKey + " not found in dump_data output")
	}
	return n, nil
}
