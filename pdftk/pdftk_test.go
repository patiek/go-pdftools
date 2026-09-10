package pdftk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCommands_arguments(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, out io.Writer, options ...Option) error
		want string
	}{
		{
			name: "Cat",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				pageRanges := []PageRange{
					{FileHandleName: "B", Qualifier: Odd},
					{FileHandleName: "A", BeginPage: 2, EndPage: 3, Rotation: East},
				}
				return Cat(ctx, out, NewInputFileMap("first.pdf", "second.pdf"), pageRanges, options...)
			},
			want: "A=first.pdf B=second.pdf cat Bodd A2-3east output -\n",
		},
		{
			name: "Cat flatten",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return Cat(ctx, out, NewInputFileMap("first.pdf"), nil, append(options, OptionFlatten())...)
			},
			want: "A=first.pdf cat output - flatten\n",
		},
		{
			name: "FillForm",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return FillForm(ctx, out, "in.pdf", strings.NewReader("fdf data"), options...)
			},
			want: "A=in.pdf fill_form - output -\nfdf data",
		},
		{
			name: "FillForm flatten",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return FillForm(ctx, out, "in.pdf", strings.NewReader("fdf data"), append(options, OptionFlatten())...)
			},
			want: "A=in.pdf fill_form - output - flatten\nfdf data",
		},
		{
			name: "Background",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return Background(ctx, out, "in.pdf", strings.NewReader("pdf data"), options...)
			},
			want: "A=in.pdf background - output -\npdf data",
		},
		{
			name: "MultiBackground",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return MultiBackground(ctx, out, "in.pdf", strings.NewReader("pdf data"), options...)
			},
			want: "A=in.pdf multibackground - output -\npdf data",
		},
		{
			name: "Stamp",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return Stamp(ctx, out, "in.pdf", strings.NewReader("pdf data"), options...)
			},
			want: "A=in.pdf stamp - output -\npdf data",
		},
		{
			name: "MultiStamp",
			call: func(ctx context.Context, out io.Writer, options ...Option) error {
				return MultiStamp(ctx, out, "in.pdf", strings.NewReader("pdf data"), options...)
			},
			want: "A=in.pdf multistamp - output -\npdf data",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := tt.call(t.Context(), &out, fakePDFtk(t, "echo")); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("pdftk called with %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNumberOfPages(t *testing.T) {
	got, err := NumberOfPages(t.Context(), "in.pdf", fakePDFtk(t, "dump_data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7 {
		t.Errorf("NumberOfPages() = %d, want 7", got)
	}
}

func TestNumberOfPages_unexpectedOutput(t *testing.T) {
	if _, err := NumberOfPages(t.Context(), "in.pdf", fakePDFtk(t, "echo")); err == nil {
		t.Error("expected error when dump_data output has no NumberOfPages")
	}
}

func TestCommands_errors(t *testing.T) {
	t.Run("stderr is reported", func(t *testing.T) {
		err := Cat(t.Context(), io.Discard, NewInputFileMap("first.pdf"), nil, fakePDFtk(t, "fail"))
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			t.Fatalf("error = %v, want *exec.ExitError", err)
		}
		got := err.Error()
		if !strings.HasPrefix(got, "pdftk error: Error: something went wrong") || !strings.HasSuffix(got, ": exit status 1") {
			t.Errorf("error = %q, want pdftk error: <stderr>: exit status 1", got)
		}
	})

	t.Run("executable not found", func(t *testing.T) {
		err := Cat(t.Context(), io.Discard, NewInputFileMap("first.pdf"), nil, OptionExecutable("go-pdftools-missing-executable"))
		if !errors.Is(err, exec.ErrNotFound) {
			t.Errorf("error = %v, want exec.ErrNotFound", err)
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		err := Cat(ctx, io.Discard, NewInputFileMap("first.pdf"), nil, fakePDFtk(t, "sleep"))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()
		err := Cat(ctx, io.Discard, NewInputFileMap("first.pdf"), nil, fakePDFtk(t, "sleep"))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("error = %v, want context.DeadlineExceeded", err)
		}
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			t.Errorf("error = %v, want *exec.ExitError to be preserved", err)
		}
	})
}
