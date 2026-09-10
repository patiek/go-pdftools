package pdftk

import (
	"reflect"
	"testing"
)

func TestInputFileMap_parameterize(t *testing.T) {
	tests := []struct {
		name string
		m    InputFileMap
		want []string
	}{
		{
			name: "simple parameters are handled",
			m: InputFileMap{
				"A": "foo.txt",
				"B": "bar.txt",
				"C": "baz.txt",
			},
			want: []string{"A=foo.txt", "B=bar.txt", "C=baz.txt"},
		},
		{
			name: "parameters are sorted by length and then lexicographically",
			m: InputFileMap{
				"A":  "foo.txt",
				"AA": "aba.txt",
				"B":  "bar.txt",
				"C":  "baz.txt",
				"CC": "foobar.txt",
			},
			want: []string{"A=foo.txt", "B=bar.txt", "C=baz.txt", "AA=aba.txt", "CC=foobar.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.parameterize(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parameterize() = %v, want %v", got, tt.want)
			}
		})
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
