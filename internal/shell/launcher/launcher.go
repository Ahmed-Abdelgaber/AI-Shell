package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/shell/prompt"
	shellpty "github.com/mr-gaber/ai-shell/internal/shell/pty"
)

// Launcher prepares environment and delegates to the PTY session runner.
type Launcher struct {
	cfg config.Config
}

func New(cfg config.Config) *Launcher {
	return &Launcher{cfg: cfg}
}

func (l *Launcher) Run() error {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}

	shellName := filepath.Base(sh)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("aish: cannot find own path: %w", err)
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("aish: cannot find home dir: %w", err)
	}

	aishAppData := filepath.Join(homeDir, ".aish")
	if err := os.MkdirAll(aishAppData, 0o700); err != nil {
		return fmt.Errorf("aish: cannot create app data dir %q: %w", aishAppData, err)
	}

	currentTime := time.Now()
	timestamp := currentTime.Unix()
	currentDate := currentTime.Format("20060102")
	pid := os.Getpid()
	folderName := fmt.Sprintf("%s_%d_%d", currentDate, timestamp, pid)
	sessionDir := filepath.Join(aishAppData, folderName)

	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		return fmt.Errorf("aish: cannot create session dir %q: %w", sessionDir, err)
	}

	historyPath := filepath.Join(sessionDir, "history.jsonl")
	if f, err := os.OpenFile(historyPath, os.O_CREATE, 0o600); err == nil {
		_ = f.Close()
	}

	snippetsPath := filepath.Join(aishAppData, "snippets.yaml")
	if _, err := os.Stat(snippetsPath); os.IsNotExist(err) {
		if f, err := os.OpenFile(snippetsPath, os.O_CREATE, 0o600); err == nil {
			_ = f.Close()
		}
	} else if err != nil {
		return fmt.Errorf("aish: cannot check snippets file %q: %w", snippetsPath, err)
	}

	logPath := filepath.Join(sessionDir, "session.log")

	cmd, cleanup, err := prompt.BuildShellCommand(sh, shellName, exe)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	cmd.Env = append(cmd.Env,
		"AISH_SESSION_LOG="+logPath,
		"AISH_HISTORY_FILE="+historyPath,
		"AISH_SNIPPETS_FILE="+snippetsPath,
	)

	session := shellpty.New()
	return session.Run(cmd, logPath)
}
