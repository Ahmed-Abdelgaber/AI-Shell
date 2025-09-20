package parser

import "strings"

// Normalize splits the raw snippet text into trimmed, non-empty lines.
func Normalize(input string) []string {
	s := strings.ReplaceAll(input, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, ln := range raw {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		out = append(out, ln)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
