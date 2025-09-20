package context

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/session/history"
	"github.com/mr-gaber/ai-shell/internal/session/logs"
	"github.com/mr-gaber/ai-shell/internal/session/redact"
	"github.com/mr-gaber/ai-shell/internal/utils"
)

type Builder struct {
	history history.Reader
	logs    logs.Reader
	limits  config.Limits
}

func NewBuilder(cfg config.Config) Builder {
	return Builder{
		history: history.Reader{
			Path:     cfg.Paths.HistoryFile,
			Lines:    cfg.Limits.HistorySize * 3,
			MaxBytes: 128 << 10,
		},
		logs: logs.Reader{
			Path:     cfg.Paths.SessionLog,
			Lines:    cfg.Limits.TailLines,
			MaxBytes: int64(cfg.Limits.TailMaxBytes),
		},
		limits: cfg.Limits,
	}
}

type histEntry struct {
	TS   string `json:"ts"`
	CWD  string `json:"cwd"`
	Cmd  string `json:"cmd"`
	Exit int    `json:"exit"`
	Git  string `json:"git"`
}

func (b Builder) Build() (string, bool, error) {
	historyLines, err := b.history.Read()
	if err != nil {
		return "", false, err
	}

	historyEntries := utils.ParseJSONL(historyLines, func(e histEntry) bool { return e.Cmd != "" })

	lastCmd, lastExit, isError := getLastCmdAndExit(historyEntries)
	if looksLikeFailure(lastCmd) {
		isError = true
	}

	if lastCmd == "" || lastExit == -1 {
		return "", false, fmt.Errorf("no recent non-helper command found in history")
	}

	recentCmds := make([]string, 0, b.limits.TailLines)
	for i := len(historyEntries) - 1; i >= 0 && len(recentCmds) < b.limits.TailLines; i-- {
		e := historyEntries[i]
		recentCmds = append(recentCmds, fmt.Sprintf("ts: %s (cmd: %s, cwd: %s, exit: %d)", e.TS, e.Cmd, e.CWD, e.Exit))
	}
	recentCmdsStr := strings.Join(recentCmds, "\n")

	logLines, err := b.logs.Read()
	if err != nil {
		return "", false, err
	}
	logsStr := strings.Join(logLines, "\n")

	block := fmt.Sprintf(`Recent commands (most recent first): %s
Last command: %s
Exit code: %d
Session logs (last %d lines):
%s
`, recentCmdsStr, lastCmd, lastExit, b.limits.TailLines, logsStr)

	return redact.Scrub(block), isError, nil
}

func getLastCmdAndExit(history []histEntry) (string, int, bool) {
	var lastCmd string
	lastExit := -1
	isError := false

	for i := len(history) - 1; i >= 0; i-- {
		if !isHelper(history[i].Cmd) {
			lastCmd = history[i].Cmd
			lastExit = history[i].Exit
			isError = lastExit != 0
			break
		}
	}
	return lastCmd, lastExit, isError
}

func isHelper(cmd string) bool {
	t := strings.TrimSpace(cmd)
	return strings.HasPrefix(t, "ai ") || strings.HasPrefix(t, "snip ")
}
