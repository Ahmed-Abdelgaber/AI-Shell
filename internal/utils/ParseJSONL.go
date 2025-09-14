package utils

import (
	"encoding/json"
	"strings"
)

func ParseJSONL[T any](lines []string, keep func(T) bool) []T {
	out := make([]T, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		var e T
		if err := json.Unmarshal([]byte(ln), &e); err != nil {
			continue
		}
		if keep == nil || keep(e) {
			out = append(out, e)
		}
	}
	return out
}
