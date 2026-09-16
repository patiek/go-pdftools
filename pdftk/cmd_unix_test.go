//go:build unix

package pdftk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Cancelling must also stop pdftk when it is started by a wrapper script that
// does not exec, as Debian's pdftk-java does.
func TestCommands_cancelWrapperScript(t *testing.T) {
	t.Setenv(fakePDFtkEnv, "sleep")
	wrapper := filepath.Join(t.TempDir(), "pdftk")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\n'"+os.Args[0]+"' \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := Cat(ctx, io.Discard, NewInputMap(strings.NewReader("first")), nil, OptionExecutable(wrapper))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed >= waitDelay {
		t.Errorf("Cat() returned after %v, want well under %v", elapsed, waitDelay)
	}
}

// A name such as /dev/fd/3 (or /dev/stdin) means something else inside the
// pdftk process, so it must never be passed on. Linux resolves it to the real
// file, other systems copy it.
func Test_command_input_pseudoPath(t *testing.T) {
	f := tempFile(t, "in.pdf", "data")
	fd, err := syscall.Dup(int(f.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	pseudo := os.NewFile(uintptr(fd), fmt.Sprintf("/dev/fd/%d", fd))
	defer pseudo.Close()

	dir := t.TempDir()
	cmd := testCommand(t, OptionTempDir(dir))
	defer cmd.cleanup()
	got, err := cmd.input(pseudo)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(got, "/dev/") || strings.HasPrefix(got, "/proc/") {
		t.Errorf("input() = %q, want a path pdftk can open", got)
	}
	if content := readFile(t, got); content != "data" {
		t.Errorf("pdftk would read %q, want data", content)
	}
}

// Cleaning link/../x.pdf lexically gives x.pdf, while the kernel may follow
// link -> sub/deep first and open sub/x.pdf; pdftk must get what Read would.
func Test_command_input_symlinkDotDot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub", "deep"), 0o700); err != nil {
		t.Fatal(err)
	}
	for dir, content := range map[string]string{root: "root", filepath.Join(root, "sub"): "sub"} {
		if err := os.WriteFile(filepath.Join(dir, "x.pdf"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(root, "sub", "deep"), filepath.Join(root, "link")); err != nil {
		t.Skip(err)
	}
	// not filepath.Join, which would clean the name before the kernel sees it
	f := openFile(t, filepath.Join(root, "link")+"/../x.pdf")

	cmd := testCommand(t, OptionTempDir(t.TempDir()))
	defer cmd.cleanup()
	got, err := cmd.input(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	want, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if content := readFile(t, got); content != string(want) {
		t.Errorf("pdftk would read %q, want %q", content, want)
	}
	if string(want) != "sub" || len(cmd.tempFiles) != 1 {
		t.Errorf("opened %q and made %d copies, want sub copied once", want, len(cmd.tempFiles))
	}
}
