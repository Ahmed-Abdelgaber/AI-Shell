package openai

import (
	"context"
	"errors"
	"strings"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type Client struct {
	sdk openai.Client
}

func New(apiKey string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &Client{sdk: client}, nil
}

func (c *Client) Ask(ctx context.Context, userQuestion string, systemMessage string) (string, error) {
	chatCompletion, err := c.sdk.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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
	return chatCompletion.Choices[0].Message.Content, nil
}
