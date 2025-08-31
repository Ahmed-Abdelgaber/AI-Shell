package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func main() {
	// Flags kept minimal for now
	flag.Parse()

	// Ensure we have a TTY (so keys reach the PTY)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr, "aish: no TTY on stdin. Run from a real terminal.")
		os.Exit(1)
	}

	// --- Choose shell ---
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	shellName := filepath.Base(sh)

	// Prepare the command we will run (we may adjust args/env below)
	var cmd *exec.Cmd
	switch shellName {
	case "bash":
		rcPath, cleanup, err := makeTempBashRC()
		if err != nil {
			fmt.Fprintf(os.Stderr, "aish: failed to make temp bash rc: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		// Use --rcfile so bash reads our wrapper (interactive)
		cmd = exec.Command(sh, "--rcfile", rcPath, "-i")

	case "zsh":
		zdotdir, cleanup, err := makeTempZshDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "aish: failed to make temp zsh rc: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		cmd = exec.Command(sh, "-i")
		// Point zsh to our temp .zshrc directory
		cmd.Env = append(os.Environ(), "ZDOTDIR="+zdotdir, "AISH=1")

	default:
		// Fallback: run shell normally; try to inject PS1 in env (some shells ignore it)
		cmd = exec.Command(sh)
		existingPS1 := os.Getenv("PS1")
		if existingPS1 == "" {
			existingPS1 = "\\u@\\h:\\w\\$ "
		}
		aishTag := "\\[\\e[1;36m\\][aish]\\[\\e[0m\\] "
		cmd.Env = append(os.Environ(),
			"AISH=1",
			"PS1="+aishTag+existingPS1,
		)
	}

	// --- Start the shell in a new PTY ---
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pty start error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = ptmx.Close() }()

	// Window-resize propagation
	resize := func() { _ = pty.InheritSize(os.Stdin, ptmx) }
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() { for range ch { resize() } }()
	resize()
	defer signal.Stop(ch)

	// Raw mode for stdin
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "raw mode error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Wire I/O
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	// Exit with the shell's code
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "shell error: %v\n", err)
		os.Exit(1)
	}
}

// --- Helpers ---

// makeTempBashRC writes a temporary rcfile that sources the user's ~/.bashrc (if present)
// and then prepends a colored [aish] to PS1. Returns (path, cleanup, error).
func makeTempBashRC() (string, func(), error) {
	home, _ := os.UserHomeDir()
	userRC := filepath.Join(home, ".bashrc")

	content := `# aish: temporary bash rc (auto-generated)
# Source the user's bashrc if it exists
if [ -f "` + escapeShell(userRC) + `" ]; then
  # shellcheck disable=SC1090
  . "` + escapeShell(userRC) + `"
fi

export AISH=1
# If PS1 is empty, set a reasonable default
if [ -z "$PS1" ]; then
  PS1="\u@\h:\w\$ "
fi
# Prepend a bold-cyan [aish] tag; \[ \] mark zero-width for correct wrapping
PS1='\[\e[1;36m\][aish]\[\e[0m\] '"$PS1"
`

	f, err := os.CreateTemp("", "aish-bashrc-*")
	if err != nil {
		return "", nil, err
	}
	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", nil, err
	}
	cleanup := func() { _ = os.Remove(f.Name()) }
	return f.Name(), cleanup, nil
}

// makeTempZshDir writes a temporary ZDOTDIR with a .zshrc that sources user's ~/.zshrc,
// then prepends a colored [aish] tag to PROMPT.
func makeTempZshDir() (string, func(), error) {
	home, _ := os.UserHomeDir()
	userRC := filepath.Join(home, ".zshrc")

	dir, err := os.MkdirTemp("", "aish-zdotdir-*")
	if err != nil {
		return "", nil, err
	}
	rcPath := filepath.Join(dir, ".zshrc")

	content := `# aish: temporary zsh rc (auto-generated)
# Source the user's zshrc if it exists
if [ -f "` + escapeShell(userRC) + `" ]; then
  source "` + escapeShell(userRC) + `"
fi

export AISH=1
# Prepend bold-cyan [aish] to the left prompt
PROMPT="%B%F{cyan}[aish]%f%b ${PROMPT}"
`

	if err := os.WriteFile(rcPath, []byte(content), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	return dir, cleanup, nil
}

// escapeShell is a minimal path escaper for the rc snippets.
func escapeShell(p string) string {
	// Wrap in double-quotes in the script; here we only need to escape embedded double-quotes.
	return strings.ReplaceAll(p, `"`, `\"`)
}
