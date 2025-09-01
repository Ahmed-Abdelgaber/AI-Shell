package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type Session struct{}

func (s *Session) Run() error {
	// Get the shell path
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}

	// Get the shell name
	shellName := filepath.Base(sh)

	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("aish: cannot find own path: %w", err)
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}

	// Build the shell command
	cmd, cleanup, err := buildShellCommand(sh, shellName, exe)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Start the shell session
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("pty start error: %w", err)
	}

	defer func() { _ = ptmx.Close() }() // Close the pty

	// Resize the pty
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

	// Switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("raw mode error: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Wire streams.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }() // user → shell
	_, _ = io.Copy(os.Stdout, ptmx)                // shell → screen

	// Propagate shell's exit code.
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		return fmt.Errorf("shell error: %w", err)
	}
	return nil
}
