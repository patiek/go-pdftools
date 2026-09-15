package pdftk_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/patiek/go-pdftools/fdf"
	"github.com/patiek/go-pdftools/pdftk"
)

func ExampleCat() {
	first, err := os.Open("first.pdf")
	if err != nil {
		// handle error
	}
	defer first.Close()

	second, err := os.Open("second.pdf")
	if err != nil {
		// handle error
	}
	defer second.Close()

	// file to write output into
	out, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer out.Close()

	err = pdftk.Cat(context.Background(), out, pdftk.NewInputMap(first, second), []pdftk.PageRange{
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
	first, err := os.Open("first.pdf")
	if err != nil {
		// handle error
	}
	defer first.Close()

	second, err := os.Open("second.pdf")
	if err != nil {
		// handle error
	}
	defer second.Close()

	third, err := os.Open("third.pdf")
	if err != nil {
		// handle error
	}
	defer third.Close()

	inputs := pdftk.InputMap{"A": first, "B": second, "C": third}

	// file to write output into
	out, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer out.Close()

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

	err = pdftk.Cat(context.Background(), out, inputs, pageRanges)
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

	form, err := os.Open("form.pdf")
	if err != nil {
		// handle error
	}
	defer form.Close()

	out, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer out.Close()

	// fill form with FDF data from buffer b and flatten
	err = pdftk.FillForm(context.Background(), out, form, &b, pdftk.OptionFlatten())
	if err != nil {
		// handle error
	}
}

func ExampleFillForm_file() {
	form, err := os.Open("form.pdf")
	if err != nil {
		// handle error
	}
	defer form.Close()

	fdfFile, err := os.Open("input.fdf")
	if err != nil {
		// handle error
	}
	defer fdfFile.Close()

	out, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer out.Close()

	// fill form with FDF data from FDF file and flatten
	err = pdftk.FillForm(context.Background(), out, form, fdfFile, pdftk.OptionFlatten())
	if err != nil {
		// handle error
	}
}

func ExampleFillForm_timeout() {
	form, err := os.Open("form.pdf")
	if err != nil {
		// handle error
	}
	defer form.Close()

	fdfFile, err := os.Open("input.fdf")
	if err != nil {
		// handle error
	}
	defer fdfFile.Close()

	out, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer out.Close()

	// give pdftk at most 30 seconds to fill the form
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = pdftk.FillForm(ctx, out, form, fdfFile)
	if err != nil {
		// handle error, errors.Is(err, context.DeadlineExceeded) on timeout
	}
}

func ExampleBackground() {
	in, err := os.Open("input.pdf")
	if err != nil {
		// handle error
	}
	defer in.Close()

	background, err := os.Open("background.pdf")
	if err != nil {
		// handle error
	}
	defer background.Close()

	out, err := os.Create("out.pdf")
	if err != nil {
		// handle error
	}
	defer out.Close()

	// add background to every page of in
	err = pdftk.Background(context.Background(), out, in, background)
	if err != nil {
		// handle error
	}
}

func ExampleStamp() {
	// PDF held in memory, e.g. downloaded or generated
	var pdfData []byte

	stamp, err := os.Open("stamp.pdf")
	if err != nil {
		// handle error
	}
	defer stamp.Close()

	var out bytes.Buffer

	// stamp every page of pdfData and keep the result in memory
	err = pdftk.Stamp(context.Background(), &out, bytes.NewReader(pdfData), stamp)
	if err != nil {
		// handle error
	}
}

func ExampleNumberOfPages() {
	in, err := os.Open("input.pdf")
	if err != nil {
		// handle error
	}
	defer in.Close()

	pages, err := pdftk.NumberOfPages(context.Background(), in)
	if err != nil {
		// handle error
	}
	fmt.Println(pages)
}

func ExampleOptionExecutable() {
	in, err := os.Open("input.pdf")
	if err != nil {
		// handle error
	}
	defer in.Close()

	// use pdftk-java installed under a different name
	pages, err := pdftk.NumberOfPages(context.Background(), in, pdftk.OptionExecutable("pdftk-java"))
	if err != nil {
		// handle error
	}
	fmt.Println(pages)
}

func ExampleOptionTempDir() {
	var pdfData, stampData []byte

	// in-memory inputs are copied to files under /var/tmp for pdftk to read
	var out bytes.Buffer
	err := pdftk.Stamp(context.Background(), &out, bytes.NewReader(pdfData), bytes.NewReader(stampData), pdftk.OptionTempDir("/var/tmp"))
	if err != nil {
		// handle error
	}
}
