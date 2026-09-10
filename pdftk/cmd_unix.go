//go:build unix

package pdftk

import (
	"os/exec"
	"syscall"
)

// Run pdftk in its own process group so cancelling also kills the java process
// started by wrapper scripts that do not exec (e.g. Debian's pdftk-java).
func killOnCancel(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error {
		return syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
	}
}
