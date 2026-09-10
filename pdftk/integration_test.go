package pdftk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

// Write a minimal PDF made of objects (numbered from 1, object 1 is the catalog).
func writePDF(t *testing.T, name string, objects ...string) string {
	t.Helper()
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

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, b.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
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

func writeBlankPDF(t *testing.T, pages int) string {
	t.Helper()
	return writePDF(t, fmt.Sprintf("blank%d.pdf", pages), blankPDFObjects(pages)...)
}

// Single page PDF with one text field named "name".
func writeFormPDF(t *testing.T) string {
	t.Helper()
	return writePDF(t, "form.pdf",
		"<< /Type /Catalog /Pages 2 0 R /AcroForm << /Fields [ 4 0 R ] /DA (/Helv 0 Tf 0 g) /DR << /Font << /Helv 5 0 R >> >> >> >>",
		"<< /Type /Pages /Kids [ 3 0 R ] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Annots [ 4 0 R ] >>",
		"<< /Type /Annot /Subtype /Widget /FT /Tx /T (name) /Rect [50 700 300 720] /F 4 /P 3 0 R /DA (/Helv 12 Tf 0 g) >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	)
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
	n, err := NumberOfPages(t.Context(), writeOutput(t, fn))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestNumberOfPages_pdftk(t *testing.T) {
	requirePDFtk(t)
	for _, pages := range []int{1, 3, 12} {
		got, err := NumberOfPages(t.Context(), writeBlankPDF(t, pages))
		if err != nil {
			t.Fatalf("NumberOfPages() error = %v", err)
		}
		if got != pages {
			t.Errorf("NumberOfPages() = %d, want %d", got, pages)
		}
	}

	missing := filepath.Join(t.TempDir(), "missing.pdf")
	if _, err := NumberOfPages(t.Context(), missing); err == nil || !strings.Contains(err.Error(), "missing.pdf") {
		t.Errorf("NumberOfPages() error = %v, want error naming missing.pdf", err)
	}
}

func TestCat_pdftk(t *testing.T) {
	requirePDFtk(t)
	fileNames := NewInputFileMap(writeBlankPDF(t, 3), writeBlankPDF(t, 5))
	tests := []struct {
		name       string
		pageRanges []PageRange
		want       int
	}{
		{
			name: "all pages",
			want: 8,
		},
		{
			name: "page ranges",
			pageRanges: []PageRange{
				{FileHandleName: "B", BeginPage: 2, EndPage: 4},
				{FileHandleName: "A", Qualifier: Odd},
				{FileHandleName: "A", BeginPage: 2, EndPage: 2, Rotation: East},
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outputPages(t, func(out io.Writer) error {
				return Cat(t.Context(), out, fileNames, tt.pageRanges)
			})
			if got != tt.want {
				t.Errorf("Cat() produced %d pages, want %d", got, tt.want)
			}
		})
	}
}

func TestFillForm_pdftk(t *testing.T) {
	requirePDFtk(t)
	var data bytes.Buffer
	if err := fdf.Write(&data, fdf.Inputs{"name": "hello world"}); err != nil {
		t.Fatal(err)
	}
	formFileName := writeFormPDF(t)

	out := writeOutput(t, func(out io.Writer) error {
		return FillForm(t.Context(), out, formFileName, &data)
	})
	var fields bytes.Buffer
	if err := createCmd(t.Context(), cmdPDFtk, &fields, nil, out, "dump_data_fields", "output", "-").run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fields.String(), "FieldValue: hello world") {
		t.Errorf("filled form fields = %q, want FieldValue: hello world", fields.String())
	}

	// flattened output is still a valid single page PDF
	data.Reset()
	if err := fdf.Write(&data, fdf.Inputs{"name": "hello world"}); err != nil {
		t.Fatal(err)
	}
	got := outputPages(t, func(out io.Writer) error {
		return FillForm(t.Context(), out, formFileName, &data, OptionFlatten())
	})
	if got != 1 {
		t.Errorf("FillForm() produced %d pages, want 1", got)
	}
}

func TestOverlays_pdftk(t *testing.T) {
	requirePDFtk(t)
	tests := []struct {
		name string
		fn   func(context.Context, io.Writer, string, io.Reader, ...Option) error
	}{
		{"Background", Background},
		{"MultiBackground", MultiBackground},
		{"Stamp", Stamp},
		{"MultiStamp", MultiStamp},
	}
	inputFileName := writeBlankPDF(t, 3)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlay, err := os.Open(writeBlankPDF(t, 2))
			if err != nil {
				t.Fatal(err)
			}
			defer overlay.Close()

			got := outputPages(t, func(out io.Writer) error {
				return tt.fn(t.Context(), out, inputFileName, overlay)
			})
			if got != 3 {
				t.Errorf("%s() produced %d pages, want 3", tt.name, got)
			}
		})
	}
}

// File names that look like a handle or an operation must still be usable.
func TestInputFileNames_pdftk(t *testing.T) {
	requirePDFtk(t)
	for _, name := range []string{"cat", "X=blank.pdf"} {
		t.Run(name, func(t *testing.T) {
			// pdftk only misparses bare names, so run it from the file's directory
			t.Chdir(filepath.Dir(writePDF(t, name, blankPDFObjects(3)...)))
			inputFileName := name
			got, err := NumberOfPages(t.Context(), inputFileName)
			if err != nil {
				t.Fatalf("NumberOfPages() error = %v", err)
			}
			if got != 3 {
				t.Errorf("NumberOfPages() = %d, want 3", got)
			}

			stamp, err := os.Open(writeBlankPDF(t, 1))
			if err != nil {
				t.Fatal(err)
			}
			defer stamp.Close()
			if got := outputPages(t, func(out io.Writer) error {
				return Stamp(t.Context(), out, inputFileName, stamp)
			}); got != 3 {
				t.Errorf("Stamp() produced %d pages, want 3", got)
			}
		})
	}
}
