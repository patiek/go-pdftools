package fdf

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

const fdfHeader = "%FDF-1.2\r%\xe2\xe3\xcf\xd3\r\n1 0 obj\r<< \r/FDF << /Fields [ "
const fdfFooter = "] \r>> \r>> \rendobj\rtrailer\r<<\r/Root 1 0 R \r\r>>\r%%EOF\r\n"

// Write FDF content to writer using inputs.
func Write(w io.Writer, inputs Inputs) error {
	// sort keys so that we can later properly handle dotted inputs
	keys := make([]string, len(inputs))
	i := 0
	for k := range inputs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	if _, err := w.Write([]byte(fdfHeader)); err != nil {
		return err
	}

	if err := writeFields(w, inputs, keys, ""); err != nil {
		return err
	}

	if _, err := w.Write([]byte(fdfFooter)); err != nil {
		return err
	}

	return nil
}

func writeFields(w io.Writer, inputs Inputs, keys []string, parentPrefix string) error {
	for i := 0; i < len(keys); i++ {

		// open dictionary
		_, _ = fmt.Fprint(w, "<< ")

		// remove prefix from key
		k := keys[i][len(parentPrefix):]
		if dotIdx := strings.IndexByte(k, '.'); dotIdx > 0 {
			fieldName := k[0:dotIdx]
			prefix := parentPrefix + fieldName + "."

			// find all keys that follow that have this prefix
			j := i + 1
			for ; j < len(keys); j++ {
				if !strings.HasPrefix(keys[j], prefix) {
					break
				}
			}

			// recurse into child keys
			_, _ = fmt.Fprintf(w, "/T (%s) ", escapeStringInput(fieldName))
			_, _ = fmt.Fprint(w, "/Kids [ ")
			if err := writeFields(w, inputs, keys[i:j], prefix); err != nil {
				return err
			}
			_, _ = fmt.Fprint(w, "] ")

			// we have processed up to and including j-1 now, so advance i forward
			i = j - 1
		} else {

			var (
				hidden, readOnly bool
				input            interface{}
			)

			input = inputs[keys[i]]

			// field name
			_, _ = fmt.Fprintf(w, "/T (%s) ", escapeStringInput(k))

			// unwrap if we have Field
			if v, ok := input.(Field); ok {
				// unwrap value
				input = v.Value

				// flags
				hidden = v.Hidden
				readOnly = v.ReadOnly
			}

			// field value
			switch v := input.(type) {
			case OptionInput:
				_, _ = fmt.Fprintf(w, "/V /%s ", escapeOptionedInput(v))
			case fmt.Stringer:
				_, _ = fmt.Fprintf(w, "/V (%s) ", escapeStringInput(v.String()))
			case string:
				_, _ = fmt.Fprintf(w, "/V (%s) ", escapeStringInput(v))
			default:
				return fmt.Errorf("invalid type for input key %s: %T", keys[i], v)
			}

			// field flags
			if hidden {
				_, _ = fmt.Fprint(w, "/SetF 2 ")
			} else {
				_, _ = fmt.Fprint(w, "/ClrF 2 ")
			}

			if readOnly {
				_, _ = fmt.Fprint(w, "/SetFf 1 ")
			} else {
				_, _ = fmt.Fprint(w, "/ClrFf 1 ")
			}
		}

		_, _ = fmt.Fprint(w, ">> \r")
	}
	return nil
}

func escapeOptionedInput(s OptionInput) string {
	var b strings.Builder
	for i := range s {
		switch {
		case s[i] == '#', s[i] < 32, s[i] > 126:
			// convert to hex
			_, _ = fmt.Fprintf(&b, "#%02x", s[i])
		default:
			b.WriteByte(s[i])
		}
	}

	return b.String()
}

func escapeStringInput(s string) string {
	var b strings.Builder
	for i := range s {
		switch {
		case s[i] == '\\', s[i] == '(', s[i] == ')':
			// prepend with backslash
			b.WriteByte('\\')
			b.WriteByte(s[i])
		case s[i] < 32, s[i] > 126:
			// convert to octal
			_, _ = fmt.Fprintf(&b, "\\%03o", s[i])
		default:
			b.WriteByte(s[i])
		}
	}

	return b.String()
}
