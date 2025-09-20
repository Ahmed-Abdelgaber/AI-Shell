package danger

import "strings"

// IsDangerous reports whether the command string matches well-known destructive patterns.
func IsDangerous(cmd string) bool {
	s := strings.ToLower(strings.TrimSpace(cmd))

	if strings.Contains(s, "rm -rf /") || strings.Contains(s, "rm -rf ~") {
		return true
	}
	if strings.Contains(s, ":(){ :|:& };:") {
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
	if strings.Contains(s, "\n") || s == "" {
		return true
	}
	return false
}
