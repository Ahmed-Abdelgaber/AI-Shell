package ai

import (
	"flag"
	"fmt"
	"io"
	"strings"

	ainternal "github.com/mr-gaber/ai-shell/internal/ai"
	"github.com/mr-gaber/ai-shell/internal/cli/shared"
	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/errs"
	sessionbuiltins "github.com/mr-gaber/ai-shell/internal/session/builtins"
	sessioncontext "github.com/mr-gaber/ai-shell/internal/session/context"
	sessiondanger "github.com/mr-gaber/ai-shell/internal/session/danger"
	"github.com/mr-gaber/ai-shell/internal/shell"
	"github.com/mr-gaber/ai-shell/internal/shell/runner"
	"github.com/mr-gaber/ai-shell/internal/ux/printer"
)

// Handler processes ai subcommands parsed by the router.
type Handler struct {
	cfg     config.Config
	svc     *ainternal.Service
	svcErr  error
	printer *printer.Printer
}

func New(cfg config.Config, p *printer.Printer) *Handler {
	return &Handler{cfg: cfg, printer: p}
}

func (h *Handler) Handle(args []string) {
	if len(args) == 0 || args[0] == "help" {
		shared.PrintUsage(h.printer, `ai usage:
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
		if h.printer != nil {
			h.printer.Warn("[aish] No AI configured. Set OPENAI_API_KEY to use 'ai' commands.")
		} else {
			fmt.Println("[aish] No AI configured. Set OPENAI_API_KEY to use 'ai' commands.")
		}
		return
	}
	if h.printer != nil {
		h.printer.Error(err)
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
		h.info("usage: ai ask <question>")
		return
	}
	if err := fs.Parse(rest); err != nil {
		h.info("usage: ai ask [-c|--context] <question>")
		return
	}

	args := fs.Args()
	if len(args) == 0 {
		h.info("usage: ai ask [-c|--context] <question>")
		return
	}

	question := strings.TrimSpace(strings.Join(rest, " "))
	if question == "" {
		h.info("usage: ai ask <question>")
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

	h.info(out)
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
	h.info(out)
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

	h.info(out)

	command := strings.TrimPrefix(out, "COMMAND: ")
	command = strings.Split(command, "\n")[0]

	if strings.TrimSpace(command) == "" {
		h.warn("[aish] No runnable command was suggested.")
		return
	}
	if sessiondanger.IsDangerous(command) {
		h.warn("[aish] This command looks dangerous; refusing to auto-run.")
		return
	}
	if sessionbuiltins.IsNonPersisting(command) {
		h.warn("[aish] Note: builtins like 'cd'/'export' won’t change your parent shell; I won’t auto-run them.")
		return
	}

	fmt.Print("[aish] Run it now? [y/N] ")
	yes := shell.ConfirmFromStdin()
	if !yes {
		return
	}
	if err := runner.Run(command); err != nil {
		if h.printer != nil {
			h.printer.Error(errs.Wrap(err, "ai-run", "[aish] run error"))
		} else {
			fmt.Println("[aish] run error:", err)
		}
	}
}

func (h *Handler) buildContext() (string, bool, error) {
	builder := sessioncontext.NewBuilder(h.cfg)
	return builder.Build()
}

func (h *Handler) info(msg string) {
	if h.printer != nil {
		h.printer.Info(msg)
		return
	}
	fmt.Println(msg)
}

func (h *Handler) warn(msg string) {
	if h.printer != nil {
		h.printer.Warn(msg)
		return
	}
	fmt.Println(msg)
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
