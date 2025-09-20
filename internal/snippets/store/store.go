package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/mr-gaber/ai-shell/internal/snippets/model"
	"gopkg.in/yaml.v3"
)

// Store persists snippets on disk using YAML.
type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) LoadAll() (map[string]model.Snippet, error) {
	db := map[string]model.Snippet{}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return db, nil
		}
		return nil, fmt.Errorf("aish: cannot read %q: %w", s.path, err)
	}

	if len(bytes.TrimSpace(b)) == 0 {
		return db, nil
	}

	if err := yaml.Unmarshal(b, &db); err != nil {
		return nil, fmt.Errorf("aish: invalid YAML in %q: %w", s.path, err)
	}

	return db, nil
}

func (s *Store) Exists(name string) (bool, error) {
	db, err := s.LoadAll()
	if err != nil {
		return false, err
	}
	_, ok := db[name]
	return ok, nil
}

func (s *Store) Save(name string, snippet model.Snippet) (bool, error) {
	db, err := s.LoadAll()
	if err != nil {
		return false, err
	}

	_, exists := db[name]
	db[name] = snippet

	out, err := yaml.Marshal(db)
	if err != nil {
		return !exists, fmt.Errorf("aish: cannot encode snippets YAML: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return !exists, fmt.Errorf("aish: cannot write temp file %q: %w", tmp, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return !exists, fmt.Errorf("aish: cannot finalize save %q: %w", s.path, err)
	}

	return !exists, nil
}
