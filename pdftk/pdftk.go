package pdftk

import (
	"io"
)

const cmdPDFtk = "pdftk"

// Fill a PDF file with FDF data and write to out
func FillForm(out io.Writer, inputFileName string, fdf io.Reader, options ...Option) error {
	cmd := createCmd(cmdPDFtk, out, fdf, inputFileName, "fill_form", "-", "output", "-")
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}

// Concatenate PDF files and write to out
// Specify pageRanges to manipulate ordering, ranges, and rotation of pages.
func Cat(out io.Writer, fileNames InputFileMap, pageRanges []PageRange, options ...Option) error {
	args := append(fileNames.parameterize(), "cat")
	args = append(args, pageRangesToStrings(pageRanges)...)
	args = append(args, "output", "-")
	cmd := createCmd(cmdPDFtk, out, nil, args...)
	cmd.applyOptions(options...)
	return cmd.runWrapError()
}
