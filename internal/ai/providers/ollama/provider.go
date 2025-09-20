package ollama

import (
	"context"
	"errors"
)

type Client struct{}

func New() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Ask(ctx context.Context, userQuestion string, systemMessage string) (string, error) {
	return "", errors.New("ollama provider is not implemented")
}
