package pdftk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

const fakePDFtkEnv = "GO_PDFTOOLS_FAKE_PDFTK"

// When re-executed with fakePDFtkEnv set, the test binary stands in for pdftk
// so that commands can be tested without pdftk installed.
func TestMain(m *testing.M) {
	switch mode := os.Getenv(fakePDFtkEnv); mode {
	case "":
		os.Exit(m.Run())
	case "echo":
		// echo the arguments, file arguments as <contents>, followed by stdin
		args := make([]string, len(os.Args)-1)
		for i, arg := range os.Args[1:] {
			args[i] = describeFileArg(arg)
		}
		fmt.Println(strings.Join(args, " "))
		_, _ = io.Copy(os.Stdout, os.Stdin)
	case "dump_data":
		if len(os.Args) != 5 || !strings.HasPrefix(os.Args[1], "A=") || os.Args[2] != "dump_data" || os.Args[3] != "output" || os.Args[4] != "-" {
			fmt.Fprintf(os.Stderr, "Error: unexpected arguments %q\n", os.Args[1:])
			os.Exit(1)
		}
		fmt.Printf("InfoBegin\nInfoKey: Title\nInfoValue: %s\nNumberOfPages: 7\nPageMediaBegin\n", os.Args[1])
	case "fail":
		fmt.Fprintln(os.Stderr, "Error: something went wrong")
		os.Exit(1)
	case "sleep":
		time.Sleep(time.Minute)
	default:
		fmt.Fprintf(os.Stderr, "unknown fake pdftk mode %q\n", mode)
		os.Exit(2)
	}
	os.Exit(0)
}

// Replace an absolute path, optionally after a handle=, with <its contents>
// so that tests need not know temp file names.
func describeFileArg(arg string) string {
	handle, path := "", arg
	if i := strings.IndexByte(arg, '='); i > 0 && strings.Trim(arg[:i], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") == "" {
		handle, path = arg[:i+1], arg[i+1:]
	}
	if !filepath.IsAbs(path) {
		return arg
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return arg
	}
	return handle + "<" + string(b) + ">"
}

// Run the test binary in place of pdftk in the given mode.
func fakePDFtk(t *testing.T, mode string) Option {
	t.Setenv(fakePDFtkEnv, mode)
	return OptionExecutable(os.Args[0])
}

// Command with the test binary as executable so pdftk need not be installed.
func testCommand(t *testing.T, options ...Option) *command {
	t.Helper()
	cmd, err := newCommand(t.Context(), io.Discard, append([]Option{OptionExecutable(os.Args[0])}, options...))
	if err != nil {
		t.Fatal(err)
	}
	return cmd
}

// Write content to a new file in a temp directory and open it.
func tempFile(t *testing.T, name, content string) *os.File {
	t.Helper()
	return openFile(t, writeFile(t, name, content))
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	return writeFileIn(t, t.TempDir(), name, content)
}

func writeFileIn(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func openFile(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func offset(t *testing.T, f *os.File) int64 {
	t.Helper()
	n, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// Fail when dir still holds files, i.e. a temp copy was not removed.
func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("%s holds %d leftover files, first %q", dir, len(entries), entries[0].Name())
	}
}

func requireSameFile(t *testing.T, path string, f *os.File) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Errorf("path %q is not absolute", path)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	named, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(fi, named) {
		t.Errorf("path %q is not %s", path, f.Name())
	}
}

// Unread regular files are passed by name and left unread.
func Test_command_input_file(t *testing.T) {
	t.Run("absolute name", func(t *testing.T) {
		f := tempFile(t, "in.pdf", "pdf data")
		cmd := testCommand(t)
		got, err := cmd.input(f)
		if err != nil {
			t.Fatal(err)
		}
		requireSameFile(t, got, f)
		if len(cmd.tempFiles) != 0 {
			t.Errorf("input() copied the file to %v", cmd.tempFiles)
		}
		if got := offset(t, f); got != 0 {
			t.Errorf("file offset = %d after input(), want 0", got)
		}
	})

	t.Run("relative name is made absolute", func(t *testing.T) {
		t.Chdir(filepath.Dir(writeFile(t, "cat", "pdf data")))
		f := openFile(t, "cat")
		got, err := testCommand(t).input(f)
		if err != nil {
			t.Fatal(err)
		}
		requireSameFile(t, got, f)
	})

	t.Run("symlink is resolved", func(t *testing.T) {
		target := writeFile(t, "real.pdf", "pdf data")
		link := filepath.Join(filepath.Dir(target), "link.pdf")
		if err := os.Symlink(target, link); err != nil {
			t.Skip(err)
		}
		f := openFile(t, link)
		got, err := testCommand(t).input(f)
		if err != nil {
			t.Fatal(err)
		}
		requireSameFile(t, got, f)
		if fi, err := os.Lstat(got); err != nil || fi.Mode()&fs.ModeSymlink != 0 {
			t.Errorf("input() = %q, want the symlink target", got)
		}
	})

	t.Run("same file under several handles", func(t *testing.T) {
		f := tempFile(t, "in.pdf", "pdf data")
		cmd := testCommand(t)
		first, err := cmd.input(f)
		if err != nil {
			t.Fatal(err)
		}
		second, err := cmd.input(f)
		if err != nil {
			t.Fatal(err)
		}
		if first != second {
			t.Errorf("input() = %q then %q, want the same path", first, second)
		}
	})
}

// Anything else is copied to a temp file that cleanup removes.
func Test_command_input_copied(t *testing.T) {
	tests := []struct {
		name      string
		reader    func(t *testing.T) io.Reader
		want      string // contents pdftk gets
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "in-memory reader",
			reader: func(*testing.T) io.Reader { return strings.NewReader("data") },
			want:   "data",
		},
		{
			name: "partially read file",
			reader: func(t *testing.T) io.Reader {
				f := tempFile(t, "in.pdf", "data")
				if _, err := f.Read(make([]byte, 1)); err != nil {
					t.Fatal(err)
				}
				return f
			},
			want: "ata",
		},
		{
			name: "pipe",
			reader: func(t *testing.T) io.Reader {
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { r.Close() })
				if _, err := w.WriteString("data"); err != nil {
					t.Fatal(err)
				}
				w.Close()
				return r
			},
			want: "data",
		},
		{
			name: "deleted file",
			reader: func(t *testing.T) io.Reader {
				f := tempFile(t, "in.pdf", "data")
				if err := os.Remove(f.Name()); err != nil {
					t.Skip(err)
				}
				return f
			},
			want: "data",
		},
		{
			name: "renamed file",
			reader: func(t *testing.T) io.Reader {
				f := tempFile(t, "in.pdf", "data")
				if err := os.Rename(f.Name(), f.Name()+".moved"); err != nil {
					t.Skip(err)
				}
				return f
			},
			want: "data",
		},
		{
			name: "replaced file",
			reader: func(t *testing.T) io.Reader {
				f := tempFile(t, "in.pdf", "data")
				if err := os.Remove(f.Name()); err != nil {
					t.Skip(err)
				}
				if err := os.WriteFile(f.Name(), []byte("other"), 0o600); err != nil {
					t.Fatal(err)
				}
				return f
			},
			want: "data",
		},
		{
			name: "relative name after changing directory",
			reader: func(t *testing.T) io.Reader {
				t.Chdir(filepath.Dir(writeFile(t, "in.pdf", "data")))
				f := openFile(t, "in.pdf")
				t.Chdir(filepath.Dir(writeFile(t, "in.pdf", "other")))
				return f
			},
			want: "data",
		},
		{
			name: "directory",
			reader: func(t *testing.T) io.Reader {
				return openFile(t, t.TempDir())
			},
			wantErr: true,
		},
		{
			name: "closed file",
			reader: func(t *testing.T) io.Reader {
				f := tempFile(t, "in.pdf", "data")
				f.Close()
				return f
			},
			wantErr:   true,
			wantErrIs: os.ErrClosed,
		},
		{
			name:      "typed nil file",
			reader:    func(*testing.T) io.Reader { return (*os.File)(nil) },
			wantErr:   true,
			wantErrIs: os.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cmd := testCommand(t, OptionTempDir(dir))
			got, err := cmd.input(tt.reader(t))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("input() = %q, want error", got)
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
				cmd.cleanup()
				requireEmptyDir(t, dir)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !filepath.IsAbs(got) || filepath.Dir(got) != dir {
				t.Errorf("input() = %q, want a file in %s", got, dir)
			}
			if content := readFile(t, got); content != tt.want {
				t.Errorf("temp file holds %q, want %q", content, tt.want)
			}
			cmd.cleanup()
			requireEmptyDir(t, dir)
		})
	}
}

// Has a slice field, so values cannot be compared.
type uncomparableReader struct {
	io.Reader
	_ []int
}

func Test_command_input_sameReader(t *testing.T) {
	t.Run("copied once", func(t *testing.T) {
		dir := t.TempDir()
		cmd := testCommand(t, OptionTempDir(dir))
		r := strings.NewReader("data")
		first, err := cmd.input(r)
		if err != nil {
			t.Fatal(err)
		}
		second, err := cmd.input(r)
		if err != nil {
			t.Fatal(err)
		}
		if first != second || len(cmd.tempFiles) != 1 {
			t.Errorf("input() = %q then %q with %d temp files, want one shared copy", first, second, len(cmd.tempFiles))
		}
		if content := readFile(t, first); content != "data" {
			t.Errorf("temp file holds %q, want data", content)
		}
	})

	t.Run("equal contents are still separate inputs", func(t *testing.T) {
		cmd := testCommand(t, OptionTempDir(t.TempDir()))
		first, err := cmd.input(strings.NewReader("data"))
		if err != nil {
			t.Fatal(err)
		}
		second, err := cmd.input(strings.NewReader("data"))
		if err != nil {
			t.Fatal(err)
		}
		if first == second {
			t.Errorf("input() = %q twice, want separate copies", first)
		}
	})

	t.Run("uncomparable reader does not panic", func(t *testing.T) {
		cmd := testCommand(t, OptionTempDir(t.TempDir()))
		r := uncomparableReader{Reader: strings.NewReader("data")}
		for range 2 {
			if _, err := cmd.input(r); err != nil {
				t.Fatal(err)
			}
		}
		if len(cmd.tempFiles) != 2 {
			t.Errorf("%d temp files, want 2", len(cmd.tempFiles))
		}
	})
}

// Cancels ctx on its first Read and never reaches EOF.
type cancelingReader struct {
	cancel context.CancelFunc
}

func (r cancelingReader) Read(p []byte) (int, error) {
	r.cancel()
	return len(p), nil
}

// Changes the working directory while being read.
type chdirReader struct {
	dir string
	io.Reader
}

func (r chdirReader) Read(p []byte) (int, error) {
	if err := os.Chdir(r.dir); err != nil {
		return 0, err
	}
	return r.Reader.Read(p)
}

func Test_command_input_errors(t *testing.T) {
	t.Run("nil reader", func(t *testing.T) {
		if _, err := testCommand(t).input(nil); !errors.Is(err, errNilInput) {
			t.Errorf("input(nil) error = %v, want %v", err, errNilInput)
		}
	})

	t.Run("canceled context stops the copy", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		dir := t.TempDir()
		cmd, err := newCommand(ctx, io.Discard, []Option{OptionExecutable(os.Args[0]), OptionTempDir(dir)})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := cmd.input(cancelingReader{cancel}); !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
		cmd.cleanup()
		requireEmptyDir(t, dir)
	})

	t.Run("read error is reported", func(t *testing.T) {
		errRead := errors.New("read failed")
		dir := t.TempDir()
		cmd := testCommand(t, OptionTempDir(dir))
		if _, err := cmd.input(iotest.ErrReader(errRead)); !errors.Is(err, errRead) {
			t.Errorf("error = %v, want %v", err, errRead)
		}
		cmd.cleanup()
		requireEmptyDir(t, dir)
	})

	// the copy must be found by pdftk and cleanup even after a change of directory
	t.Run("relative temp directory", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := os.Mkdir("tmp", 0o700); err != nil {
			t.Fatal(err)
		}
		dir, err := filepath.Abs("tmp")
		if err != nil {
			t.Fatal(err)
		}
		cmd := testCommand(t, OptionTempDir("tmp"))
		got, err := cmd.input(chdirReader{t.TempDir(), strings.NewReader("data")})
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(got) || filepath.Dir(got) != dir {
			t.Errorf("input() = %q, want a file in %s", got, dir)
		}
		if content := readFile(t, got); content != "data" {
			t.Errorf("temp file holds %q, want data", content)
		}
		cmd.cleanup()
		requireEmptyDir(t, dir)
	})

	t.Run("missing temp directory", func(t *testing.T) {
		cmd := testCommand(t, OptionTempDir(filepath.Join(t.TempDir(), "missing")))
		if _, err := cmd.input(strings.NewReader("data")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("error = %v, want fs.ErrNotExist", err)
		}
	})
}
