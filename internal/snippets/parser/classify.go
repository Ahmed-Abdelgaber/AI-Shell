package parser

import (
	"regexp"
	"strings"
)

var inlineEnvRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

func classify(line string) stepKind {
	s := strings.TrimSpace(line)
	if s == "" {
		return stepExec
	}
	if looksShelly(s) {
		return stepCmd
	}
	return stepExec
}

func looksShelly(s string) bool {
	if inlineEnvRE.MatchString(s) {
		return true
	}

	shellyNeedles := []string{
		"&&", "||", ">>", "2>&1",
		"|", ";", ">", "<",
		"$(", "`", "<(", ">(",
	}

	for _, n := range shellyNeedles {
		if strings.Contains(s, n) {
			return true
		}
	}

	if strings.Contains(s, "$") {
		return true
	}
	if strings.ContainsAny(s, "*?[]") {
		return true
	}
	if strings.HasPrefix(s, "~") || strings.Contains(s, " ~/") {
		return true
	}

	return false
}

func looksLikeSplitSingleCommand(ln string) bool {
	s := strings.TrimSpace(ln)
	return strings.HasSuffix(s, "\\") ||
		strings.HasSuffix(s, "|") ||
		strings.HasSuffix(s, "&&") ||
		strings.HasSuffix(s, "||")
}
