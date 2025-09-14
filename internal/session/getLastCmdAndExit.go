package session

import "strings"

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

func isHelper(cmd string) bool {
	t := strings.TrimSpace(cmd)
	return strings.HasPrefix(t, "ai ") || strings.HasPrefix(t, "snip ")
}
