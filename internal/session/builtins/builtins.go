package builtins

import "strings"

// IsNonPersisting reports whether the command is a shell builtin whose effects won’t persist.
func IsNonPersisting(cmd string) bool {
	s := strings.TrimSpace(cmd)
	return strings.HasPrefix(s, "cd ") || strings.HasPrefix(s, "export ") || s == "cd" || strings.HasPrefix(s, "alias ")
}
