package providers

import (
	"context"
	"fmt"

	"github.com/mr-gaber/ai-shell/internal/ai/providers/ollama"
	"github.com/mr-gaber/ai-shell/internal/ai/providers/openai"
	"github.com/mr-gaber/ai-shell/internal/config"
)

// Provider encapsulates a backing large language model client.
type Provider interface {
	Ask(ctx context.Context, userQuestion string, systemMessage string) (string, error)
}

// FromConfig instantiates a Provider based on configuration.
func FromConfig(cfg config.Config) (Provider, error) {
	switch cfg.AIProvider {
	case "openai":
		return openai.New(cfg.OpenAIKey)
	case "ollama":
		return ollama.New()
	default:
		return nil, fmt.Errorf("unknown AI provider %q", cfg.AIProvider)
	}
}
