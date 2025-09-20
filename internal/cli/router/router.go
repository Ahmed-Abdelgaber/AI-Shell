package router

import (
	"github.com/mr-gaber/ai-shell/internal/cli/ai"
	clipsnip "github.com/mr-gaber/ai-shell/internal/cli/snip"
	"github.com/mr-gaber/ai-shell/internal/config"
)

// Router dispatches internal commands to their handlers based on argv.
type Router struct {
	ai   *ai.Handler
	snip *clipsnip.Handler
}

func New(cfg config.Config) *Router {
	return &Router{
		ai:   ai.New(cfg),
		snip: clipsnip.New(cfg),
	}
}

func (r *Router) Route(args []string) bool {
	if len(args) > 1 {
		switch args[1] {
		case "__ai":
			r.ai.Handle(args[2:])
			return true
		case "__snip":
			r.snip.Handle(args[2:])
			return true
		}
	}
	return false
}
