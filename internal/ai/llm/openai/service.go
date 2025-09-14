package openai

import (
	"context"
	"errors"

	openai "github.com/openai/openai-go/v2"
)

// OpenAIService implements the AIProvider interface for OpenAI.
type OpenAIService struct{}

func (s *OpenAIService) Ask(ctx context.Context, userQuestion string, systemMessage string) (string, error) {
	// Initialize OpenAI client
	c, err := NewClient()
	if err != nil {
		return "", err
	}
	// Create chat completion request
	chatCompletion, err := c.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(userQuestion),
			openai.SystemMessage(systemMessage),
		},
		Model: openai.ChatModelGPT4oMini,
	})
	if err != nil {
		return "", errors.New("failed to create completion: " + err.Error())
	}

	if len(chatCompletion.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}
	// Return the response text
	return chatCompletion.Choices[0].Message.Content, nil
}
