package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/ai"
	"github.com/mr-gaber/ai-shell/internal/session"
	"github.com/mr-gaber/ai-shell/internal/shell"
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
		context, _, err := session.BuildSessionContext()
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

func handleWhy() {
	// Build session context
	context, _, err := session.BuildSessionContext()
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
	context, _, err := session.BuildSessionContext()

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
	if session.DangerousCommand(command) {
		fmt.Println("[aish] This command looks dangerous; refusing to auto-run.")
		return
	}
	if session.StartsWithBuiltinThatWontPersist(command) {
		fmt.Println("[aish] Note: builtins like 'cd'/'export' won’t change your parent shell; I won’t auto-run them.")
		return
	}

	// Confirm & execute
	fmt.Print("[aish] Run it now? [y/N] ")
	yes := shell.ConfirmFromStdin()
	if !yes {
		return
	}
	if err := shell.RunInUserShell(command); err != nil {
		fmt.Println("[aish] run error:", err)
	}
}
