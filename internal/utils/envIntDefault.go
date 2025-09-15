package utils

import (
	"os"
	"strconv"
	"strings"
)

/*
EnvIntDefault reads an environment variable and parses it as an integer,
falling back to the provided default when unset or invalid. It centralists this
pattern for consistent configuration handling throughout the app.
*/
func EnvIntDefault(name string, def int) int {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
