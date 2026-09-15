package pdftk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/iotest"
	"time"
)

func TestCommands_arguments(t *testing.T) {
	tests := []struct {
		name string
		call func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error
		want string
	}{
		{
			name: "Cat",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				pageRanges := []PageRange{
					{FileHandleName: "B", Qualifier: Odd},
					{FileHandleName: "A", BeginPage: 2, EndPage: 3, Rotation: East},
				}
				inputs := NewInputMap(tempFile(t, "first.pdf", "first"), tempFile(t, "second.pdf", "second"))
				return Cat(ctx, out, inputs, pageRanges, options...)
			},
			want: "A=<first> B=<second> cat Bodd A2-3east output -\n",
		},
		{
			name: "Cat copies readers",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				inputs := NewInputMap(strings.NewReader("first"), tempFile(t, "second.pdf", "second"), strings.NewReader("third"))
				return Cat(ctx, out, inputs, nil, options...)
			},
			want: "A=<first> B=<second> C=<third> cat output -\n",
		},
		{
			name: "Cat orders handles",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				inputs := InputMap{"BA": strings.NewReader("third"), "B": strings.NewReader("second"), "A": strings.NewReader("first")}
				return Cat(ctx, out, inputs, nil, options...)
			},
			want: "A=<first> B=<second> BA=<third> cat output -\n",
		},
		{
			name: "Cat same reader twice",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				r := strings.NewReader("first")
				return Cat(ctx, out, NewInputMap(r, r), nil, options...)
			},
			want: "A=<first> B=<first> cat output -\n",
		},
		{
			name: "Cat flatten",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return Cat(ctx, out, NewInputMap(tempFile(t, "first.pdf", "first")), nil, append(options, OptionFlatten())...)
			},
			want: "A=<first> cat output - flatten\n",
		},
		{
			name: "FillForm",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return FillForm(ctx, out, tempFile(t, "in.pdf", "pdf data"), strings.NewReader("fdf data"), options...)
			},
			want: "A=<pdf data> fill_form <fdf data> output -\n",
		},
		{
			name: "FillForm flatten",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return FillForm(ctx, out, tempFile(t, "in.pdf", "pdf data"), strings.NewReader("fdf data"), append(options, OptionFlatten())...)
			},
			want: "A=<pdf data> fill_form <fdf data> output - flatten\n",
		},
		{
			name: "FillForm from readers",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return FillForm(ctx, out, strings.NewReader("pdf data"), strings.NewReader("fdf data"), options...)
			},
			want: "A=<pdf data> fill_form <fdf data> output -\n",
		},
		{
			name: "FillForm from FDF file",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return FillForm(ctx, out, strings.NewReader("pdf data"), tempFile(t, "data.fdf", "fdf data"), options...)
			},
			want: "A=<pdf data> fill_form <fdf data> output -\n",
		},
		{
			name: "Background",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return Background(ctx, out, tempFile(t, "in.pdf", "pdf data"), strings.NewReader("overlay"), options...)
			},
			want: "A=<pdf data> background <overlay> output -\n",
		},
		{
			name: "MultiBackground",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return MultiBackground(ctx, out, tempFile(t, "in.pdf", "pdf data"), strings.NewReader("overlay"), options...)
			},
			want: "A=<pdf data> multibackground <overlay> output -\n",
		},
		{
			name: "Stamp",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return Stamp(ctx, out, tempFile(t, "in.pdf", "pdf data"), strings.NewReader("overlay"), options...)
			},
			want: "A=<pdf data> stamp <overlay> output -\n",
		},
		{
			name: "Stamp from readers",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return Stamp(ctx, out, strings.NewReader("pdf data"), strings.NewReader("overlay"), options...)
			},
			want: "A=<pdf data> stamp <overlay> output -\n",
		},
		{
			name: "MultiStamp",
			call: func(t *testing.T, ctx context.Context, out io.Writer, options ...Option) error {
				return MultiStamp(ctx, out, tempFile(t, "in.pdf", "pdf data"), strings.NewReader("overlay"), options...)
			},
			want: "A=<pdf data> multistamp <overlay> output -\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			dir := t.TempDir()
			if err := tt.call(t, t.Context(), &out, fakePDFtk(t, "echo"), OptionTempDir(dir)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("pdftk called with %q, want %q", got, tt.want)
			}
			requireEmptyDir(t, dir)
		})
	}
}

// An unread file is left for pdftk to read; a copied one is read to the end.
func TestCommands_readerConsumption(t *testing.T) {
	unread := tempFile(t, "in.pdf", "pdf data")
	partial := tempFile(t, "overlay.pdf", "overlay")
	if _, err := partial.Read(make([]byte, 3)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Stamp(t.Context(), &out, unread, partial, fakePDFtk(t, "echo")); err != nil {
		t.Fatal(err)
	}
	if want := "A=<pdf data> stamp <rlay> output -\n"; out.String() != want {
		t.Errorf("pdftk called with %q, want %q", out.String(), want)
	}
	if got := offset(t, unread); got != 0 {
		t.Errorf("unread file offset = %d, want 0", got)
	}
	if got := offset(t, partial); got != int64(len("overlay")) {
		t.Errorf("copied file offset = %d, want EOF at %d", got, len("overlay"))
	}
}

func TestNumberOfPages(t *testing.T) {
	inputs := map[string]io.Reader{
		"file":   tempFile(t, "in.pdf", "pdf data"),
		"reader": strings.NewReader("pdf data"),
	}
	for name, in := range inputs {
		t.Run(name, func(t *testing.T) {
			got, err := NumberOfPages(t.Context(), in, fakePDFtk(t, "dump_data"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != 7 {
				t.Errorf("NumberOfPages() = %d, want 7", got)
			}
		})
	}
}

func TestNumberOfPages_unexpectedOutput(t *testing.T) {
	if _, err := NumberOfPages(t.Context(), strings.NewReader("pdf data"), fakePDFtk(t, "echo")); err == nil {
		t.Error("expected error when dump_data output has no NumberOfPages")
	}
}

func TestCommands_errors(t *testing.T) {
	t.Run("stderr is reported", func(t *testing.T) {
		err := Cat(t.Context(), io.Discard, NewInputMap(strings.NewReader("first")), nil, fakePDFtk(t, "fail"))
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			t.Fatalf("error = %v, want *exec.ExitError", err)
		}
		got := err.Error()
		if !strings.HasPrefix(got, "pdftk error: Error: something went wrong") || !strings.HasSuffix(got, ": exit status 1") {
			t.Errorf("error = %q, want pdftk error: <stderr>: exit status 1", got)
		}
	})

	t.Run("executable not found", func(t *testing.T) {
		dir := t.TempDir()
		options := []Option{OptionExecutable("go-pdftools-missing-executable"), OptionTempDir(dir)}
		in, data := strings.NewReader("pdf data"), strings.NewReader("overlay")
		calls := map[string]func() error{
			"Cat": func() error {
				return Cat(t.Context(), io.Discard, NewInputMap(in), nil, options...)
			},
			"Stamp": func() error {
				return Stamp(t.Context(), io.Discard, in, data, options...)
			},
			"NumberOfPages": func() error {
				_, err := NumberOfPages(t.Context(), in, options...)
				return err
			},
		}
		for name, call := range calls {
			if err := call(); !errors.Is(err, exec.ErrNotFound) {
				t.Errorf("%s error = %v, want exec.ErrNotFound", name, err)
			}
			// nothing is copied for an executable that cannot run
			if in.Len() != int(in.Size()) || data.Len() != int(data.Size()) {
				t.Errorf("%s read its inputs before failing", name)
			}
		}
		requireEmptyDir(t, dir)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		err := Cat(ctx, io.Discard, NewInputMap(tempFile(t, "first.pdf", "first")), nil, fakePDFtk(t, "sleep"))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()
		err := Cat(ctx, io.Discard, NewInputMap(strings.NewReader("first")), nil, fakePDFtk(t, "sleep"))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("error = %v, want context.DeadlineExceeded", err)
		}
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			t.Errorf("error = %v, want *exec.ExitError to be preserved", err)
		}
	})

	t.Run("context canceled while copying an input", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		dir := t.TempDir()
		var out bytes.Buffer
		err := Stamp(ctx, &out, strings.NewReader("pdf data"), cancelingReader{cancel}, fakePDFtk(t, "echo"), OptionTempDir(dir))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
		if out.Len() != 0 {
			t.Errorf("pdftk was run with %q, want not run", out.String())
		}
		requireEmptyDir(t, dir)
	})

	t.Run("input read error after copies were made", func(t *testing.T) {
		errRead := errors.New("read failed")
		dir := t.TempDir()
		var out bytes.Buffer
		inputs := NewInputMap(strings.NewReader("first"), strings.NewReader("second"), iotest.ErrReader(errRead))
		err := Cat(t.Context(), &out, inputs, nil, fakePDFtk(t, "echo"), OptionTempDir(dir))
		if !errors.Is(err, errRead) {
			t.Errorf("error = %v, want %v", err, errRead)
		}
		if out.Len() != 0 {
			t.Errorf("pdftk was run with %q, want not run", out.String())
		}
		requireEmptyDir(t, dir)
	})

	t.Run("missing temp directory", func(t *testing.T) {
		var out bytes.Buffer
		err := Stamp(t.Context(), &out, strings.NewReader("pdf data"), strings.NewReader("overlay"), fakePDFtk(t, "echo"), OptionTempDir(t.TempDir()+"/missing"))
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("error = %v, want fs.ErrNotExist", err)
		}
		if out.Len() != 0 {
			t.Errorf("pdftk was run with %q, want not run", out.String())
		}
	})

	t.Run("temp files are removed when pdftk fails", func(t *testing.T) {
		dir := t.TempDir()
		err := Stamp(t.Context(), io.Discard, strings.NewReader("pdf data"), strings.NewReader("overlay"), fakePDFtk(t, "fail"), OptionTempDir(dir))
		if err == nil {
			t.Fatal("expected error")
		}
		requireEmptyDir(t, dir)
	})

	t.Run("nil input", func(t *testing.T) {
		dir := t.TempDir()
		pdf, overlay, first := strings.NewReader("pdf data"), strings.NewReader("overlay"), strings.NewReader("first")
		calls := map[string]func() error{
			"Stamp input": func() error {
				return Stamp(t.Context(), io.Discard, nil, overlay, fakePDFtk(t, "echo"), OptionTempDir(dir))
			},
			"Stamp overlay": func() error {
				return Stamp(t.Context(), io.Discard, pdf, nil, fakePDFtk(t, "echo"), OptionTempDir(dir))
			},
			"Cat": func() error {
				return Cat(t.Context(), io.Discard, InputMap{"A": first, "B": nil}, nil, fakePDFtk(t, "echo"), OptionTempDir(dir))
			},
			"NumberOfPages": func() error {
				_, err := NumberOfPages(t.Context(), nil, fakePDFtk(t, "dump_data"), OptionTempDir(dir))
				return err
			},
		}
		for name, call := range calls {
			if err := call(); !errors.Is(err, errNilInput) {
				t.Errorf("%s error = %v, want %v", name, err, errNilInput)
			}
		}
		// nothing is copied when another input is nil
		for _, r := range []*strings.Reader{pdf, overlay, first} {
			if r.Len() != int(r.Size()) {
				t.Error("an input was read although another one was nil")
			}
		}
		requireEmptyDir(t, dir)
	})
}

// Concurrent calls sharing a temp directory must not interfere.
func TestCommands_concurrent(t *testing.T) {
	dir := t.TempDir()
	options := []Option{fakePDFtk(t, "echo"), OptionTempDir(dir)}
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			n := strconv.Itoa(i)
			var out bytes.Buffer
			if err := Stamp(t.Context(), &out, strings.NewReader("pdf "+n), strings.NewReader("overlay "+n), options...); err != nil {
				t.Error(err)
				return
			}
			if want := "A=<pdf " + n + "> stamp <overlay " + n + "> output -\n"; out.String() != want {
				t.Errorf("pdftk called with %q, want %q", out.String(), want)
			}
		})
	}
	wg.Wait()
	requireEmptyDir(t, dir)
}
