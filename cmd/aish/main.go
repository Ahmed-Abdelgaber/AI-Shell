package main

import (
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
	// flag.Parse()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "__ai":
			handleAI(os.Args[2:])
			return
		case "__snip":
			handleSnip(os.Args[2:])
			return
		}
	}

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

	exe, _ := os.Executable()
	exe, _ = filepath.EvalSymlinks(exe)
	aishEnv := []string{"AISH=1", "AISH_EXE=" + exe}

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
		cmd.Env = append(os.Environ(), aishEnv...)
	case "zsh":
		zdotdir, cleanup, err := makeTempZshDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "aish: failed to make temp zsh rc: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		cmd = exec.Command(sh, "-i")
		// Point zsh to our temp .zshrc directory
		// cmd.Env = append(os.Environ(), "ZDOTDIR="+zdotdir, "AISH=1")
		cmd.Env = append(os.Environ(), append(aishEnv, "ZDOTDIR="+zdotdir)...)

	default:
		// Fallback: run shell normally; try to inject PS1 in env (some shells ignore it)
		cmd = exec.Command(sh)
		// cmd.Env = append(os.Environ(), aishEnv...)
		existingPS1 := os.Getenv("PS1")
		if existingPS1 == "" {
			existingPS1 = "\\u@\\h:\\w\\$ "
		}
		aishTag := "\\[\\e[1;36m\\][aish]\\[\\e[0m\\] "
		cmd.Env = append(os.Environ(),
			append(aishEnv, "PS1="+aishTag+existingPS1)...,
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
	go func() {
		for range ch {
			resize()
		}
	}()
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
# aish shell functions (call the parent binary)
if [ -n "$AISH_EXE" ]; then
  ai()   { "$AISH_EXE" __ai "$@"; }
  snip() { "$AISH_EXE" __snip "$@"; }
fi
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
# aish shell functions (call the parent binary)
if [[ -n "$AISH_EXE" ]]; then
  function ai()   { "$AISH_EXE" __ai "$@"; }
  function snip() { "$AISH_EXE" __snip "$@"; }
fi
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

// e.g., from bash: ai ask "..."  -> runs: aish __ai ask "..."
func handleAI(args []string) {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("ai usage:\n  ai ask <question>\n  ai why\n  ai fix")
		return
	}
	switch args[0] {
	case "ask":
		if len(args) < 2 {
			fmt.Println("usage: ai ask <question>")
			return
		}
		question := strings.Join(args[1:], " ")
		fmt.Printf("[aish] (placeholder) You asked: %q\n", question)
		fmt.Println("→ AI not wired yet; next step will call the model.")
	case "why":
		fmt.Println("[aish] (placeholder) explain last error: not implemented yet.")
	case "fix":
		fmt.Println("[aish] (placeholder) propose a fix: not implemented yet.")
	default:
		fmt.Printf("ai: unknown subcommand %q\n", args[0])
	}
}

func handleSnip(args []string) {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("snip usage:\n  snip ls\n  snip add <name> <command...>\n  snip run <name>")
		return
	}
	switch args[0] {
	case "ls":
		fmt.Println("[aish] (placeholder) no snippets yet")
	case "add":
		if len(args) < 3 {
			fmt.Println("usage: snip add <name> <command...>")
			return
		}
		name := args[1]
		cmd := strings.Join(args[2:], " ")
		fmt.Printf("[aish] (placeholder) would add snippet %q: %q\n", name, cmd)
	case "run":
		if len(args) < 2 {
			fmt.Println("usage: snip run <name>")
			return
		}
		fmt.Printf("[aish] (placeholder) would run snippet %q\n", args[1])
	default:
		fmt.Printf("snip: unknown subcommand %q\n", args[0])
	}
}
