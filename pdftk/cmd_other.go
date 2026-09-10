//go:build !unix

package pdftk

import "os/exec"

func killOnCancel(*exec.Cmd) {}
