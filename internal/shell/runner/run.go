package runner

import (
	"os"
	"os/exec"
)

// Run executes the provided command string inside the user's shell.
func Run(cmdline string) error {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	cmd := exec.Command(sh, "-lc", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	return cmd.Run()
}
