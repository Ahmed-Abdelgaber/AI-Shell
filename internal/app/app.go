package app

import (
	"fmt"
	"os"

	"github.com/mr-gaber/ai-shell/internal/cli/router"
	"github.com/mr-gaber/ai-shell/internal/config"
	"github.com/mr-gaber/ai-shell/internal/shell/launcher"
	"github.com/mr-gaber/ai-shell/internal/ux/printer"
	"golang.org/x/term"
)

type App struct {
	cfg     config.Config
	router  *router.Router
	printer *printer.Printer
}

func New() *App {
	cfg := config.LoadFromEnv()
	p := printer.New(os.Stdout, os.Stderr)
	return &App{
		cfg:     cfg,
		router:  router.New(cfg, p),
		printer: p,
	}
}

func (a *App) Run(args []string) error {
	if a.router.Route(args) {
		return nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("no TTY on stdin. Run from a real terminal")
	}

	launch := launcher.New(a.cfg)
	if err := launch.Run(); err != nil {
		return fmt.Errorf("shell session error: %w", err)
	}

	return nil
}
