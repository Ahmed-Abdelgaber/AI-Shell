package logs

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/utils"
)

// Reader tails the session log file.
type Reader struct {
	Path     string
	Lines    int
	MaxBytes int64
}

func (r Reader) Read() ([]string, error) {
	path := strings.TrimSpace(r.Path)
	if path == "" {
		return nil, fmt.Errorf("no session log available")
	}

	lines := r.Lines
	if lines <= 0 {
		lines = 120
	}
	maxBytes := r.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 256 << 10
	}

	logLines, err := utils.ReadLastNLines(path, lines, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("reading session log: %w", err)
	}
	if len(logLines) == 0 {
		return nil, fmt.Errorf("no session log available")
	}
	return logLines, nil
}
