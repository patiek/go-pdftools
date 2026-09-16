package pdftk

import (
	"bytes"
	"context"
	"io"
)

const cmdPDFtk = "pdftk"

// Pass a single input as handle A so pdftk never mistakes its file name for
// a handle assignment or an operation.
func inputHandle(input string) string {
	return "A=" + input
}

// Concatenate PDFs and write to out.
// Specify pageRanges to manipulate ordering, ranges, and rotation of pages.
func Cat(ctx context.Context, out io.Writer, inputs InputMap, pageRanges []PageRange, options ...Option) error {
	cmd, err := newCommand(ctx, out, options)
	if err != nil {
		return err
	}
	defer cmd.cleanup()
	args, err := cmd.inputArgs(inputs)
	if err != nil {
		return err
	}
	args = append(args, "cat")
	args = append(args, pageRangesToStrings(pageRanges)...)
	return cmd.run(append(args, "output", "-")...)
}

// Fill a PDF form with FDF or XFDF data and write to out.
func FillForm(ctx context.Context, out io.Writer, in, fdf io.Reader, options ...Option) error {
	return combine(ctx, out, "fill_form", in, fdf, options)
}

// Applies a PDF watermark to the background of each page of input pdf.
// If background is multiple pages, only first page is used for the background.
func Background(ctx context.Context, out io.Writer, in, background io.Reader, options ...Option) error {
	return combine(ctx, out, "background", in, background, options)
}

// Applies a PDF watermark to the background of each page of input pdf.
// Similar to background but applies each page of background to corresponding
// input page. If input pdf has more pages than background, last page of
// background is used for the remainder of the input pages.
func MultiBackground(ctx context.Context, out io.Writer, in, background io.Reader, options ...Option) error {
	return combine(ctx, out, "multibackground", in, background, options)
}

// Stamp (overlay) each page of input pdf with a stamp PDF and write to out.
// If stamp is multiple pages, only the first page is used for the stamp.
func Stamp(ctx context.Context, out io.Writer, in, stamp io.Reader, options ...Option) error {
	return combine(ctx, out, "stamp", in, stamp, options)
}

// MultiStamp (overlay) each page of input pdf with stamp PDF and write to out.
// Similar to stamp but applies each page of stamp to corresponding input page.
// If input pdf has more pages than stamp, last page of stamp is used for the
// remainder of the input pages.
func MultiStamp(ctx context.Context, out io.Writer, in, stamp io.Reader, options ...Option) error {
	return combine(ctx, out, "multistamp", in, stamp, options)
}

// Get the number of pages in a PDF using dump_data.
func NumberOfPages(ctx context.Context, in io.Reader, options ...Option) (int, error) {
	var out bytes.Buffer
	cmd, err := newCommand(ctx, &out, options)
	if err != nil {
		return 0, err
	}
	defer cmd.cleanup()
	arg, err := cmd.input(in)
	if err != nil {
		return 0, err
	}
	if err := cmd.run(inputHandle(arg), "dump_data", "output", "-"); err != nil {
		return 0, err
	}
	return parseNumberOfPages(out.String())
}

// Run an operation that combines the input PDF with a second input.
func combine(ctx context.Context, out io.Writer, operation string, in, data io.Reader, options []Option) error {
	if data == nil {
		return errNilInput
	}
	cmd, err := newCommand(ctx, out, options)
	if err != nil {
		return err
	}
	defer cmd.cleanup()
	inArg, err := cmd.input(in)
	if err != nil {
		return err
	}
	dataArg, err := cmd.input(data)
	if err != nil {
		return err
	}
	return cmd.run(inputHandle(inArg), operation, dataArg, "output", "-")
}
