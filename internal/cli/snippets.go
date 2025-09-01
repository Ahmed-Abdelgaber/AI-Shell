package cli

import (
	"fmt"
	"strings"
)

func handleSnip(args []string) {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("snip usage:\n  snip ls\n  snip add <name> <command...>\n  snip run <name>")
		return
	}
	switch args[0] {
	case "ls":
		fmt.Println("[aish] (placeholder) no snippets yet")
	case "add":
		if len(args) < 3 {
			fmt.Println("usage: snip add <name> <command...>")
			return
		}
		name := args[1]
		cmd := strings.Join(args[2:], " ")
		fmt.Printf("[aish] (placeholder) would add snippet %q: %q\n", name, cmd)
	case "run":
		if len(args) < 2 {
			fmt.Println("usage: snip run <name>")
			return
		}
		fmt.Printf("[aish] (placeholder) would run snippet %q\n", args[1])
	default:
		fmt.Printf("snip: unknown subcommand %q\n", args[0])
	}
}
