package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

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

	// Get or create the app data directory
	// This is where session logs and other data will be stored.
	// Default to ~/.aish
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("aish: cannot find home dir: %w", err)
	}

	// Define the app data directory path
	aishAppData := filepath.Join(homeDir, ".aish")

	// Create the app data directory if it doesn't exist
	if _, err := os.Stat(aishAppData); os.IsNotExist(err) {
		err := os.MkdirAll(aishAppData, 0o700)
		if err != nil {
			return fmt.Errorf("aish: cannot create app data dir %q: %w", aishAppData, err)
		}
	}

	// Get DataTime including timestamp for session log folder (DD/MM/YYYY_timestamp)
	// This ensures each session has a unique folder.
	currentTime := time.Now()
	timestamp := currentTime.Unix()
	currentDate := currentTime.Format("20060102")

	// Use PID to further ensure uniqueness
	pid := os.Getpid()

	// Folder name format: YYYYMMDD_timestamp_pid
	folderName := fmt.Sprintf("%s_%d_%d", currentDate, timestamp, pid)

	// Full session directory path
	sessionDir := filepath.Join(aishAppData, folderName)

	// Create the session directory
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		return fmt.Errorf("aish: cannot create session dir %q: %w", sessionDir, err)
	}

	// Create an empty history.jsonl file in the session directory
	historyPath := filepath.Join(sessionDir, "history.jsonl")
	// touch the file so it exists
	if f, err := os.OpenFile(historyPath, os.O_CREATE, 0o600); err == nil {
		_ = f.Close()
	}

	// Create an empty session.log file in the session directory
	logPath := filepath.Join(sessionDir, "session.log")

	// Build the shell command
	cmd, cleanup, err := buildShellCommand(sh, shellName, exe)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Set up environment variables for the shell session
	// These inform the shell and other components about the session context.
	// AISH_SESSION_LOG: path to the session log file
	// AISH_HISTORY_FILE: path to the history file
	cmd.Env = append(cmd.Env,
		"AISH_SESSION_LOG="+logPath,
		"AISH_HISTORY_FILE="+historyPath,
	)

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

	// Open the log file for appending
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	// Ensure the log file is closed when we're done
	defer logFile.Close()

	// Create a multi-writer to write both to stdout and the log file
	out := io.MultiWriter(os.Stdout, newCleanWriter(logFile))

	// Wire streams.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }() // user → shell
	_, _ = io.Copy(out, ptmx)                      // shell → screen

	// Propagate shell's exit code.
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		return fmt.Errorf("shell error: %w", err)
	}
	return nil
}

var ansiRE = regexp.MustCompile(`\x1B\[[0-9;?]*[ -/]*[@-~]`)

type cleanWriter struct{ w io.Writer }

func newCleanWriter(w io.Writer) *cleanWriter { return &cleanWriter{w: w} }

func (cw *cleanWriter) Write(p []byte) (int, error) {
	// work on a copy
	b := append([]byte(nil), p...)

	// normalize CRLF/CR to LF
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))

	// strip ANSI escapes
	b = ansiRE.ReplaceAll(b, nil)

	// collapse backspaces
	// (rune-safe so multi-byte chars aren’t broken)
	rs := bytes.Runes(b)
	out := make([]rune, 0, len(rs))
	for _, r := range rs {
		if r == '\b' {
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
			continue
		}
		out = append(out, r)
	}
	cleaned := []byte(string(out))

	// ensure valid UTF-8
	cleaned = bytes.ToValidUTF8(cleaned, []byte{'?'})

	// write cleaned bytes to the underlying file
	if _, err := cw.w.Write(cleaned); err != nil {
		return 0, err
	}
	// report that we consumed len(p) so io.Copy doesn't retry
	return len(p), nil
}
