package model

import "time"

// Step models a single operation in a snippet, either shell string or exec argv.
type Step struct {
	Cmd  string   `yaml:"cmd"`
	Exec []string `yaml:"exec"`
}

// Snippet captures the serialized representation of a snippet entry.
type Snippet struct {
	Steps     []Step     `yaml:"steps"`
	Notes     string     `yaml:"notes,omitempty"`
	Tags      []string   `yaml:"tags,omitempty"`
	CreatedAt *time.Time `yaml:"created_at,omitempty"`
	UpdatedAt *time.Time `yaml:"updated_at,omitempty"`
	Runs      int        `yaml:"runs,omitempty"`
	Vars      []string   `yaml:"vars,omitempty"`
	VarsCount int        `yaml:"vars_count,omitempty"`
}
