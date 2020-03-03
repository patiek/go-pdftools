package pdftk

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type command struct {
	*exec.Cmd
}

func createCmd(name string, stdout io.Writer, stdin io.Reader, args ...string) command {
	var stderr bytes.Buffer
	cmd := command{Cmd: exec.Command(name, args...)}
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	cmd.Stderr = &stderr
	return cmd
}

func (cmd command) applyOptions(options ...Option) {
	for _, option := range options {
		option(cmd)
	}
}

func (cmd command) runWrapError() error {
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pdftk error: %s", strings.TrimSpace(cmd.Stderr.(*bytes.Buffer).String()))
	}
	return nil
}
