package pdftk

import (
	"strings"
	"testing"
)

func Test_parseNumberOfPages(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{
			name: "full dump_data output",
			in: "InfoBegin\nInfoKey: Creator\nInfoValue: test\nPdfID0: 1a2b\nPdfID1: 3c4d\n" +
				"NumberOfPages: 12\nPageMediaBegin\nPageMediaNumber: 1\nPageMediaRotation: 0\n",
			want: 12,
		},
		{
			name: "only number of pages",
			in:   "NumberOfPages: 1\n",
			want: 1,
		},
		{
			name: "windows line endings",
			in:   "NumberOfPages: 3\r\nPageMediaBegin\r\n",
			want: 3,
		},
		{
			name: "long info value",
			in:   "InfoBegin\nInfoKey: Title\nInfoValue: " + strings.Repeat("x", 70000) + "\nNumberOfPages: 3\n",
			want: 3,
		},
		{
			name:    "forged line in info value",
			in:      "InfoBegin\nInfoKey: Title\nInfoValue: abc\nNumberOfPages: 999\ndef\nNumberOfPages: 3\n",
			wantErr: true,
		},
		{
			name:    "missing",
			in:      "InfoBegin\nInfoKey: Creator\nInfoValue: test\n",
			wantErr: true,
		},
		{
			name:    "invalid value",
			in:      "NumberOfPages: many\n",
			wantErr: true,
		},
		{
			name:    "empty",
			in:      "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNumberOfPages(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNumberOfPages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseNumberOfPages() = %d, want %d", got, tt.want)
			}
		})
	}
}
