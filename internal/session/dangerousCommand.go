package session

import "strings"

/*
DangerousCommand returns true when the user-supplied command string matches
patterns that represent obviously destructive or risky shell operations. The
function normalizes the input, scans for known hazardous substrings (e.g. rm -rf
or piping curl into sh), and treats multi-line or empty commands as dangerous so
the caller can block or warn before execution.
*/

func DangerousCommand(cmd string) bool {
	s := strings.ToLower(strings.TrimSpace(cmd))

	// obviously destructive
	if strings.Contains(s, "rm -rf /") || strings.Contains(s, "rm -rf ~") {
		return true
	}
	if strings.Contains(s, ":(){ :|:& };:") { // fork bomb
		return true
	}
	if strings.Contains(s, "mkfs") || strings.Contains(s, " dd ") || strings.Contains(s, " dd if=") || strings.Contains(s, " of=/dev/") {
		return true
	}
	if strings.Contains(s, "chmod -r 777 /") || strings.Contains(s, "chown -r ") && strings.Contains(s, " /") {
		return true
	}
	if strings.Contains(s, "curl") && strings.Contains(s, "|") && strings.Contains(s, "sh") {
		return true
	}
	if strings.Contains(s, "wget") && strings.Contains(s, "|") && strings.Contains(s, "sh") {
		return true
	}
	if strings.HasPrefix(s, "docker system prune -a") {
		return true
	}
	// multi-line or empty
	if strings.Contains(s, "\n") || s == "" {
		return true
	}
	return false
}
