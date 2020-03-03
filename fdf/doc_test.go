package fdf_test

import (
	"bytes"
	"github.com/patiek/go-pdftools/fdf"
	"os"
)

func ExampleWriteFDF() {
	var (
		b   bytes.Buffer
		err error
	)

	// write FDF data into buffer
	err = fdf.Write(&b, fdf.Inputs{
		"field 1":     "field 1 value",
		"field 2":     "field 2 value",
		"foo.bar.baz": "structured fields also work",
	})
	if err != nil {
		// handle error
	}

	// create and write FDF data into out.fdf
	f, err := os.Create("out.fdf")
	if err != nil {
		// handle error
	}
	defer f.Close()

	f.Write(b.Bytes())
}

func ExampleOptionInput() {
	inputs := fdf.Inputs{
		"foo": fdf.OptionInput("Yes"), // mark field "foo" as checked
		"bar": "field 2 value",
		"baz": fdf.OptionInput("United States"), // select the value "United States" for field "baz"
	}
	// fdf.Write(w, inputs)
}
