package pdftk

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestInputMap_handles(t *testing.T) {
	r := strings.NewReader("")
	tests := []struct {
		name string
		m    InputMap
		want []string
	}{
		{
			name: "single letters",
			m:    InputMap{"A": r, "B": r, "C": r},
			want: []string{"A", "B", "C"},
		},
		{
			name: "sorted by length and then lexicographically",
			m:    InputMap{"A": r, "AA": r, "B": r, "C": r, "CC": r},
			want: []string{"A", "B", "C", "AA", "CC"},
		},
		{
			name: "empty",
			m:    InputMap{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.handles(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handles() = %v, want %v", got, tt.want)
			}
		})
	}
}

// NewInputMap assigns handles in argument order, past Z.
func TestNewInputMap(t *testing.T) {
	readers := make([]io.Reader, 28)
	for i := range readers {
		readers[i] = strings.NewReader(strconv.Itoa(i))
	}
	m := NewInputMap(readers...)
	handles := m.handles()
	if len(handles) != len(readers) {
		t.Fatalf("NewInputMap() has %d handles, want %d", len(handles), len(readers))
	}
	for i, handle := range handles {
		if m[handle] != readers[i] {
			t.Errorf("handle %s does not hold input %d", handle, i)
		}
	}
	if got := handles[26:]; !reflect.DeepEqual(got, []string{"BA", "BB"}) {
		t.Errorf("handles after Z = %v, want [BA BB]", got)
	}
}

func TestInputHandleName(t *testing.T) {
	tests := []struct {
		num  int
		name string
	}{
		{0, "A"}, {1, "B"}, {9, "J"}, {10, "K"}, {25, "Z"}, {26, "BA"}, {675, "ZZ"}, {676, "BAA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InputHandleNameFromInt(tt.num); got != tt.name {
				t.Errorf("InputHandleNameFromInt(%d) = %q, want %q", tt.num, got, tt.name)
			}
			if got, err := InputHandleNameToInt(tt.name); err != nil || got != tt.num {
				t.Errorf("InputHandleNameToInt(%q) = %d, %v, want %d", tt.name, got, err, tt.num)
			}
		})
	}
}
