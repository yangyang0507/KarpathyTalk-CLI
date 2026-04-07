package display

import (
	"fmt"
	"os"
	"regexp"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

var imageRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

// StripMarkdownImages replaces ![alt](url) with [image: url] so glamour
// doesn't emit raw markdown image syntax in the terminal.
func StripMarkdownImages(text string) string {
	return imageRegex.ReplaceAllStringFunc(text, func(match string) string {
		groups := imageRegex.FindStringSubmatch(match)
		if len(groups) >= 3 {
			return fmt.Sprintf("[image: %s]", groups[2])
		}
		return match
	})
}

// termWidth returns the current terminal width, falling back to 80.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// RenderMarkdownWidth renders Markdown text to ANSI terminal output at a given column width.
func RenderMarkdownWidth(text string, width int) string {
	if width < 20 {
		width = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return out
}

// renderMarkdown renders Markdown text to ANSI terminal output.
// Falls back to plain text if rendering fails.
func renderMarkdown(text string) string {
	return RenderMarkdownWidth(text, termWidth()-4)
}
