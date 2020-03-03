package pdftk_test

import (
	"bytes"
	"github.com/patiek/go-pdftools/fdf"
	"github.com/patiek/go-pdftools/pdftk"
	"log"
	"os"
)

func ExampleCat() {
	// file to write output into
	f, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}

	err = pdftk.Cat(f, pdftk.NewInputFileMap("first.pdf", "second.pdf", "third.pdf"), []pdftk.PageRange{
		{
			FileHandleName: pdftk.InputHandleNameFromInt(1),
			Rotation:       pdftk.East,
		},
		{
			FileHandleName: pdftk.InputHandleNameFromInt(0),
		},
	}, pdftk.OptionFlatten())
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCat_pageRanges() {
	// file to write output into
	f, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}

	inputFiles := pdftk.InputFileMap{
		"A": "first.pdf",
		"B": "second.pdf",
		"C": "third.pdf",
	}

	pageRanges := []pdftk.PageRange{
		// first we take page 2 from C
		{
			FileHandleName: "C",
			BeginPage:      2,
			EndPage:        2,
		},
		// then we take pages 4 until end of A and rotate them all
		{
			FileHandleName: "A",
			BeginPage:      4,
			Rotation:       pdftk.East,
		},
		// then we take all odd pages of B
		{
			FileHandleName: "B",
			Qualifier:      pdftk.Odd,
		},
		// finally we add in the first page from C
		{
			FileHandleName: "C",
			BeginPage:      1,
			EndPage:        1,
		},
	}

	err = pdftk.Cat(f, inputFiles, pageRanges)
	if err != nil {
		// handle error
	}
}

func ExampleFillForm() {
	var b bytes.Buffer
	if err := fdf.Write(&b, fdf.Inputs{
		"first field":  "hello",
		"second field": "world",
	}); err != nil {
		// handle error
	}

	f, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer f.Close()

	// fill form with FDF data from buffer b and flatten
	err = pdftk.FillForm(f, "test.pdf", &b, pdftk.OptionFlatten())
	if err != nil {
		// handle error
	}
}

func ExampleFillForm_file() {
	fdfFile, err := os.Open("input.fdf")
	if err != nil {
		// handle error
	}
	defer fdfFile.Close()

	outFile, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer outFile.Close()

	// fill form with FDF data from FDF file and flatten
	err = pdftk.FillForm(outFile, "test.pdf", fdfFile, pdftk.OptionFlatten())
	if err != nil {
		// handle error
	}
}
