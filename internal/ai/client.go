package ai

import (
	"os"

	openai "github.com/openai/openai-go/v2"
)

func NewClient() *openai.Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable is not set")
	}
	client := openai.NewClient(openai.WithAPIKey(apiKey))
	return client
}
