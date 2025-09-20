package ai

import (
	"context"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/ai/prompts"
	"github.com/mr-gaber/ai-shell/internal/ai/providers"
	"github.com/mr-gaber/ai-shell/internal/config"
)

type Service struct {
	provider providers.Provider
}

func NewService(cfg config.Config) (*Service, error) {
	p, err := providers.FromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{provider: p}, nil
}

func (s *Service) Ask(question string) (string, error) {
	return s.provider.Ask(context.Background(), strings.TrimSpace(question), prompts.AskSystem)
}

func (s *Service) Why(contextText string) (string, error) {
	return s.provider.Ask(context.Background(), contextText, prompts.WhySystem)
}

func (s *Service) Fix(contextText string) (string, error) {
	return s.provider.Ask(context.Background(), contextText, prompts.FixSystem)
}
