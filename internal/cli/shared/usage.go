package shared

import (
	"fmt"

	"github.com/mr-gaber/ai-shell/internal/ux/printer"
)

// PrintUsage displays CLI usage text, leveraging the shared printer when available.
func PrintUsage(p *printer.Printer, text string) {
	if p != nil {
		p.Info(text)
		return
	}
	fmt.Println(text)
}
