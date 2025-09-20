package main

import (
	"fmt"
	"os"

	"github.com/mr-gaber/ai-shell/internal/app"
)

func main() {
	a := app.New()
	if err := a.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "aish:", err)
		os.Exit(1)
	}
}
