package session

import (
	"regexp"
	"strings"
)

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
