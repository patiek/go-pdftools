package pdftk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/patiek/go-pdftools/fdf"
)

// Tests in this file run the real pdftk and are skipped when it is not installed.

func requirePDFtk(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(cmdPDFtk); err != nil {
		t.Skip(err)
	}
}

// Minimal PDF made of objects (numbered from 1, object 1 is the catalog).
func pdfBytes(objects ...string) []byte {
	var b bytes.Buffer
	offsets := make([]int, len(objects))
	b.WriteString("%PDF-1.4\n")
	for i, obj := range objects {
		offsets[i] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return b.Bytes()
}

func writePDF(t *testing.T, name string, objects ...string) string {
	t.Helper()
	return writeFile(t, name, string(pdfBytes(objects...)))
}

func blankPDFObjects(pages int) []string {
	kids := make([]string, pages)
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", ""}
	for i := range pages {
		kids[i] = fmt.Sprintf("%d 0 R", i+3)
		objects = append(objects, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>")
	}
	objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [ %s ] /Count %d >>", strings.Join(kids, " "), pages)
	return objects
}

func blankPDF(pages int) io.Reader {
	return bytes.NewReader(pdfBytes(blankPDFObjects(pages)...))
}

func openBlankPDF(t *testing.T, pages int) *os.File {
	t.Helper()
	return openFile(t, writePDF(t, fmt.Sprintf("blank%d.pdf", pages), blankPDFObjects(pages)...))
}

// Single page PDF with one text field named "name".
func formPDFObjects() []string {
	return fieldPDFObjects("/FT /Tx /DA (/Helv 12 Tf 0 g)")
}

// Single page PDF with one field named "name" made of entries.
// Object 6 is an empty stream for appearances.
func fieldPDFObjects(entries string) []string {
	return []string{
		"<< /Type /Catalog /Pages 2 0 R /AcroForm << /Fields [ 4 0 R ] /DA (/Helv 0 Tf 0 g) /DR << /Font << /Helv 5 0 R >> >> >> >>",
		"<< /Type /Pages /Kids [ 3 0 R ] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Annots [ 4 0 R ] >>",
		"<< /Type /Annot /Subtype /Widget /T (name) /Rect [50 700 300 720] /F 4 /P 3 0 R " + entries + " >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Length 0 >>\nstream\n\nendstream",
	}
}

func fdfData(t *testing.T, value any) string {
	t.Helper()
	var b bytes.Buffer
	if err := fdf.Write(&b, fdf.Inputs{"name": value}); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func xfdfData(value string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<xfdf xmlns="http://ns.adobe.com/xfdf/" xml:space="preserve"><fields><field name="name"><value>` +
		value + `</value></field></fields></xfdf>` + "\n"
}

// Write the output of fn to a file and return its path.
func writeOutput(t *testing.T, fn func(out io.Writer) error) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out.pdf")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := fn(f); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func outputPages(t *testing.T, fn func(out io.Writer) error) int {
	t.Helper()
	n, err := NumberOfPages(t.Context(), openFile(t, writeOutput(t, fn)))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// Values of key for the form field "name" in the PDF at path, via dump_data_fields_utf8.
func fieldData(t *testing.T, path, key string) []string {
	t.Helper()
	var fields bytes.Buffer
	cmd, err := newCommand(t.Context(), &fields, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.run(path, "dump_data_fields_utf8", "output", "-"); err != nil {
		t.Fatal(err)
	}
	var values []string
	for line := range strings.Lines(fields.String()) {
		if v, ok := strings.CutPrefix(line, key+": "); ok {
			values = append(values, strings.TrimSpace(v))
		}
	}
	return values
}

func fieldValue(t *testing.T, path string) string {
	t.Helper()
	if values := fieldData(t, path, "FieldValue"); len(values) > 0 {
		return values[0]
	}
	return ""
}

func TestNumberOfPages_pdftk(t *testing.T) {
	requirePDFtk(t)
	for _, pages := range []int{1, 3, 12} {
		inputs := map[string]io.Reader{
			"file":   openBlankPDF(t, pages),
			"reader": blankPDF(pages),
		}
		for name, in := range inputs {
			t.Run(fmt.Sprintf("%d pages from %s", pages, name), func(t *testing.T) {
				got, err := NumberOfPages(t.Context(), in)
				if err != nil {
					t.Fatalf("NumberOfPages() error = %v", err)
				}
				if got != pages {
					t.Errorf("NumberOfPages() = %d, want %d", got, pages)
				}
			})
		}
	}

	t.Run("invalid input", func(t *testing.T) {
		for name, in := range map[string]io.Reader{"garbage": strings.NewReader("not a pdf"), "empty": bytes.NewReader(nil)} {
			_, err := NumberOfPages(t.Context(), in)
			if _, ok := errors.AsType[*exec.ExitError](err); !ok {
				t.Errorf("NumberOfPages(%s) error = %v, want *exec.ExitError", name, err)
			}
		}
	})

	// the file opened before the change of directory is the one counted
	t.Run("relative name after changing directory", func(t *testing.T) {
		t.Chdir(filepath.Dir(writePDF(t, "in.pdf", blankPDFObjects(3)...)))
		in := openFile(t, "in.pdf")
		t.Chdir(filepath.Dir(writePDF(t, "in.pdf", blankPDFObjects(1)...)))
		got, err := NumberOfPages(t.Context(), in)
		if err != nil {
			t.Fatalf("NumberOfPages() error = %v", err)
		}
		if got != 3 {
			t.Errorf("NumberOfPages() = %d, want 3", got)
		}
	})
}

func TestCat_pdftk(t *testing.T) {
	requirePDFtk(t)
	tests := []struct {
		name       string
		inputs     func(t *testing.T) InputMap
		pageRanges []PageRange
		want       int
	}{
		{
			name: "files",
			inputs: func(t *testing.T) InputMap {
				return NewInputMap(openBlankPDF(t, 3), openBlankPDF(t, 5))
			},
			want: 8,
		},
		{
			name: "readers",
			inputs: func(*testing.T) InputMap {
				return NewInputMap(blankPDF(3), blankPDF(5))
			},
			want: 8,
		},
		{
			name: "mixed inputs with page ranges",
			inputs: func(t *testing.T) InputMap {
				return NewInputMap(openBlankPDF(t, 3), blankPDF(5), blankPDF(2))
			},
			pageRanges: []PageRange{
				{FileHandleName: "B", BeginPage: 2, EndPage: 4},
				{FileHandleName: "A", Qualifier: Odd},
				{FileHandleName: "C", Rotation: East},
				{FileHandleName: "A", BeginPage: 2, EndPage: 2, Rotation: East},
			},
			want: 8,
		},
		{
			name: "same file under two handles",
			inputs: func(t *testing.T) InputMap {
				f := openBlankPDF(t, 3)
				return NewInputMap(f, f)
			},
			want: 6,
		},
		{
			name: "same reader under two handles",
			inputs: func(*testing.T) InputMap {
				r := blankPDF(3)
				return NewInputMap(r, r)
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			got := outputPages(t, func(out io.Writer) error {
				return Cat(t.Context(), out, tt.inputs(t), tt.pageRanges, OptionTempDir(dir))
			})
			if got != tt.want {
				t.Errorf("Cat() produced %d pages, want %d", got, tt.want)
			}
			requireEmptyDir(t, dir)
		})
	}
}

func TestFillForm_pdftk(t *testing.T) {
	requirePDFtk(t)
	forms := map[string]func(t *testing.T) io.Reader{
		"form file":   func(t *testing.T) io.Reader { return openFile(t, writePDF(t, "form.pdf", formPDFObjects()...)) },
		"form reader": func(*testing.T) io.Reader { return bytes.NewReader(pdfBytes(formPDFObjects()...)) },
	}
	data := map[string]string{
		"FDF":  fdfData(t, "hello fdf"),
		"XFDF": xfdfData("hello xfdf"),
	}
	for formName, form := range forms {
		for dataName, content := range data {
			t.Run(formName+" "+dataName, func(t *testing.T) {
				want := strings.TrimPrefix(strings.ToLower("hello "+dataName), "hello ")
				dir := t.TempDir()
				out := writeOutput(t, func(out io.Writer) error {
					return FillForm(t.Context(), out, form(t), strings.NewReader(content), OptionTempDir(dir))
				})
				requireEmptyDir(t, dir)
				if got := fieldValue(t, out); got != "hello "+want {
					t.Errorf("filled form field = %q, want %q", got, "hello "+want)
				}

				// flattened output is still a valid single page PDF
				got := outputPages(t, func(out io.Writer) error {
					return FillForm(t.Context(), out, form(t), strings.NewReader(content), OptionFlatten())
				})
				if got != 1 {
					t.Errorf("FillForm() produced %d pages, want 1", got)
				}
			})
		}
	}
}

// FDF text values must read back from the filled form unchanged.
func TestFillFormText_pdftk(t *testing.T) {
	requirePDFtk(t)
	values := []string{"1FT(J)W\\35", "José Muñoz", "José (Muñoz) \\ 車", "😀 Ünïcödé"}
	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			out := writeOutput(t, func(out io.Writer) error {
				return FillForm(t.Context(), out, bytes.NewReader(pdfBytes(formPDFObjects()...)), strings.NewReader(fdfData(t, value)))
			})
			if got := fieldValue(t, out); got != value {
				t.Errorf("filled form field = %q, want %q", got, value)
			}
		})
	}
}

// An option listed by pdftk for a field must select it when filled as an OptionInput.
func TestFillFormOption_pdftk(t *testing.T) {
	requirePDFtk(t)
	checkbox := "/FT /Btn /V /Off /AS /Off /AP << /N << /Off 6 0 R /%s 6 0 R >> >>"
	fields := []struct {
		entries string
		typed   string // also selects the one listed option
	}{
		{entries: fmt.Sprintf(checkbox, "Yes")},
		{entries: fmt.Sprintf(checkbox, "United#20States")},
		{entries: fmt.Sprintf(checkbox, "a#2fb#28c#29#3cd#3e#5be#5d#7bf#7d#25g#23")},
		{entries: fmt.Sprintf(checkbox, "S#ed")},
		// UTF-8 name, listed by pdftk as Latin-1
		{entries: fmt.Sprintf(checkbox, "S#c3#ad#e8#bb#8a"), typed: "Sí車"},
		{entries: "/FT /Ch /Ff 131072 /DA (/Helv 12 Tf 0 g) /Opt [ (Espa\\361a) (United States) ]"},
	}
	for _, field := range fields {
		form := writePDF(t, "form.pdf", fieldPDFObjects(field.entries)...)
		options := slices.DeleteFunc(fieldData(t, form, "FieldStateOption"), func(o string) bool { return o == "Off" })
		if len(options) == 0 {
			t.Fatalf("no options listed for %s", field.entries)
		}
		inputs := make(map[string]string)
		for _, option := range options {
			inputs[option] = option
		}
		if field.typed != "" {
			inputs[field.typed] = options[0]
		}
		for input, want := range inputs {
			t.Run(input, func(t *testing.T) {
				out := writeOutput(t, func(out io.Writer) error {
					return FillForm(t.Context(), out, openFile(t, form), strings.NewReader(fdfData(t, fdf.OptionInput(input))))
				})
				if got := fieldValue(t, out); got != want {
					t.Errorf("filled form field = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestOverlays_pdftk(t *testing.T) {
	requirePDFtk(t)
	type overlayFunc func(context.Context, io.Writer, io.Reader, io.Reader, ...Option) error
	inputs := []struct {
		name        string
		in, overlay func(t *testing.T) io.Reader
	}{
		{"files", func(t *testing.T) io.Reader { return openBlankPDF(t, 3) }, func(t *testing.T) io.Reader { return openBlankPDF(t, 2) }},
		{"readers", func(*testing.T) io.Reader { return blankPDF(3) }, func(*testing.T) io.Reader { return blankPDF(2) }},
		{"file and reader", func(t *testing.T) io.Reader { return openBlankPDF(t, 3) }, func(*testing.T) io.Reader { return blankPDF(2) }},
		{"reader and file", func(*testing.T) io.Reader { return blankPDF(3) }, func(t *testing.T) io.Reader { return openBlankPDF(t, 2) }},
	}
	tests := []struct {
		name string
		fn   overlayFunc
		in   int // index into inputs
	}{
		{"Background", Background, 0},
		{"MultiBackground", MultiBackground, 0},
		{"Stamp", Stamp, 0},
		{"MultiStamp", MultiStamp, 0},
		{"Stamp", Stamp, 1},
		{"Stamp", Stamp, 2},
		{"Stamp", Stamp, 3},
	}
	for _, tt := range tests {
		in := inputs[tt.in]
		t.Run(tt.name+" "+in.name, func(t *testing.T) {
			dir := t.TempDir()
			got := outputPages(t, func(out io.Writer) error {
				return tt.fn(t.Context(), out, in.in(t), in.overlay(t), OptionTempDir(dir))
			})
			if got != 3 {
				t.Errorf("%s() produced %d pages, want 3", tt.name, got)
			}
			requireEmptyDir(t, dir)
		})
	}

	t.Run("temp files are removed when pdftk fails", func(t *testing.T) {
		dir := t.TempDir()
		err := Stamp(t.Context(), io.Discard, blankPDF(3), strings.NewReader("not a pdf"), OptionTempDir(dir))
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			t.Errorf("Stamp() error = %v, want *exec.ExitError", err)
		}
		requireEmptyDir(t, dir)
	})
}

// Files whose bare name pdftk would take for a keyword, handle assignment or
// option must be usable in every position.
func TestInputNames_pdftk(t *testing.T) {
	requirePDFtk(t)
	names := []string{"cat", "Output", "verbose", "PROMPT", "-x.pdf", "X=blank.pdf", "we ird=dir/s t=amp.pdf", "dïr 日本/in.pdf"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			// pdftk only misparses bare names, so open them from the working directory
			dir := t.TempDir()
			writeFileIn(t, dir, name, string(pdfBytes(blankPDFObjects(3)...)))
			t.Chdir(dir)
			got, err := NumberOfPages(t.Context(), openFile(t, name))
			if err != nil {
				t.Fatalf("NumberOfPages() error = %v", err)
			}
			if got != 3 {
				t.Errorf("NumberOfPages() = %d, want 3", got)
			}

			if got := outputPages(t, func(out io.Writer) error {
				return Stamp(t.Context(), out, openBlankPDF(t, 2), openFile(t, name))
			}); got != 2 {
				t.Errorf("Stamp() produced %d pages, want 2", got)
			}

			form := openFile(t, writePDF(t, "form.pdf", formPDFObjects()...))
			fdfDir := t.TempDir()
			writeFileIn(t, fdfDir, name, fdfData(t, "hello"))
			t.Chdir(fdfDir)
			data := openFile(t, name)
			out := writeOutput(t, func(out io.Writer) error {
				return FillForm(t.Context(), out, form, data)
			})
			if got := fieldValue(t, out); got != "hello" {
				t.Errorf("filled form field = %q, want hello", got)
			}
		})
	}
}
