package ai

import (
	"context"
	"fmt"
	"strings"
)

func ask(question string) (answer string, err error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH, a terse terminal assistant. Prefer one good command with a one-line explanation. Be concise."
	user := strings.TrimSpace(question)
	// Initialize the AI provider.
	aiProvider, err := NewAIProvider()
	if err != nil {
		return "", err
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

func why(lastCmd string, lastExit int, tail string) (string, error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH. Propose ONE safe fix command and a one-sentence rationale. Output strictly in the following format:\nCOMMAND: <single-line>\nWHY: <one sentence>"
	user := fmt.Sprintf(
		"Last command: %q\nExit code: %d\nRecent output:\n%s",
		strings.TrimSpace(lastCmd), lastExit, tail,
	)
	// Initialize the AI provider.
	aiProvider, err := NewAIProvider()
	if err != nil {
		return "", err
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}

func fix(lastCmd string, lastExit int, tail string) (string, error) {
	// Prepare the context and messages for the AI provider.
	ctx := context.Background()
	// Prepare system and user messages.
	// The system message sets the behavior of the AI assistant.
	// The user message contains the actual question to be answered.
	sys := "You are AISH. Propose ONE safe fix command and a one-sentence rationale. Output strictly in the following format:\nCOMMAND: <single-line>\nWHY: <one sentence>"
	user := fmt.Sprintf(
		"Last command: %q\nExit code: %d\nRecent output:\n%s",
		strings.TrimSpace(lastCmd), lastExit, tail,
	)
	// Initialize the AI provider.
	aiProvider, err := NewAIProvider()
	if err != nil {
		return "", err
	}
	// Ask the question using the AI provider.
	return aiProvider.Ask(ctx, user, sys)
}
