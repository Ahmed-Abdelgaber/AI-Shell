package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
		handleWhy()
	case "fix":
		handleFix()
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
	// Create a flag set for parsing options
	fs := flag.NewFlagSet("ai ask", flag.ContinueOnError)
	// Define flags
	includeContext := fs.Bool("c", false, "include recent command history as context")
	_ = fs.Bool("context", false, "include recent command history as context")
	// Suppress flag parsing errors to avoid cluttering output
	fs.SetOutput(io.Discard)
	// Parse flags
	if len(rest) == 0 {
		fmt.Println("usage: ai ask <question>")
		return
	}
	if err := fs.Parse(rest); err != nil {
		fmt.Println("usage: ai ask [-c|--context] <question>")
		return
	}
	if err := fs.Parse(rest); err != nil {
		fmt.Println("usage: ai ask [-c|--context] <question>")
		return
	}
	args := fs.Args()
	if len(args) == 0 {
		fmt.Println("usage: ai ask [-c|--context] <question>")
		return
	}

	// Remaining args are the question
	question := strings.TrimSpace(strings.Join(rest, " "))
	if question == "" {
		fmt.Println("usage: ai ask <question>")
		return
	}

	// If -c/--context is set, build session context
	if *includeContext {
		context, _, err := buildSessionContext()
		if err != nil {
			printAIError(err)
			return
		}

		// Combine context and question
		// We’ll format it as: "Given the following context from my recent terminal session, answer the question concisely.\n<context>\nQuestion: <question>"
		// This helps the AI understand the context in which the question is asked.

		question = fmt.Sprintf(`Given the following context from my recent terminal session, answer the question concisely.) \n %s \n Question: %s`, context, question)

	}

	// Ask the question using the AI service
	out, err := ai.Ask(question)
	if err != nil {
		printAIError(err)
		return
	}

	// Print the AI's response
	fmt.Println(out)
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

func handleWhy() {
	// Build session context
	context, _, err := buildSessionContext()
	// if !isError {
	// 	fmt.Println("[aish] Last command was not an error; nothing to explain.")
	// 	return
	// }
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

func handleFix() {

	// Build session context
	context, _, err := buildSessionContext()

	// if !isError {
	// 	fmt.Println("[aish] Last command was not an error; nothing to fix.")
	// 	return
	// }
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
