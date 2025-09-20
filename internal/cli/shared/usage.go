package shared

import "fmt"

// PrintUsage centralises usage text printing to keep handlers lean.
func PrintUsage(text string) {
    fmt.Println(text)
}
