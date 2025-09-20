package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/mr-gaber/ai-shell/internal/snippets/model"
	"github.com/mr-gaber/ai-shell/internal/snippets/parser"
	"github.com/mr-gaber/ai-shell/internal/snippets/store"
)

// Service encapsulates snippet parsing and persistence workflows.
type Service struct {
	store *store.Store
}

func New(path string) (*Service, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("[aish] Cannot find snippets yaml file")
	}
	return &Service{store: store.New(path)}, nil
}

func (s *Service) Add(name string, raw string, force bool) (bool, []string, error) {
	if err := validateName(name); err != nil {
		return false, nil, err
	}

	if !force {
		exists, err := s.store.Exists(name)
		if err != nil {
			return false, nil, err
		}
		if exists {
			return false, nil, fmt.Errorf("[aish] snippet %q already exists. Use --force to overwrite", name)
		}
	}

	lines := parser.Normalize(raw)
	script, err := parser.Build(lines)
	if err != nil {
		return false, nil, err
	}

	now := time.Now()
	snippet := model.Snippet{
		Steps:     script.Steps,
		CreatedAt: &now,
		UpdatedAt: &now,
		Runs:      0,
		Vars:      script.Vars,
		VarsCount: len(script.Vars),
	}

	created, err := s.store.Save(name, snippet)
	if err != nil {
		return created, nil, err
	}
	return created, script.Warnings, nil
}
