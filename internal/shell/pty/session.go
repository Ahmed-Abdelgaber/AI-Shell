package pty

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// Session wires stdio through a pseudo terminal and captures output to a log.
type Session struct{}

func New() *Session {
	return &Session{}
}

func (s *Session) Run(cmd *exec.Cmd, logPath string) error {
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("pty start error: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

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

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("raw mode error: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer logFile.Close()

	out := io.MultiWriter(os.Stdout, logFile)

	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(out, ptmx)

	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		return fmt.Errorf("shell error: %w", err)
	}
	return nil
}
