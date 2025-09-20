package redact

import (
	"regexp"
	"strings"
)

// Scrub removes patterns resembling sensitive credentials from the provided text.
func Scrub(s string) string {
	if s == "" {
		return s
	}

	var (
		rePrivKey = regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----[.\s\S]+?-----END [A-Z ]+PRIVATE KEY-----`)
		reBearer  = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`)
		reAPIKey  = regexp.MustCompile(`(?i)(api|token|secret|key)\s*[:=]\s*["']?[A-Za-z0-9_\-\.]{12,}["']?`)
		reHexLong = regexp.MustCompile(`\b[0-9A-Fa-f]{24,}\b`)
		reB64Long = regexp.MustCompile(`\b[A-Za-z0-9+/]{32,}={0,2}\b`)
		reEnvLine = regexp.MustCompile(`(?m)^(?:[A-Z0-9_]{3,})\s*=\s*.+$`)
	)

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
