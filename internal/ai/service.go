package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func Ask(question string) (answer string, err error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH, a terse terminal assistant. Prefer one good command with a one-line explanation. Be concise."
	user := strings.TrimSpace(question)
	// Initialize the AI provider.
	aiProvider, err := NewAIProvider()
	if err != nil {
		return "", err
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

func Why(lastCmd string, lastExit int, tail string) (string, error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH. Propose ONE safe fix command and a one-sentence rationale. Output strictly in the following format:\nCOMMAND: <single-line>\nWHY: <one sentence>"
	user := fmt.Sprintf(
		"Last command: %q\nExit code: %d\nRecent output:\n%s",
		strings.TrimSpace(lastCmd), lastExit, redact(strings.TrimSpace(tail)),
	)
	// Initialize the AI provider.
	aiProvider, err := NewAIProvider()
	if err != nil {
		return "", err
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

func Fix(lastCmd string, lastExit int, tail string) (string, error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH. Propose ONE safe fix command and a one-sentence rationale. Output strictly in the following format:\nCOMMAND: <single-line>\nWHY: <one sentence>"
	user := fmt.Sprintf(
		"Last command: %q\nExit code: %d\nRecent output:\n%s",
		strings.TrimSpace(lastCmd), lastExit, redact(strings.TrimSpace(tail)),
	)
	// Initialize the AI provider.
	aiProvider, err := NewAIProvider()
	if err != nil {
		return "", err
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

// redact attempts to remove sensitive information from a string.
func redact(s string) string {
	var (
		// Regular expressions to match sensitive information patterns.
		// These patterns include private keys, bearer tokens, API keys, long hex strings,
		rePrivKey = regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----[.\s\S]+?-----END [A-Z ]+PRIVATE KEY-----`)
		// long base64 strings, and environment variable assignments.
		reBearer = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`)
		// The patterns are designed to catch common formats of sensitive data.
		reAPIKey = regexp.MustCompile(`(?i)(api|token|secret|key)\s*[:=]\s*["']?[A-Za-z0-9_\-\.]{12,}["']?`)
		// Long hex strings (24 or more hex characters)
		reHexLong = regexp.MustCompile(`\b[0-9A-Fa-f]{24,}\b`)
		// Long base64 strings (32 or more base64 characters)
		reB64Long = regexp.MustCompile(`\b[A-Za-z0-9+/]{32,}={0,2}\b`)
		// Environment variable assignment lines (e.g., VAR=value)
		reEnvLine = regexp.MustCompile(`(?m)^(?:[A-Z0-9_]{3,})\s*=\s*.+$`)
	)

	if s == "" {
		return s
	}
	out := s
	out = rePrivKey.ReplaceAllString(out, "[REDACTED_PRIVATE_KEY]")
	out = reBearer.ReplaceAllString(out, "Bearer [REDACTED]")
	out = reAPIKey.ReplaceAllString(out, "$1: [REDACTED]")
	out = reHexLong.ReplaceAllString(out, "[HEX_REDACTED]")
	out = reB64Long.ReplaceAllString(out, "[B64_REDACTED]")
	out = reEnvLine.ReplaceAllStringFunc(out, func(line string) string {
		if i := strings.Index(line, "="); i > 0 {
			return line[:i+1] + " [REDACTED]"
		}
		return line
	})
	return out
}
