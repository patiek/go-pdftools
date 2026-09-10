package fdf_test

import (
	"bytes"
	"os"

	"github.com/patiek/go-pdftools/fdf"
)

func ExampleWrite() {
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
	var b bytes.Buffer
	err := fdf.Write(&b, fdf.Inputs{
		"foo": fdf.OptionInput("Yes"), // mark field "foo" as checked
		"bar": "field 2 value",
		"baz": fdf.OptionInput("United States"), // select the value "United States" for field "baz"
	})
	if err != nil {
		// handle error
	}
}
