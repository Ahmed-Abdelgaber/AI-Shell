package session

import "strings"

/*
StartsWithBuiltinThatWontPersist reports whether the command begins with a
builtin that only affects the current shell process. The caller can use this to
avoid recording or replaying commands like cd, export, or alias that have no
lasting effect.
*/
func StartsWithBuiltinThatWontPersist(cmd string) bool {
	s := strings.TrimSpace(cmd)
	// Commands that only change the current shell process
	return strings.HasPrefix(s, "cd ") || strings.HasPrefix(s, "export ") || s == "cd" || strings.HasPrefix(s, "alias ")
}
