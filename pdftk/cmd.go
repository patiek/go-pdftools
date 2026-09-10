package pdftk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// How long to wait for I/O to finish once the process exits or ctx is done.
// Keeps run from hanging if pdftk leaves a child process holding the pipes.
const waitDelay = 5 * time.Second

// The exec.Cmd is only built in run so that options can adjust the command first.
type command struct {
	ctx    context.Context
	name   string
	args   []string
	stdin  io.Reader
	stdout io.Writer
}

func createCmd(ctx context.Context, name string, stdout io.Writer, stdin io.Reader, args ...string) *command {
	return &command{
		ctx:    ctx,
		name:   name,
		args:   args,
		stdin:  stdin,
		stdout: stdout,
	}
}

func (cmd *command) applyOptions(options ...Option) {
	for _, option := range options {
		option(cmd)
	}
}

func (cmd *command) run() error {
	var stderr bytes.Buffer
	c := exec.CommandContext(cmd.ctx, cmd.name, cmd.args...)
	c.Stdin = cmd.stdin
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
