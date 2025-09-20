package snip

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/cli/shared"
	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/shell"
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
		h.handleList()
	case "add":
		h.handleAdd(args[1:])
	case "run":
		h.handleRun(args[1:])
	case "view":
		h.handleView(args[1:])
	case "delete":
		h.handleDelete(args[1:])
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

	fmt.Print("[aish] Run it now? [y/N] ")
	yes := shell.ConfirmFromStdin()
	if !yes {
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		fmt.Println(err)
		return
	}

	vars := rest[1:]

	if err := svc.Run(rest[0], vars); err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) handleView(rest []string) {
	if len(rest) < 1 {
		fmt.Println("usage: snip view <name>")
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := svc.View(rest[0]); err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) handleList() {
	svc, err := h.ensureService()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := svc.List(); err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) handleDelete(rest []string) {
	if len(rest) < 1 {
		fmt.Println("usage: snip delete <name>")
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := svc.Delete(rest[0]); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("[aish]: Snippet: %s Deleted Successfully!\n", rest[0])
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
