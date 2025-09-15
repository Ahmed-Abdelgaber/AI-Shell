package shell

import (
	"os"
	"os/exec"
)

func RunInUserShell(cmdline string) error {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	// -l would make a login shell; we only need -c and a clean command env.
	cmd := exec.Command(sh, "-lc", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	return cmd.Run()
}
