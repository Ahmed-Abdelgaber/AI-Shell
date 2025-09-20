package config

import (
	"os"
	"strconv"
	"strings"
)

func intDefault(name string, def int) int {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// Int reads an environment variable as an integer, falling back to def when unset or invalid.
func Int(name string, def int) int {
	return intDefault(name, def)
}
