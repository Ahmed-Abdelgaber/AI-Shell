package session

import (
	"fmt"
	"os"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/utils"
)

func GetSessionHistoryLines() ([]string, error) {
	// Get history file from env
	historyLinesReference := strings.TrimSpace(os.Getenv("AISH_HISTORY_FILE"))
	if historyLinesReference == "" {
		return nil, fmt.Errorf("no session history available")
	}

	// Read last N lines from history file
	historyLines, err := utils.ReadLastNLines(historyLinesReference, utils.EnvIntDefault("AISH_HISTORY_SIZE", 5)*3, 128<<10)
	if err != nil {
		return nil, fmt.Errorf("reading session history: %w", err)
	}
	if len(historyLines) == 0 {
		return nil, fmt.Errorf("no session history available")
	}
	return historyLines, nil
}
