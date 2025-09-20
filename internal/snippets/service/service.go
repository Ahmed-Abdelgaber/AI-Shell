package service

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/mr-gaber/ai-shell/internal/errs"
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
		return nil, errs.New("snip-no-path", "[aish] Cannot find snippets yaml file")
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
			return false, nil, errs.Wrap(err, "snip-store-exists", "failed to inspect snippets store")
		}
		if exists {
			return false, nil, errs.New("snip-exists", fmt.Sprintf("[aish] snippet %q already exists. Use --force to overwrite", name))
		}
	}

	lines := parser.Normalize(raw)
	script, err := parser.Build(lines)
	if err != nil {
		return false, nil, errs.Wrap(err, "snip-parse", "failed to build snippet script")
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
		return created, nil, errs.Wrap(err, "snip-store-save", "failed to persist snippet")
	}
	return created, script.Warnings, nil
}

func (s *Service) Run(name string, vars []string) error {
	snip, err := s.store.GetOne(name)
	if err != nil {
		return errs.Wrap(err, "snip-missing", "failed to load snippet", errs.WithFields(map[string]string{"name": name}))
	}

	varsMap := map[string]string{}

	for _, v := range vars {
		varString := strings.SplitN(v, "=", 2)
		if len(varString) != 2 {
			return errs.New("snip-var-format", fmt.Sprintf("[aish] invalid variable format %q (expected name=value)", v))
		}
		varName := strings.TrimSpace(varString[0])
		varValue := varString[1]
		if varName == "" {
			return errs.New("snip-var-name", fmt.Sprintf("[aish] invalid variable name in %q", v))
		}
		varsMap[varName] = varValue
	}

	if len(vars) < snip.VarsCount {
		missingVars := make([]string, 0, snip.VarsCount)
		for _, v := range snip.Vars {
			if _, ok := varsMap[v]; !ok {
				missingVars = append(missingVars, v)
			}
		}

		return errs.New("snip-missing-vars", fmt.Sprintf("[aish] missing variables %s", strings.Join(missingVars, ",")))
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
				return errs.Wrap(err, "snip-step-cmd", "[aish] command failed", errs.WithFields(map[string]string{"line": fmt.Sprintf("%d", i+1), "command": step.Cmd}))
			}
		} else {
			fmt.Println(step.Exec)
			cmd := exec.Command(step.Exec[0], step.Exec[1:]...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				return errs.Wrap(err, "snip-step-exec", "[aish] command failed", errs.WithFields(map[string]string{"line": fmt.Sprintf("%d", i+1), "command": strings.Join(step.Exec, " ")}))
			}
		}
	}

	return nil
}

func (s *Service) View(name string) error {
	snip, err := s.store.GetOne(name)
	if err != nil {
		return errs.Wrap(err, "snip-missing", "failed to load snippet", errs.WithFields(map[string]string{"name": name}))
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
		return errs.Wrap(err, "snip-list", "failed to read snippets")
	}

	names := slices.Collect(maps.Keys(snips))

	for i, name := range names {
		fmt.Printf("[%d] %s\n", i, name)
	}

	return nil
}

func (s *Service) Delete(name string) error {
	if err := s.store.DeleteOne(name); err != nil {
		return errs.Wrap(err, "snip-delete", "failed to delete snippet", errs.WithFields(map[string]string{"name": name}))
	}
	return nil
}
