package pdftk

import (
	"io"
)

const cmdPDFtk = "pdftk"

// Concatenate PDF files and write to out.
// Specify pageRanges to manipulate ordering, ranges, and rotation of pages.
func Cat(out io.Writer, fileNames InputFileMap, pageRanges []PageRange, options ...Option) error {
	args := append(fileNames.parameterize(), "cat")
	args = append(args, pageRangesToStrings(pageRanges)...)
	args = append(args, "output", "-")
	cmd := createCmd(cmdPDFtk, out, nil, args...)
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}

// Fill a PDF file with FDF data and write to out.
func FillForm(out io.Writer, inputFileName string, fdf io.Reader, options ...Option) error {
	cmd := createCmd(cmdPDFtk, out, fdf, inputFileName, "fill_form", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}

// Applies a PDF watermark to the background of a each page of input pdf.
// If background is multiple pages, only first page is used for the background.
func Background(out io.Writer, inputFileName string, background io.Reader, options ...Option) error {
	cmd := createCmd(cmdPDFtk, out, background, inputFileName, "background", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}

// Applies a PDF watermark to the background of a each page of input pdf.
// Similar to background but applies each page of background to corresponding
// input page. If input pdf has more pages than background, last page of
// background is used for the remainder of the input pages.
func MultiBackground(out io.Writer, inputFileName string, background io.Reader, options ...Option) error {
	cmd := createCmd(cmdPDFtk, out, background, inputFileName, "multibackground", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}

// Stamp (overlay) each page of input pdf with a stamp PDF and write to out.
// If stamp is multiple pages, only the first page is used for the stamp.
func Stamp(out io.Writer, inputFileName string, stamp io.Reader, options ...Option) error {
	cmd := createCmd(cmdPDFtk, out, stamp, inputFileName, "stamp", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}

// MultiStamp (overlay) each page of input pdf with stamp PDF and write to out.
// Similar to stamp but applies each page of stamp to corresponding input page.
// If input pdf has more pages than stamp, last page of stamp is used for the
// remainder of the input pages.
func MultiStamp(out io.Writer, inputFileName string, stamp io.Reader, options ...Option) error {
	cmd := createCmd(cmdPDFtk, out, stamp, inputFileName, "multistamp", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}
