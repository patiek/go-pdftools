# go-pdftools
PDF utilities to fill PDFs via FDF and manipulating them with PDFtk

Requires Go 1.26 or later.

## FDF
Package fdf provides functions to generate FDF (Forms Data Format) useful for filling PDF forms with data.
### Documentation
https://pkg.go.dev/github.com/patiek/go-pdftools/fdf

## PDFtk
Package pdftk provides wrapper functions for calling PDFtk commands.

Expects command line executable of pdftk or pdftk-java to be installed.

Every command takes a `context.Context` so pdftk can be cancelled or given a deadline:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

pages, err := pdftk.NumberOfPages(ctx, "input.pdf")
```
### Documentation
https://pkg.go.dev/github.com/patiek/go-pdftools/pdftk

## Docker
The sample [Dockerfile](Dockerfile) builds an image with Go and pdftk-java installed and runs the tests, including those that need pdftk:

```sh
docker build -t go-pdftools .
```

Use the same `apt-get install pdftk-java` step in your own image to make pdftk available to your application.

## Upgrading to v0.2.0
All pdftk functions now take a `context.Context` as their first argument. Pass `context.Background()` to keep the previous behavior.

Errors now wrap the underlying `exec` or context error, so check them with `errors.Is` (for example `context.DeadlineExceeded`) or `errors.As` (for `*exec.ExitError`) rather than matching the message text.

`OptionExecutable` previously had no effect and `pdftk` was always run; it now runs the named executable.
