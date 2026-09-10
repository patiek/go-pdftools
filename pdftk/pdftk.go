package pdftk

import (
	"bytes"
	"context"
	"io"
)

const cmdPDFtk = "pdftk"

// Pass a single input as handle A so pdftk never mistakes its file name for
// a handle assignment or an operation.
func inputHandle(inputFileName string) string {
	return "A=" + inputFileName
}

// Concatenate PDF files and write to out.
// Specify pageRanges to manipulate ordering, ranges, and rotation of pages.
func Cat(ctx context.Context, out io.Writer, fileNames InputFileMap, pageRanges []PageRange, options ...Option) error {
	args := append(fileNames.parameterize(), "cat")
	args = append(args, pageRangesToStrings(pageRanges)...)
	args = append(args, "output", "-")
	cmd := createCmd(ctx, cmdPDFtk, out, nil, args...)
	cmd.applyOptions(options...)
	return cmd.run()
}

// Fill a PDF file with FDF data and write to out.
func FillForm(ctx context.Context, out io.Writer, inputFileName string, fdf io.Reader, options ...Option) error {
	cmd := createCmd(ctx, cmdPDFtk, out, fdf, inputHandle(inputFileName), "fill_form", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.run()
}

// Applies a PDF watermark to the background of a each page of input pdf.
// If background is multiple pages, only first page is used for the background.
func Background(ctx context.Context, out io.Writer, inputFileName string, background io.Reader, options ...Option) error {
	cmd := createCmd(ctx, cmdPDFtk, out, background, inputHandle(inputFileName), "background", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.run()
}

// Applies a PDF watermark to the background of a each page of input pdf.
// Similar to background but applies each page of background to corresponding
// input page. If input pdf has more pages than background, last page of
// background is used for the remainder of the input pages.
func MultiBackground(ctx context.Context, out io.Writer, inputFileName string, background io.Reader, options ...Option) error {
	cmd := createCmd(ctx, cmdPDFtk, out, background, inputHandle(inputFileName), "multibackground", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.run()
}

// Stamp (overlay) each page of input pdf with a stamp PDF and write to out.
// If stamp is multiple pages, only the first page is used for the stamp.
func Stamp(ctx context.Context, out io.Writer, inputFileName string, stamp io.Reader, options ...Option) error {
	cmd := createCmd(ctx, cmdPDFtk, out, stamp, inputHandle(inputFileName), "stamp", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.run()
}

// MultiStamp (overlay) each page of input pdf with stamp PDF and write to out.
// Similar to stamp but applies each page of stamp to corresponding input page.
// If input pdf has more pages than stamp, last page of stamp is used for the
// remainder of the input pages.
func MultiStamp(ctx context.Context, out io.Writer, inputFileName string, stamp io.Reader, options ...Option) error {
	cmd := createCmd(ctx, cmdPDFtk, out, stamp, inputHandle(inputFileName), "multistamp", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.run()
}

// Get the number of pages in a PDF file using dump_data.
func NumberOfPages(ctx context.Context, inputFileName string, options ...Option) (int, error) {
	var out bytes.Buffer
	cmd := createCmd(ctx, cmdPDFtk, &out, nil, inputHandle(inputFileName), "dump_data", "output", "-")
	cmd.applyOptions(options...)
	if err := cmd.run(); err != nil {
		return 0, err
	}
	return parseNumberOfPages(out.String())
}
