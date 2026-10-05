package fdf

import (
	"bytes"
	"testing"
)

type stringer string

func (s stringer) String() string { return string(s) }

func TestWriteFDF(t *testing.T) {
	type args struct {
		inputs Inputs
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "testing basic write",
			args: args{inputs: Inputs{
				"field 1":     "value 1",
				"field 2":     "value 2",
				"foo.bar.baz": "bazval",
				"foo.bar.buz": "buzval",
			}},
			wantW:   "%FDF-1.2\r%\xe2\xe3\xcf\xd3\r\n1 0 obj\r<< \r/FDF << /Fields [ << /T (field 1) /V (value 1) /ClrF 2 /ClrFf 1 >> \r<< /T (field 2) /V (value 2) /ClrF 2 /ClrFf 1 >> \r<< /T (foo) /Kids [ << /T (bar) /Kids [ << /T (baz) /V (bazval) /ClrF 2 /ClrFf 1 >> \r<< /T (buz) /V (buzval) /ClrF 2 /ClrFf 1 >> \r] >> \r] >> \r] \r>> \r>> \rendobj\rtrailer\r<<\r/Root 1 0 R \r\r>>\r%%EOF\r\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := Write(w, tt.args.inputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Write() \n\tgotW = %q,\n\twant = %q", gotW, tt.wantW)
			}
		})
	}
}

func Test_escapeOptionedInput(t *testing.T) {
	type args struct {
		v OptionInput
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "optioned input properly escaped",
			args: args{v: OptionInput("foo #bar ##baz \r buz \n end\177")},
			want: "foo #23bar #23#23baz #0d buz #0a end#7f",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeOptionedInput(tt.args.v); got != tt.want {
				t.Errorf("escapeOptionedInput()\nresult = %q,\n\twant = %q", got, tt.want)
			}
		})
	}
}

func Test_escapeStringInput(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "strings properly escaped",
			args: args{s: "foo (bar) \\ baz \nbuz\r \\)( end\177"},
			want: "foo \\(bar\\) \\\\ baz \\012buz\\015 \\\\\\)\\( end\\177",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeStringInput(tt.args.s); got != tt.want {
				t.Errorf("escapeStringInput()\nresult = %q,\n\twant = %q", got, tt.want)
			}
		})
	}
}

func Test_escapeTextValue(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "ascii same as escapeStringInput",
			s:    "foo (bar) \\ baz \nbuz\r \\)( end\177",
			want: "foo \\(bar\\) \\\\ baz \\012buz\\015 \\\\\\)\\( end\\177",
		},
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "non-ascii as utf-16be with bom",
			s:    "José",
			want: `\376\377\000\112\000\157\000\163\000\351`,
		},
		{
			name: "delimiters in non-ascii escaped",
			s:    "é()\\",
			want: `\376\377\000\351\000\050\000\051\000\134`,
		},
		{
			name: "astral as surrogate pair",
			s:    "😀",
			want: `\376\377\330\075\336\000`,
		},
		{
			name: "invalid utf-8 replaced",
			s:    "a\xff",
			want: `\376\377\000\141\377\375`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeTextValue(tt.s); got != tt.want {
				t.Errorf("escapeTextValue()\nresult = %q,\n\twant = %q", got, tt.want)
			}
		})
	}
}

func Test_writeFields(t *testing.T) {
	type args struct {
		inputs       Inputs
		keys         []string
		parentPrefix string
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "basic test",
			args: args{
				inputs: Inputs{
					"foo": "foo val",
					"bar": "bar val",
				},
				keys:         []string{"foo", "bar"},
				parentPrefix: "",
			},
			wantW:   "<< /T (foo) /V (foo val) /ClrF 2 /ClrFf 1 >> \r<< /T (bar) /V (bar val) /ClrF 2 /ClrFf 1 >> \r",
			wantErr: false,
		},
		{
			name: "non-ascii text values",
			args: args{
				inputs: Inputs{
					"str":    "é",
					"strer":  stringer("ñ"),
					"field":  Field{ReadOnly: true, Value: "ü"},
					"option": OptionInput("Ja"),
				},
				keys:         []string{"str", "strer", "field", "option"},
				parentPrefix: "",
			},
			wantW: `<< /T (str) /V (\376\377\000\351) /ClrF 2 /ClrFf 1 >> ` + "\r" +
				`<< /T (strer) /V (\376\377\000\361) /ClrF 2 /ClrFf 1 >> ` + "\r" +
				`<< /T (field) /V (\376\377\000\374) /ClrF 2 /SetFf 1 >> ` + "\r" +
				`<< /T (option) /V /Ja /ClrF 2 /ClrFf 1 >> ` + "\r",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := writeFields(w, tt.args.inputs, tt.args.keys, tt.args.parentPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("writeFields() \n\tgotW = %q,\n\twant = %q", gotW, tt.wantW)
			}
		})
	}
}
