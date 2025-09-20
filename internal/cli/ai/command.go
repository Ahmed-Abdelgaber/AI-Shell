package ai

import (
	"flag"
	"fmt"
	"io"
	"strings"

	ainternal "github.com/mr-gaber/ai-shell/internal/ai"
	"github.com/mr-gaber/ai-shell/internal/cli/shared"
	"github.com/mr-gaber/ai-shell/internal/config"
	sessionbuiltins "github.com/mr-gaber/ai-shell/internal/session/builtins"
	sessioncontext "github.com/mr-gaber/ai-shell/internal/session/context"
	sessiondanger "github.com/mr-gaber/ai-shell/internal/session/danger"
	"github.com/mr-gaber/ai-shell/internal/shell"
	"github.com/mr-gaber/ai-shell/internal/shell/runner"
)

// Handler processes ai subcommands parsed by the router.
type Handler struct {
	cfg    config.Config
	svc    *ainternal.Service
	svcErr error
}

func New(cfg config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) Handle(args []string) {
	if len(args) == 0 || args[0] == "help" {
		shared.PrintUsage(`ai usage:
  ai ask <question> // Ask a question
  ai why    // Explain the last error
  ai fix   // Propose a fix for the last error
  `)
		return
	}

	switch args[0] {
	case "ask":
		h.handleAsk(args[1:])
	case "why":
		h.handleWhy()
	case "fix":
		h.handleFix()
	default:
		fmt.Printf("ai: unknown subcommand %q\n", args[0])
	}
}

func (h *Handler) printError(err error) {
	if err == nil {
		return
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "openai_api_key") || strings.Contains(msg, "no api key") {
		fmt.Println("[aish] No AI configured. Set OPENAI_API_KEY to use 'ai' commands.")
		return
	}
	fmt.Println("ai:", err.Error())
}

func (h *Handler) handleAsk(rest []string) {
	fs := flag.NewFlagSet("ai ask", flag.ContinueOnError)
	includeContext := fs.Bool("c", false, "include recent command history as context")
	_ = fs.Bool("context", false, "include recent command history as context")
	fs.SetOutput(io.Discard)

	if len(rest) == 0 {
		fmt.Println("usage: ai ask <question>")
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

	question := strings.TrimSpace(strings.Join(rest, " "))
	if question == "" {
		fmt.Println("usage: ai ask <question>")
		return
	}

	if *includeContext {
		context, _, err := h.buildContext()
		if err != nil {
			h.printError(err)
			return
		}

		question = fmt.Sprintf("Given the following context from my recent terminal session, answer the question concisely.)\n%s\nQuestion: %s", context, question)
	}

	svc, err := h.service()
	if err != nil {
		h.printError(err)
		return
	}

	out, err := svc.Ask(question)
	if err != nil {
		h.printError(err)
		return
	}

	fmt.Println(out)
}

func (h *Handler) handleWhy() {
	context, _, err := h.buildContext()
	if err != nil {
		h.printError(err)
		return
	}

	svc, err := h.service()
	if err != nil {
		h.printError(err)
		return
	}

	out, err := svc.Why(context)
	if err != nil {
		h.printError(err)
		return
	}
	fmt.Println(out)
}

func (h *Handler) handleFix() {
	context, _, err := h.buildContext()
	if err != nil {
		h.printError(err)
		return
	}

	svc, err := h.service()
	if err != nil {
		h.printError(err)
		return
	}

	out, err := svc.Fix(context)
	if err != nil {
		h.printError(err)
		return
	}

	fmt.Println(out)

	command := strings.TrimPrefix(out, "COMMAND: ")
	command = strings.Split(command, "\n")[0]

	if strings.TrimSpace(command) == "" {
		fmt.Println("[aish] No runnable command was suggested.")
		return
	}
	if sessiondanger.IsDangerous(command) {
		fmt.Println("[aish] This command looks dangerous; refusing to auto-run.")
		return
	}
	if sessionbuiltins.IsNonPersisting(command) {
		fmt.Println("[aish] Note: builtins like 'cd'/'export' won’t change your parent shell; I won’t auto-run them.")
		return
	}

	fmt.Print("[aish] Run it now? [y/N] ")
	yes := shell.ConfirmFromStdin()
	if !yes {
		return
	}
	if err := runner.Run(command); err != nil {
		fmt.Println("[aish] run error:", err)
	}
}

func (h *Handler) buildContext() (string, bool, error) {
	builder := sessioncontext.NewBuilder(h.cfg)
	return builder.Build()
}

func (h *Handler) service() (*ainternal.Service, error) {
	if h.svc != nil {
		return h.svc, nil
	}
	if h.svcErr != nil {
		return nil, h.svcErr
	}

	svc, err := ainternal.NewService(h.cfg)
	if err != nil {
		h.svcErr = err
		return nil, err
	}
	h.svc = svc
	return svc, nil
}
