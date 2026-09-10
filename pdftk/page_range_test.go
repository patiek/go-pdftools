package pdftk

import "testing"

func TestPageRange_String(t *testing.T) {
	tests := []struct {
		name string
		pr   PageRange
		want string
	}{
		{
			name: "whole document",
			pr:   PageRange{FileHandleName: "A"},
			want: "A",
		},
		{
			name: "single page",
			pr:   PageRange{FileHandleName: "A", BeginPage: 3, EndPage: 3},
			want: "A3",
		},
		{
			name: "page range",
			pr:   PageRange{FileHandleName: "B", BeginPage: 2, EndPage: 5},
			want: "B2-5",
		},
		{
			name: "begin page to end of document",
			pr:   PageRange{FileHandleName: "A", BeginPage: 4},
			want: "A4-end",
		},
		{
			name: "qualifier and rotation",
			pr:   PageRange{FileHandleName: "A", BeginPage: 1, EndPage: 10, Qualifier: Even, Rotation: Left},
			want: "A1-10evenleft",
		},
		{
			name: "rotation only",
			pr:   PageRange{FileHandleName: "AB", Rotation: South},
			want: "ABsouth",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
