package service

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
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

	created, err := s.store.Create(name, snippet)
	if err != nil {
		return created, nil, err
	}
	return created, script.Warnings, nil
}

func (s *Service) Run(name string, vars []string) error {
	snip, err := s.store.GetOne(name)
	if err != nil {
		return err
	}

	varsMap := map[string]string{}

	for _, v := range vars {
		varString := strings.Split(v, "=")
		if len(varString) != 2 {
			return fmt.Errorf("[aish]:Error at %s Expected variable_name=value", v)
		}
		varName := varString[0]
		varValue := varString[1]
		varsMap[varName] = varValue
	}

	if len(vars) < snip.VarsCount {
		missingVars := make([]string, 0, snip.VarsCount)
		for _, v := range snip.Vars {
			if _, ok := varsMap[v]; !ok {
				missingVars = append(missingVars, v)
			}
		}

		return fmt.Errorf("[aish]: Missing Variables %s", strings.Join(missingVars, ","))
	}

	for i := range snip.Steps {
		cmd := snip.Steps[i].Cmd
		exec := strings.Join(snip.Steps[i].Exec, "\x00")
		for _, v := range snip.Vars {
			placeholder := "[[" + v + "]]"
			if strings.Contains(cmd, placeholder) {
				cmd = strings.ReplaceAll(cmd, placeholder, varsMap[v])
			}
			if strings.Contains(exec, placeholder) {
				exec = strings.ReplaceAll(exec, placeholder, varsMap[v])
			}
		}
		snip.Steps[i].Cmd = cmd
		snip.Steps[i].Exec = strings.Split(exec, "\x00")
	}

	for i, step := range snip.Steps {
		if len(step.Cmd) > 0 {
			fmt.Println(step.Cmd)
			cmd := exec.Command("/bin/sh", "-c", step.Cmd)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("[aish]:Error at Line: %d \n Command: %s", i, step.Cmd)
			}
		} else {
			fmt.Println(step.Exec)
			cmd := exec.Command(step.Exec[0], step.Exec[1:]...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("[aish]:Error at Line: %d \n Command: %s", i, strings.Join(step.Exec, " "))
			}
		}
	}

	return nil
}

func (s *Service) View(name string) error {
	snip, err := s.store.GetOne(name)
	if err != nil {
		return err
	}

	fmt.Printf("snippet: %s\n", name)
	if snip.CreatedAt != nil {
		fmt.Printf("created: %s\n", snip.CreatedAt.Format(time.RFC3339))
	}
	if snip.VarsCount > 0 {
		fmt.Printf("vars: %s (%d)\n", strings.Join(snip.Vars, ", "), snip.VarsCount)
	}
	if len(snip.Steps) > 0 {
		fmt.Println("steps:")
		for i, step := range snip.Steps {
			if len(step.Cmd) > 0 {
				fmt.Printf("[%d] $ %s\n", i, step.Cmd)
			} else {
				fmt.Printf("[%d] $ %s\n", i, strings.Join(step.Exec, " "))
			}

		}
	}

	return nil
}

func (s *Service) List() error {
	snips, err := s.store.LoadAll()
	if err != nil {
		return err
	}

	names := slices.Collect(maps.Keys(snips))

	for i, name := range names {
		fmt.Printf("[%d] %s\n", i, name)
	}

	return nil
}

func (s *Service) Delete(name string) error {
	return s.store.DeleteOne(name)
}
