package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/ai"
	"github.com/mr-gaber/ai-shell/internal/session"
)

func handleAI(args []string) {
	if len(args) == 0 || args[0] == "help" {
		printAIUsage()
		return
	}
	switch args[0] {
	case "ask":
		handleAsk(args[1:])
	case "why":
		handleWhy(args[1:])
	case "fix":
		handleFix(args[1:])
	default:
		fmt.Printf("ai: unknown subcommand %q\n", args[0])
	}
}

func printAIUsage() {
	fmt.Println(`ai usage:
  ai ask <question> // Ask a question
  ai why    // Explain the last error
  ai fix   // Propose a fix for the last error
  `)
}

func printAIError(err error) {
	// Friendly message if no key or provider setup issue
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "openai_api_key") || strings.Contains(msg, "no api key") {
		fmt.Println("[aish] No AI configured. Set OPENAI_API_KEY to use 'ai' commands.")
		return
	}
	fmt.Println("ai:", err.Error())
}

func handleAsk(rest []string) {
	if len(rest) == 0 {
		fmt.Println("usage: ai ask <question>")
		return
	}
	question := strings.TrimSpace(strings.Join(rest, " "))
	if question == "" {
		fmt.Println("usage: ai ask <question>")
		return
	}

	out, err := ai.Ask(question)
	if err != nil {
		printAIError(err)
		return
	}
	fmt.Println(out)
}

// histEntry is a single entry in the session history log
type histEntry struct {
	TS   string `json:"ts"`   // Timestamp
	CWD  string `json:"cwd"`  // Current working directory
	Cmd  string `json:"cmd"`  // Command executed
	Exit int    `json:"exit"` // Exit code
	Git  string `json:"git"`  // Git branch (if any)
}

func parseHistoryJSONL(lines []string) []histEntry {
	out := make([]histEntry, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		var e histEntry
		if err := json.Unmarshal([]byte(ln), &e); err != nil && e.Cmd != "" {
			continue
		}
		out = append(out, e)
	}
	return out
}

func isHelper(cmd string) bool {
	t := strings.TrimSpace(cmd)
	return strings.HasPrefix(t, "ai ") || strings.HasPrefix(t, "snip ")
}

func envIntDefault(name string, def int) int {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getSessionHistoryLines() ([]string, error) {
	// Get history file from env
	historyLinesReference := strings.TrimSpace(os.Getenv("AISH_HISTORY_FILE"))
	if historyLinesReference == "" {
		return nil, fmt.Errorf("no session history available")
	}

	// Read last N lines from history file
	historyLines, err := session.ReadLastNLines(historyLinesReference, envIntDefault("AISH_HISTORY_SIZE", 5)*3, 128<<10)
	if err != nil {
		return nil, fmt.Errorf("reading session history: %w", err)
	}
	if len(historyLines) == 0 {
		return nil, fmt.Errorf("no session history available")
	}
	return historyLines, nil
}

func getSessionLogs() ([]string, error) {
	// Get session log file from env
	sessionLogFile := strings.TrimSpace(os.Getenv("AISH_SESSION_LOG"))
	if sessionLogFile == "" {
		return nil, fmt.Errorf("no session log available")
	}

	tailLines := envIntDefault("AISH_TAIL_LINES", 120)
	tailBytes := envIntDefault("AISH_TAIL_MAX_BYTES", 256<<10)

	logTails, err := session.ReadLastNLines(sessionLogFile, tailLines, int64(tailBytes))
	if err != nil {
		return nil, fmt.Errorf("reading session log: %w", err)
	}
	if len(logTails) == 0 {
		return nil, fmt.Errorf("no session log available")
	}
	return logTails, nil
}

func buildSessionContext() (string, error) {
	// Get session history lines
	historyLines, err := getSessionHistoryLines()
	if err != nil {
		return "", err
	}

	// Parse history lines
	history := parseHistoryJSONL(historyLines)

	// Find last non-helper command
	var lastCmd string = ""
	var lastExit = -1
	// Scan backwards
	for i := len(history) - 1; i >= 0; i-- {
		if !isHelper(history[i].Cmd) {
			lastCmd = history[i].Cmd
			lastExit = history[i].Exit
			break
		}
	}
	// If none found, report and exit
	if lastCmd == "" || lastExit == -1 {
		return "", fmt.Errorf("no recent non-helper command found in history")
	}

	// Build recent commands context (excluding helpers)
	var recentCmds []string
	recentN := envIntDefault("AISH_TAIL_LINES", 5)
	for i := len(history) - 1; i >= 0 && len(recentCmds) < recentN; i-- {
		e := history[i]
		recentCmds = append(recentCmds, fmt.Sprintf("ts: %s (cmd: %s, cwd: %s, exit: %d)", e.TS, e.Cmd, e.CWD, e.Exit))
	}
	recentCmdsStr := strings.Join(recentCmds, "\n")

	logTails, err := getSessionLogs()
	if err != nil {
		return "", err
	}
	logs := strings.Join(logTails, "\n")

	recentBlock := fmt.Sprintf(`Recent commands (most recent first): %s
Last command: %s
Exit code: %d	
Session logs (last %d lines):
%s
`, recentCmdsStr, lastCmd, lastExit, envIntDefault("AISH_TAIL_LINES", 120), logs)

	return redact(recentBlock), nil
}

// redact attempts to remove sensitive information from a string.
func redact(s string) string {
	var (
		// Regular expressions to match sensitive information patterns.
		// These patterns include private keys, bearer tokens, API keys, long hex strings,
		rePrivKey = regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----[.\s\S]+?-----END [A-Z ]+PRIVATE KEY-----`)
		// long base64 strings, and environment variable assignments.
		reBearer = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]+=*`)
		// The patterns are designed to catch common formats of sensitive data.
		reAPIKey = regexp.MustCompile(`(?i)(api|token|secret|key)\s*[:=]\s*["']?[A-Za-z0-9_\-\.]{12,}["']?`)
		// Long hex strings (24 or more hex characters)
		reHexLong = regexp.MustCompile(`\b[0-9A-Fa-f]{24,}\b`)
		// Long base64 strings (32 or more base64 characters)
		reB64Long = regexp.MustCompile(`\b[A-Za-z0-9+/]{32,}={0,2}\b`)
		// Environment variable assignment lines (e.g., VAR=value)
		reEnvLine = regexp.MustCompile(`(?m)^(?:[A-Z0-9_]{3,})\s*=\s*.+$`)
	)

	if s == "" {
		return s
	}
	out := s
	out = rePrivKey.ReplaceAllString(out, "[REDACTED_PRIVATE_KEY]")
	out = reBearer.ReplaceAllString(out, "Bearer [REDACTED]")
	out = reAPIKey.ReplaceAllString(out, "$1: [REDACTED]")
	out = reHexLong.ReplaceAllString(out, "[HEX_REDACTED]")
	out = reB64Long.ReplaceAllString(out, "[B64_REDACTED]")
	out = reEnvLine.ReplaceAllStringFunc(out, func(line string) string {
		if i := strings.Index(line, "="); i > 0 {
			return line[:i+1] + " [REDACTED]"
		}
		return line
	})
	return out
}

func handleWhy(rest []string) {
	if len(rest) == 0 {
		printAIUsage()
		return
	}

	// Build session context
	context, err := buildSessionContext()
	if err != nil {
		printAIError(err)
		return
	}

	out, err := ai.Why(context)
	if err != nil {
		printAIError(err)
		return
	}
	fmt.Println(out)
}

func handleFix(rest []string) {
	if len(rest) == 0 {
		printAIUsage()
		return
	}

	// Build session context
	context, err := buildSessionContext()
	if err != nil {
		printAIError(err)
		return
	}

	out, err := ai.Fix(context)
	if err != nil {
		printAIError(err)
		return
	}

	fmt.Println(out)

	// Parse command from output
	// Expecting format:
	// COMMAND: <single-line>
	// WHY: <one sentence>
	// We’ll extract the command after "COMMAND: " and before the next newline.
	// If the format is not as expected, we’ll handle it gracefully.
	command := strings.TrimPrefix(out, "COMMAND: ")
	command = strings.Split(command, "\n")[0]

	if strings.TrimSpace(command) == "" {
		fmt.Println("[aish] No runnable command was suggested.")
		return
	}
	if dangerousCommand(command) {
		fmt.Println("[aish] This command looks dangerous; refusing to auto-run.")
		return
	}
	if startsWithBuiltinThatWontPersist(command) {
		fmt.Println("[aish] Note: builtins like 'cd'/'export' won’t change your parent shell; I won’t auto-run them.")
		return
	}

	// Confirm & execute
	fmt.Print("[aish] Run it now? [y/N] ")
	yes := confirmFromStdin()
	if !yes {
		return
	}
	if err := runInUserShell(command); err != nil {
		fmt.Println("[aish] run error:", err)
	}
}

func confirmFromStdin() bool {
	in := bufio.NewReader(os.Stdin)
	line, _ := in.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

func runInUserShell(cmdline string) error {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	// -l would make a login shell; we only need -c and a clean command env.
	cmd := exec.Command(sh, "-lc", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	return cmd.Run()
}

// Conservative danger checks (v1). We’ll expand over time.
func dangerousCommand(cmd string) bool {
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

func startsWithBuiltinThatWontPersist(cmd string) bool {
	s := strings.TrimSpace(cmd)
	// Commands that only change the current shell process
	return strings.HasPrefix(s, "cd ") || strings.HasPrefix(s, "export ") || s == "cd" || strings.HasPrefix(s, "alias ")
}
