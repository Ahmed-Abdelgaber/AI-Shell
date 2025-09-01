package main

import (
	"fmt"
	"os"

	"github.com/mr-gaber/ai-shell/internal/cli"
	"github.com/mr-gaber/ai-shell/internal/shell"
	"golang.org/x/term"
)

func main() {
	// Must happen BEFORE any PTY/raw-mode/signal setup to avoid side effects.
	if handled := cli.RouteCommand(os.Args); handled {
		return
	}

	// Ensure we have a TTY (so keys reach the PTY)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr, "aish: no TTY on stdin. Run from a real terminal.")
		os.Exit(1)
	}

	// Start the shell session
	s := &shell.Session{}
	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "aish: shell session error:", err)
		os.Exit(1)
	}

}
