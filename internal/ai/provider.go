package ai

import (
	"context"
	"errors"
	"os"

	"github.com/mr-gaber/ai-shell/internal/ai/llm/openai"
)

// AIProvider defines the interface for interacting with different AI service providers.
type AIProvider interface {
	Ask(ctx context.Context, userQuestion string, systemMessage string) (string, error)
}

func NewAIProvider() (AIProvider, error) {
	// Determine the AI provider based on an environment variable.
	name := os.Getenv("AI_PROVIDER")
	if name == "" {
		name = "openai" // Default to OpenAI if not specified.
	}

	// Instantiate the appropriate AI provider based on the name.
	switch name {
	case "openai":
		return &openai.OpenAIService{}, nil
	default:
		return nil, errors.New("unknown AI provider")
	}
}
