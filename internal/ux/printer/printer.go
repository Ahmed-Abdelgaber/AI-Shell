package printer

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/mr-gaber/ai-shell/internal/errs"
	"golang.org/x/term"
)

// Printer handles user-facing output with optional color formatting.
type Printer struct {
	out   io.Writer
	err   io.Writer
	color bool
}

// New constructs a printer writing to the provided io.Writers.
// Color output is enabled only when the destination is a TTY and NO_COLOR/AISH_NO_COLOR are unset.
func New(out, err io.Writer) *Printer {
	p := &Printer{out: out, err: err}
	p.color = shouldColor(out) || shouldColor(err)
	if noColorEnv() {
		p.color = false
	}
	return p
}

// Error prints an error message, including structured metadata when available.
func (p *Printer) Error(err error) {
	if err == nil {
		return
	}

	label := "error"
	code := ""
	fields := map[string]string{}
	severity := errs.SeverityError

	if enriched, ok := errs.From(err); ok {
		if enriched.Code() != "" {
			code = enriched.Code()
		}
		if enriched.Fields() != nil {
			fields = enriched.Fields()
		}
		severity = enriched.Severity()
		if enriched.Error() != "" {
			err = enriched
		}
	}

	switch severity {
	case errs.SeverityWarn:
		label = "warn"
	case errs.SeverityInfo:
		label = "info"
	}

	header := strings.TrimSpace(err.Error())
	if header == "" {
		header = "an unexpected error occurred"
	}

	codePart := ""
	if code != "" {
		codePart = fmt.Sprintf("(%s) ", code)
	}

	line := fmt.Sprintf("[%s] %s%s", label, codePart, header)
	p.write(p.err, p.applyColor(line, colorForSeverity(severity)))

	if len(fields) > 0 {
		for _, k := range sortedKeys(fields) {
			p.write(p.err, p.applyColor(fmt.Sprintf("    %s: %s", k, fields[k]), faintColor))
		}
	}
}

// Warn prints a warning message.
func (p *Printer) Warn(msg string) {
	if strings.TrimSpace(msg) == "" {
		return
	}
	p.write(p.out, p.applyColor(fmt.Sprintf("[warn] %s", msg), colorYellow))
}

// Info prints an informational message.
func (p *Printer) Info(msg string) {
	if strings.TrimSpace(msg) == "" {
		return
	}
	p.write(p.out, msg)
}

// Success prints a success message.
func (p *Printer) Success(msg string) {
	if strings.TrimSpace(msg) == "" {
		return
	}
	p.write(p.out, p.applyColor(msg, colorGreen))
}

func (p *Printer) write(dst io.Writer, msg string) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	_, _ = io.WriteString(dst, msg)
}

func (p *Printer) applyColor(msg string, code string) string {
	if !p.color || code == "" {
		return msg
	}
	return code + msg + colorReset
}

func shouldColor(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func noColorEnv() bool {
	for _, k := range []string{"NO_COLOR", "AISH_NO_COLOR"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	faintColor  = "\033[2m"
	colorReset  = "\033[0m"
)

func colorForSeverity(sev errs.Severity) string {
	switch sev {
	case errs.SeverityWarn:
		return colorYellow
	case errs.SeverityInfo:
		return faintColor
	default:
		return colorRed
	}
}
