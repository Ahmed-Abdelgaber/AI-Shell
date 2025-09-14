package ai

import (
	"context"
	"strings"
)

var aiProvider, providerErr = NewAIProvider()

func Ask(question string) (answer string, err error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH, a terse terminal assistant. Prefer one good command with a one-line explanation. Be concise."
	user := strings.TrimSpace(question)
	// Initialize the AI provider.
	// aiProvider, err := NewAIProvider()
	if providerErr != nil {
		return "", providerErr
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

func Why(tail string) (string, error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH. Propose ONE safe fix command and a one-sentence rationale. Output strictly in the following format:\nCOMMAND: <single-line>\nWHY: <one sentence>"
	user := tail
	// Initialize the AI provider.
	// aiProvider, err := NewAIProvider()
	if providerErr != nil {
		return "", providerErr
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

func Fix(tail string) (string, error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH. Propose ONE safe fix command and a one-sentence rationale. Output strictly in the following format:\nCOMMAND: <single-line>\nWHY: <one sentence>"
	user := tail
	// Initialize the AI provider.
	// aiProvider, err := NewAIProvider()
	if providerErr != nil {
		return "", providerErr
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}
