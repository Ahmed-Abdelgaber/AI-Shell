package snip

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/cli/shared"
	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/snippets/service"
)

// Handler processes snippet subcommands parsed by the router.
type Handler struct {
	cfg config.Config
	svc *service.Service
}

func New(cfg config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) Handle(args []string) {
	if len(args) == 0 || args[0] == "help" {
		shared.PrintUsage("snip usage:\n  snip ls\n  snip add <name> <command...>\n  snip run <name>")
		return
	}

	switch args[0] {
	case "ls":
		fmt.Println("[aish] (placeholder) no snippets yet")
	case "add":
		h.handleAdd(args[1:])
	case "run":
		h.handleRun(args[1:])
	default:
		fmt.Printf("snip: unknown subcommand %q\n", args[0])
	}
}

func (h *Handler) handleAdd(rest []string) {
	if len(rest) < 2 {
		fmt.Println("usage: snip add <name> <command...>")
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		fmt.Println(err)
		return
	}

	name := rest[0]
	raw := strings.Join(rest[1:], " ")

	created, warnings, err := svc.Add(name, raw, false)
	if err != nil {
		fmt.Println("", err)
		return
	}

	for _, warning := range warnings {
		fmt.Println("War: " + warning)
	}

	if !created {
		fmt.Println("Snippet not created")
		return
	}

	fmt.Println("Snippet created successfully!")
}

func (h *Handler) handleRun(rest []string) {
	if len(rest) < 1 {
		fmt.Println("usage: snip run <name>")
		return
	}
	fmt.Printf("[aish] (placeholder) would run snippet %q\n", rest[0])
}

func (h *Handler) ensureService() (*service.Service, error) {
	if h.svc != nil {
		return h.svc, nil
	}

	svc, err := service.New(h.cfg.Paths.SnippetsFile)
	if err != nil {
		return nil, err
	}
	h.svc = svc
	return h.svc, nil
}
