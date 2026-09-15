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

Every command reads its inputs from an `io.Reader` and takes a `context.Context` so pdftk can be cancelled or given a deadline; commands that produce a PDF write it to an `io.Writer`:

```go
in, err := os.Open("input.pdf")
if err != nil {
	// handle error
}
defer in.Close()

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

pages, err := pdftk.NumberOfPages(ctx, in)
```

An unread regular `*os.File` is passed to pdftk by name and left unread, so there is no copy. Any other reader, such as a `*bytes.Reader` or an HTTP response body, is copied to a temp file for pdftk to open, because pdftk needs seekable files and holds anything streamed on stdin in memory. `OptionTempDir` chooses where the copies go; they are removed when the command returns.
### Documentation
https://pkg.go.dev/github.com/patiek/go-pdftools/pdftk

## Docker
The sample [Dockerfile](Dockerfile) builds an image with Go and pdftk-java installed and runs the tests, including those that need pdftk:

```sh
docker build -t go-pdftools .
```

Use the same `apt-get install pdftk-java` step in your own image to make pdftk available to your application. Set `LANG=C.UTF-8` (or another UTF-8 locale) as well: pdftk-java decodes file paths using the locale, and the paths this package passes include the working directory and the temp directory.

## Upgrading to v0.3.0
pdftk functions take `io.Reader` inputs instead of file names. `InputFileMap` and `NewInputFileMap` are replaced by `InputMap` and `NewInputMap`, which hold readers; `PageRange.FileHandleName` is unchanged and names a key in `InputMap`. Open files with `os.Open` and pass the `*os.File`:

```go
// before
pages, err := pdftk.NumberOfPages(ctx, "input.pdf")
err = pdftk.FillForm(ctx, out, "form.pdf", &fdfData)
err = pdftk.Cat(ctx, out, pdftk.NewInputFileMap("first.pdf", "second.pdf"), nil)

// after
in, err := os.Open("input.pdf")
if err != nil {
	// handle error
}
defer in.Close()
pages, err := pdftk.NumberOfPages(ctx, in)

form, err := os.Open("form.pdf")
// ...
err = pdftk.FillForm(ctx, out, form, &fdfData)

first, err := os.Open("first.pdf")
// ...
second, err := os.Open("second.pdf")
// ...
err = pdftk.Cat(ctx, out, pdftk.NewInputMap(first, second), nil)
```

Behavior to be aware of:

- An unread regular `*os.File` is still opened by pdftk itself, so the file case performs as before. Every other reader is copied to a temp file first (see `OptionTempDir`), including FDF data, which used to be streamed on stdin.
- Readers are read at most once and never closed. The same reader (a pointer or other comparable value) under several `Cat` handles is copied once.
- A missing input file is now reported by `os.Open` in your code before pdftk runs, rather than in pdftk's error text.
- pdftk's error text names the temp copy (`go-pdftools-*`, already removed) for reader inputs rather than your reader, so with several `Cat` readers it does not say which handle failed. Pass an unread `*os.File` to have your file name reported.
- Files whose bare name is a pdftk keyword, such as `cat` or `output`, work in every position because absolute paths are always passed. pdftk-java decodes those paths using the locale, so non-ASCII directory names need a UTF-8 locale such as `LANG=C.UTF-8`.
- A nil input returns an error instead of running pdftk.

## Upgrading to v0.2.0
All pdftk functions now take a `context.Context` as their first argument. Pass `context.Background()` to keep the previous behavior.

Errors now wrap the underlying `exec` or context error, so check them with `errors.Is` (for example `context.DeadlineExceeded`) or `errors.As` (for `*exec.ExitError`) rather than matching the message text.

`OptionExecutable` previously had no effect and `pdftk` was always run; it now runs the named executable.
