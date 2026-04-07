package display

import (
	"fmt"
	"os"
)

// PrintMarkdown writes raw markdown text to stdout (direct passthrough).
// Suitable for piping to llm tools or other text processors.
func PrintMarkdown(text string) {
	fmt.Fprint(os.Stdout, text)
	// Ensure trailing newline
	if len(text) > 0 && text[len(text)-1] != '\n' {
		fmt.Fprintln(os.Stdout)
	}
}
