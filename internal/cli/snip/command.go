package snip

import (
	"fmt"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/cli/shared"
	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/errs"
	"github.com/mr-gaber/ai-shell/internal/shell"
	"github.com/mr-gaber/ai-shell/internal/snippets/service"
	"github.com/mr-gaber/ai-shell/internal/ux/printer"
)

// Handler processes snippet subcommands parsed by the router.
type Handler struct {
	cfg     config.Config
	svc     *service.Service
	printer *printer.Printer
}

func New(cfg config.Config, p *printer.Printer) *Handler {
	return &Handler{cfg: cfg, printer: p}
}

func (h *Handler) Handle(args []string) {
	if len(args) == 0 || args[0] == "help" {
		shared.PrintUsage(h.printer, "snip usage:\n  snip ls\n  snip add <name> <command...>\n  snip run <name>\n  snip view <name>\n  snip delete <name>")
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
		h.warn(fmt.Sprintf("snip: unknown subcommand %q", args[0]))
	}
}

func (h *Handler) handleAdd(rest []string) {
	if len(rest) < 2 {
		h.info("usage: snip add <name> <command...>")
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		h.error(err)
		return
	}

	name := rest[0]
	raw := strings.Join(rest[1:], " ")

	created, warnings, err := svc.Add(name, raw, false)
	if err != nil {
		h.error(err)
		return
	}

	for _, warning := range warnings {
		h.warn("War: " + warning)
	}

	if !created {
		h.warn("Snippet not created")
		return
	}

	h.success("Snippet created successfully!")
}

func (h *Handler) handleRun(rest []string) {
	if len(rest) < 1 {
		h.info("usage: snip run <name>")
		return
	}

	fmt.Print("[aish] Run it now? [y/N] ")
	yes := shell.ConfirmFromStdin()
	if !yes {
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		h.error(err)
		return
	}

	vars := rest[1:]

	if err := svc.Run(rest[0], vars); err != nil {
		h.error(err)
	}
}

func (h *Handler) handleView(rest []string) {
	if len(rest) < 1 {
		h.info("usage: snip view <name>")
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		h.error(err)
		return
	}

	if err := svc.View(rest[0]); err != nil {
		h.error(err)
	}
}

func (h *Handler) handleList() {
	svc, err := h.ensureService()
	if err != nil {
		h.error(err)
		return
	}

	if err := svc.List(); err != nil {
		h.error(err)
	}
}

func (h *Handler) handleDelete(rest []string) {
	if len(rest) < 1 {
		h.info("usage: snip delete <name>")
		return
	}

	svc, err := h.ensureService()
	if err != nil {
		h.error(err)
		return
	}

	if err := svc.Delete(rest[0]); err != nil {
		h.error(err)
		return
	}

	h.success(fmt.Sprintf("[aish]: Snippet: %s deleted successfully!", rest[0]))
}

func (h *Handler) ensureService() (*service.Service, error) {
	if h.svc != nil {
		return h.svc, nil
	}

	svc, err := service.New(h.cfg.Paths.SnippetsFile)
	if err != nil {
		return nil, errs.Wrap(err, "snip-init", "failed to prepare snippets storage")
	}
	h.svc = svc
	return h.svc, nil
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

func (h *Handler) success(msg string) {
	if h.printer != nil {
		h.printer.Success(msg)
		return
	}
	fmt.Println(msg)
}

func (h *Handler) error(err error) {
	if err == nil {
		return
	}
	if h.printer != nil {
		h.printer.Error(err)
		return
	}
	fmt.Println(err)
}
