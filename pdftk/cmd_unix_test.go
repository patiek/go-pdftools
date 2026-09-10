//go:build unix

package pdftk

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	err := Cat(ctx, io.Discard, NewInputFileMap("first.pdf"), nil, OptionExecutable(wrapper))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed >= waitDelay {
		t.Errorf("Cat() returned after %v, want well under %v", elapsed, waitDelay)
	}
}
