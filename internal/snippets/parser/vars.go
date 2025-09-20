package parser

import (
	"regexp"

	mapset "github.com/deckarep/golang-set/v2"
)

var rxBrackets = regexp.MustCompile(`\[\[([A-Za-z_][A-Za-z0-9_]*)\]\]`)

func detectVarsInString(s string) []string {
	if s == "" {
		return nil
	}

	vars := []string{}

	for _, m := range rxBrackets.FindAllStringSubmatchIndex(s, -1) {
		start := m[0]
		nameStart := m[2]
		nameEnd := m[3]
		if start > 0 && s[start-1] == '\\' {
			continue
		}
		vars = append(vars, s[nameStart:nameEnd])
	}

	return vars
}

func collectVars(lines []string) []string {
	set := mapset.NewSet[string]()

	for _, line := range lines {
		stepVars := detectVarsInString(line)
		if len(stepVars) == 0 {
			continue
		}
		for _, v := range stepVars {
			set.Add(v)
		}
	}

	return set.ToSlice()
}
