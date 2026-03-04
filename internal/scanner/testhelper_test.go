package scanner_test

import (
	"os/exec"
)

func newCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd
}
