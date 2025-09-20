package service

import (
	"fmt"
	"regexp"
)

var nameAllowed = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("snip add: missing snippet name")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("snip add: invalid name %q", name)
	}
	if len(name) > 64 {
		return fmt.Errorf("snip add: name too long (max 64 chars)")
	}
	if !nameAllowed.MatchString(name) {
		return fmt.Errorf(`snip add: invalid name %q (allowed: [A-Za-z0-9._-])`, name)
	}
	return nil
}
