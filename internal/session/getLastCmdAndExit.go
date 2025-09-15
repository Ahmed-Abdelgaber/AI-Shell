package session

import "strings"

/*
getLastCmdAndExit walks backwards through the provided history entries to find
the most recent command that is expected to persist across shell sessions. It
returns the command string, its exit code, and whether that exit code represents
an error, allowing callers to understand recent shell state.
*/
func getLastCmdAndExit(history []histEntry) (string, int, bool) {
	// Find last non-helper command
	var lastCmd string = ""
	var lastExit = -1
	var isError bool = false
	// Scan backwards
	for i := len(history) - 1; i >= 0; i-- {
		if !isHelper(history[i].Cmd) {
			lastCmd = history[i].Cmd
			lastExit = history[i].Exit
			isError = lastExit != 0
			break
		}
	}
	return lastCmd, lastExit, isError
}

/*
isHelper reports whether the command is an internal helper (ai or snip) that
should be ignored when determining the last user command.
*/
func isHelper(cmd string) bool {
	t := strings.TrimSpace(cmd)
	return strings.HasPrefix(t, "ai ") || strings.HasPrefix(t, "snip ")
}
