package session

import "strings"

func looksLikeFailure(s string) bool {
	t := strings.ToLower(s)
	keywords := []string{
		"error", "failed", "fatal", "exception", "traceback",
		"not found", "no such file or directory", "permission denied",
		"invalid operation", "cannot", "unrecognized", "unknown command",
	}
	for _, kw := range keywords {
		if strings.Contains(t, kw) {
			return true
		}
	}

	if strings.Contains(s, "\nE:") || strings.HasPrefix(s, "E:") {
		return true
	}
	return false
}
