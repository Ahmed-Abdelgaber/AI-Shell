package cli

import (
	"fmt"
	"strings"
)

func handleAI(args []string) {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("ai usage:\n  ai ask <question>\n  ai why\n  ai fix")
		return
	}
	switch args[0] {
	case "ask":
		if len(args) < 2 {
			fmt.Println("usage: ai ask <question>")
			return
		}
		question := strings.Join(args[1:], " ")
		fmt.Printf("[aish] (placeholder) You asked: %q\n", question)
		fmt.Println("→ AI not wired yet; next step will call the model.")
	case "why":
		fmt.Println("[aish] (placeholder) explain last error: not implemented yet.")
	case "fix":
		fmt.Println("[aish] (placeholder) propose a fix: not implemented yet.")
	default:
		fmt.Printf("ai: unknown subcommand %q\n", args[0])
	}
}
