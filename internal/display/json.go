package display

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintJSON writes value as indented JSON to stdout.
// Errors go to stderr to keep stdout clean for piping.
func PrintJSON(value any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, "error encoding JSON:", err)
	}
}
