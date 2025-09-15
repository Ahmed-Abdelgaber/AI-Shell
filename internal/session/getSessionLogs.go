package session

import (
	"fmt"
	"os"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/utils"
)

func GetSessionLogs() ([]string, error) {
	// Get session log file from env
	sessionLogFile := strings.TrimSpace(os.Getenv("AISH_SESSION_LOG"))
	if sessionLogFile == "" {
		return nil, fmt.Errorf("no session log available")
	}

	tailLines := utils.EnvIntDefault("AISH_TAIL_LINES", 120)
	tailBytes := utils.EnvIntDefault("AISH_TAIL_MAX_BYTES", 256<<10)

	logTails, err := utils.ReadLastNLines(sessionLogFile, tailLines, int64(tailBytes))
	if err != nil {
		return nil, fmt.Errorf("reading session log: %w", err)
	}
	if len(logTails) == 0 {
		return nil, fmt.Errorf("no session log available")
	}
	return logTails, nil
}
