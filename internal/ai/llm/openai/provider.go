package openai

import (
	"errors"
	"os"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func NewClient() (*openai.Client, error) {
	// Load the API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	// Check if the API key is set
	if apiKey == "" {
		// If the API key is not set, return an error
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}
	// Create a new OpenAI client with the API key
	client := openai.NewClient(option.WithAPIKey(apiKey))
	// Return the client and no error
	return &client, nil
}
