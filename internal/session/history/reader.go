package history

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/utils"
)

// Reader tails the session history JSONL file.
type Reader struct {
	Path     string
	Lines    int
	MaxBytes int64
}

func (r Reader) Read() ([]string, error) {
	path := strings.TrimSpace(r.Path)
	if path == "" {
		return nil, fmt.Errorf("no session history available")
	}

	lines := r.Lines
	if lines <= 0 {
		lines = 1
	}
	maxBytes := r.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 128 << 10
	}

	historyLines, err := utils.ReadLastNLines(path, lines, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("reading session history: %w", err)
	}
	if len(historyLines) == 0 {
		return nil, fmt.Errorf("no session history available")
	}
	return historyLines, nil
}
