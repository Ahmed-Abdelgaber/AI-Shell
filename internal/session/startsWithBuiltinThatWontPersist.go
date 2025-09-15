package session

import "strings"

func StartsWithBuiltinThatWontPersist(cmd string) bool {
	s := strings.TrimSpace(cmd)
	// Commands that only change the current shell process
	return strings.HasPrefix(s, "cd ") || strings.HasPrefix(s, "export ") || s == "cd" || strings.HasPrefix(s, "alias ")
}
