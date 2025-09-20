package parser

import (
	"fmt"
	"sort"

	"github.com/mr-gaber/ai-shell/internal/snippets/model"
)

type stepKind int

const (
	stepExec stepKind = iota
	stepCmd
)

// Script holds the parsed representation of a snippet along with auxiliary data.
type Script struct {
	Steps    []model.Step
	Vars     []string
	Warnings []string
}

// Build constructs a Script from normalized lines, classifying each line and tokenizing as needed.
func Build(lines []string) (Script, error) {
	if len(lines) == 0 {
		return Script{}, fmt.Errorf("snip add: nothing to save (no lines)")
	}

	warnings := make([]string, 0)
	steps := make([]model.Step, 0, len(lines))

	vars := collectVars(lines)
	sort.Strings(vars)

	for i, ln := range lines {
		if looksLikeSplitSingleCommand(ln) {
			warnings = append(warnings,
				fmt.Sprintf("[aish] line %d looks like a split single command; paste as one line if you intended a single cmd", i+1))
		}

		switch classify(ln) {
		case stepCmd:
			steps = append(steps, model.Step{Cmd: ln})
		case stepExec:
			argv, err := tokenizeExec(ln)
			if err != nil {
				return Script{}, err
			}
			if len(argv) == 0 {
				continue
			}
			steps = append(steps, model.Step{Exec: argv})
		}
	}

	if len(steps) == 0 {
		return Script{}, fmt.Errorf("snip add: nothing to save (no valid steps)")
	}

	return Script{Steps: steps, Vars: vars, Warnings: warnings}, nil
}
