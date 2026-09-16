package pdftk

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"
)

// How long to wait for I/O to finish once the process exits or ctx is done.
// Keeps run from hanging if pdftk leaves a child process holding the pipes.
const waitDelay = 5 * time.Second

var errNilInput = errors.New("pdftk error: nil input")

// Options are applied before inputs are resolved and the exec.Cmd is only
// built in run.
type command struct {
	ctx     context.Context
	name    string
	tempDir string
	stdout  io.Writer
	// output options such as flatten, which pdftk expects last
	outputArgs []string
	// resolved so far, so a comparable reader given twice is copied once
	inputs []resolvedInput
	// copies of inputs, removed by cleanup
	tempFiles []string
}

type resolvedInput struct {
	r    io.Reader
	path string
}

// Fails on a missing executable before any input is copied.
func newCommand(ctx context.Context, stdout io.Writer, options []Option) (*command, error) {
	cmd := &command{ctx: ctx, name: cmdPDFtk, stdout: stdout}
	for _, option := range options {
		option(cmd)
	}
	path, err := exec.LookPath(cmd.name)
	if err != nil {
		return nil, fmt.Errorf("pdftk error: %w", err)
	}
	cmd.name = path
	// a relative temp dir would give bare names pdftk may misparse and that
	// a later change of directory would leave behind
	if cmd.tempDir, err = filepath.Abs(cmp.Or(cmd.tempDir, os.TempDir())); err != nil {
		return nil, fmt.Errorf("pdftk error: %w", err)
	}
	return cmd, nil
}

// Argument that makes pdftk read r: the path of an unread regular file, else
// the path of a temp copy, as pdftk needs seekable files and holds anything
// streamed on stdin in memory.
func (cmd *command) input(r io.Reader) (string, error) {
	if r == nil {
		return "", errNilInput
	}
	comparable := reflect.ValueOf(r).Comparable()
	if comparable {
		for _, in := range cmd.inputs {
			if in.r == r {
				return in.path, nil
			}
		}
	}
	path, ok := filePath(r)
	if !ok {
		var err error
		if path, err = cmd.copyToTempFile(r); err != nil {
			return "", err
		}
	}
	if comparable {
		cmd.inputs = append(cmd.inputs, resolvedInput{r, path})
	}
	return path, nil
}

// handle=input arguments in handle order, checked up front so an invalid
// call copies nothing.
func (cmd *command) inputArgs(inputs InputMap) ([]string, error) {
	handles := inputs.handles()
	for _, handle := range handles {
		if inputs[handle] == nil {
			return nil, fmt.Errorf("%w %s", errNilInput, handle)
		}
	}
	args := make([]string, len(handles))
	for i, handle := range handles {
		arg, err := cmd.input(inputs[handle])
		if err != nil {
			return nil, err
		}
		args[i] = handle + "=" + arg
	}
	return args, nil
}

// Path pdftk can open to get what Read would return: a regular file with
// nothing read yet that is still present under its resolved name and not a
// per-process pseudo file such as /dev/fd/0.
func filePath(r io.Reader) (string, bool) {
	f, ok := r.(*os.File)
	if !ok {
		return "", false
	}
	fi, err := f.Stat()
	if err != nil || !fi.Mode().IsRegular() {
		return "", false
	}
	if offset, err := f.Seek(0, io.SeekCurrent); err != nil || offset != 0 {
		return "", false
	}
	// pdftk mistakes bare names such as "cat" for keywords
	path, err := filepath.Abs(f.Name())
	if err != nil {
		return "", false
	}
	if path, err = filepath.EvalSymlinks(path); err != nil {
		return "", false
	}
	if strings.HasPrefix(path, "/dev/fd/") || strings.HasPrefix(path, "/proc/") {
		return "", false
	}
	named, err := os.Stat(path)
	if err != nil || !os.SameFile(fi, named) {
		return "", false
	}
	return path, true
}

func (cmd *command) copyToTempFile(r io.Reader) (string, error) {
	f, err := os.CreateTemp(cmd.tempDir, "go-pdftools-*")
	if err != nil {
		return "", fmt.Errorf("pdftk error: %w", err)
	}
	cmd.tempFiles = append(cmd.tempFiles, f.Name())
	_, err = io.Copy(f, contextReader{cmd.ctx, r})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", fmt.Errorf("pdftk error: copying input: %w", err)
	}
	return f.Name(), nil
}

// Best effort; pdftk has exited by the time this runs.
func (cmd *command) cleanup() {
	for _, name := range cmd.tempFiles {
		os.Remove(name)
	}
}

// Stops a copy between Reads once ctx is done.
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr contextReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}

func (cmd *command) run(args ...string) error {
	var stderr bytes.Buffer
	c := exec.CommandContext(cmd.ctx, cmd.name, slices.Concat(args, cmd.outputArgs)...)
	c.Stdout = cmd.stdout
	c.Stderr = &stderr
	c.WaitDelay = waitDelay
	killOnCancel(c)
	if err := c.Run(); err != nil {
		// report cancellation and timeouts so callers can check for them
		if ctxErr := cmd.ctx.Err(); ctxErr != nil && !errors.Is(err, ctxErr) {
			err = fmt.Errorf("%w: %w", ctxErr, err)
		}
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("pdftk error: %s: %w", msg, err)
		}
		return fmt.Errorf("pdftk error: %w", err)
	}
	return nil
}
