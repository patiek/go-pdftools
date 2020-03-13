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
			name:"simple parameters are handled",
			m: InputFileMap{
				"A": "foo.txt",
				"B": "bar.txt",
				"C": "baz.txt",
			},
			want: []string{"A=foo.txt", "B=bar.txt", "C=baz.txt"},
		},
		{
			name:"parameters are sorted by length and then lexicographically",
			m: InputFileMap{
				"A": "foo.txt",
				"AA": "aba.txt",
				"B": "bar.txt",
				"C": "baz.txt",
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
