package cli

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/ai"
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
		fmt.Println("[aish] (placeholder) explain last error: not implemented yet.")
	case "fix":
		fmt.Println("[aish] (placeholder) propose a fix: not implemented yet.")
	default:
		fmt.Printf("ai: unknown subcommand %q\n", args[0])
	}
}

func printAIUsage() {
	fmt.Println(`ai usage:
  ai ask <question>
  ai why 
  ai fix 

Environment fallback (if flags omitted):
  AISH_LAST_CMD, AISH_LAST_EXIT, AISH_SESSION_LOG, AISH_TAIL_LINES, AISH_TAIL_MAX_BYTES`)
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

func printAIError(err error) {
	// Friendly message if no key or provider setup issue
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "openai_api_key") || strings.Contains(msg, "no api key") {
		fmt.Println("[aish] No AI configured. Set OPENAI_API_KEY to use 'ai' commands.")
		return
	}
	fmt.Println("ai:", err.Error())
}
