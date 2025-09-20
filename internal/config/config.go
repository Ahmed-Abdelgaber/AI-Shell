package config

import (
	"os"
	"strings"
)

// Config captures environment-driven settings and derived defaults used across the app.
type Config struct {
	Paths      Paths
	Limits     Limits
	AIProvider string
	OpenAIKey  string
}

// Paths groups filesystem locations discovered from the environment.
type Paths struct {
	SnippetsFile string
	SessionLog   string
	HistoryFile  string
}

// Limits collects numeric tuning knobs sourced from env vars.
type Limits struct {
	TailLines    int
	TailMaxBytes int
	HistorySize  int
}

// LoadFromEnv constructs a Config populated from environment variables, applying defaults.
func LoadFromEnv() Config {
	provider := strings.TrimSpace(os.Getenv("AI_PROVIDER"))
	if provider == "" {
		provider = "openai"
	}

	return Config{
		Paths: Paths{
			SnippetsFile: strings.TrimSpace(os.Getenv("AISH_SNIPPETS_FILE")),
			SessionLog:   strings.TrimSpace(os.Getenv("AISH_SESSION_LOG")),
			HistoryFile:  strings.TrimSpace(os.Getenv("AISH_HISTORY_FILE")),
		},
		Limits: Limits{
			TailLines:    intDefault("AISH_TAIL_LINES", 120),
			TailMaxBytes: intDefault("AISH_TAIL_MAX_BYTES", 256<<10),
			HistorySize:  intDefault("AISH_HISTORY_SIZE", 5),
		},
		AIProvider: provider,
		OpenAIKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
	}
}
