package parser

import (
	"fmt"
	"unicode"
)

func tokenizeExec(line string) ([]string, error) {
	var args []string
	var cur []rune

	inSingle := false
	inDouble := false
	escape := false

	flush := func() {
		if len(cur) > 0 {
			args = append(args, string(cur))
			cur = cur[:0]
		}
	}

	for _, r := range line {
		switch {
		case escape:
			cur = append(cur, r)
			escape = false

		case r == '\\' && !inSingle:
			escape = true

		case r == '\'' && !inDouble:
			inSingle = !inSingle

		case r == '"' && !inSingle:
			inDouble = !inDouble

		case !inSingle && !inDouble && unicode.IsSpace(r):
			flush()

		default:
			cur = append(cur, r)
		}
	}

	if escape {
		cur = append(cur, '\\')
	}
	if inSingle {
		return nil, fmt.Errorf("snip add: unmatched single quote")
	}
	if inDouble {
		return nil, fmt.Errorf("snip add: unmatched double quote")
	}

	flush()
	return args, nil
}
