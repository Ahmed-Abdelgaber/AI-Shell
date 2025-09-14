package session

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/utils"
)

type histEntry struct {
	TS   string `json:"ts"`   // Timestamp
	CWD  string `json:"cwd"`  // Current working directory
	Cmd  string `json:"cmd"`  // Command executed
	Exit int    `json:"exit"` // Exit code
	Git  string `json:"git"`  // Git branch (if any)
}

func BuildSessionContext() (string, bool, error) {
	// Get session history lines
	historyLines, err := GetSessionHistoryLines()
	if err != nil {
		return "", false, err
	}

	// Parse history lines
	history := utils.ParseJSONL[histEntry](historyLines, func(e histEntry) bool { return e.Cmd != "" })

	lastCmd, lastExit, isError := getLastCmdAndExit(history)

	if err := looksLikeFailure(lastCmd); err {
		isError = true
	}

	// If none found, report and exit
	if lastCmd == "" || lastExit == -1 {
		return "", false, fmt.Errorf("no recent non-helper command found in history")
	}

	// Build recent commands context (excluding helpers)
	var recentCmds []string
	recentN := utils.EnvIntDefault("AISH_TAIL_LINES", 5)
	for i := len(history) - 1; i >= 0 && len(recentCmds) < recentN; i-- {
		e := history[i]
		recentCmds = append(recentCmds, fmt.Sprintf("ts: %s (cmd: %s, cwd: %s, exit: %d)", e.TS, e.Cmd, e.CWD, e.Exit))
	}
	recentCmdsStr := strings.Join(recentCmds, "\n")

	logTails, err := GetSessionLogs()
	if err != nil {
		return "", false, err
	}
	logs := strings.Join(logTails, "\n")

	recentBlock := fmt.Sprintf(`Recent commands (most recent first): %s
Last command: %s
Exit code: %d	
Session logs (last %d lines):
%s
`, recentCmdsStr, lastCmd, lastExit, utils.EnvIntDefault("AISH_TAIL_LINES", 120), logs)

	return redact(recentBlock), isError, nil
}
