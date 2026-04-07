package tui

import (
	"strings"
	"testing"
)

// ── stripMarkdownSyntax ───────────────────────────────────────────────────────

func TestStripMarkdownSyntax_Headings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"# Title", "Title"},
		{"## Section", "Section"},
		{"### Sub", "Sub"},
	}
	for _, c := range cases {
		got := stripMarkdownSyntax(c.in)
		if got != c.want {
			t.Errorf("input %q: got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestStripMarkdownSyntax_Bold(t *testing.T) {
	got := stripMarkdownSyntax("this is **bold** text")
	want := "this is bold text"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownSyntax_Italic(t *testing.T) {
	got := stripMarkdownSyntax("this is *italic* text")
	want := "this is italic text"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownSyntax_InlineCode(t *testing.T) {
	got := stripMarkdownSyntax("run `go test ./...` to test")
	want := "run go test ./... to test"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownSyntax_Links(t *testing.T) {
	got := stripMarkdownSyntax("see [the docs](https://example.com) for more")
	want := "see the docs for more"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripMarkdownSyntax_CodeBlock(t *testing.T) {
	input := "intro\n```go\nfmt.Println(\"hello\")\n```\noutro"
	got := stripMarkdownSyntax(input)
	if strings.Contains(got, "fmt.Println") {
		t.Errorf("code block content should be stripped, got: %q", got)
	}
}

func TestStripMarkdownSyntax_PlainText(t *testing.T) {
	input := "just a plain sentence with no markdown"
	got := stripMarkdownSyntax(input)
	if got != input {
		t.Errorf("plain text should be unchanged: got %q", got)
	}
}

// ── cardPreview ───────────────────────────────────────────────────────────────

func TestCardPreview_Short(t *testing.T) {
	line1, line2 := cardPreview("hello world", 80)
	if line1 != "  hello world" {
		t.Errorf("line1: got %q, want %q", line1, "  hello world")
	}
	if line2 != "" {
		t.Errorf("line2: got %q, want empty", line2)
	}
}

func TestCardPreview_TwoLines(t *testing.T) {
	input := "first line\nsecond line"
	line1, line2 := cardPreview(input, 80)
	if line1 != "  first line" {
		t.Errorf("line1: got %q, want %q", line1, "  first line")
	}
	if line2 != "  second line" {
		t.Errorf("line2: got %q, want %q", line2, "  second line")
	}
}

func TestCardPreview_SkipsBlankLines(t *testing.T) {
	input := "\n\nfirst non-empty\n\nsecond non-empty\n\n"
	line1, line2 := cardPreview(input, 80)
	if line1 != "  first non-empty" {
		t.Errorf("line1: got %q", line1)
	}
	if line2 != "  second non-empty" {
		t.Errorf("line2: got %q", line2)
	}
}

func TestCardPreview_Truncates(t *testing.T) {
	// width=20, maxLen = 20-4 = 16
	long := "this is a very long line that exceeds the limit"
	line1, _ := cardPreview(long, 20)
	// Should end with "..."
	if len(line1) == 0 {
		t.Fatal("line1 is empty")
	}
	runes := []rune(line1)
	last3 := string(runes[len(runes)-3:])
	if last3 != "..." {
		t.Errorf("expected truncated line to end with '...', got %q", line1)
	}
}

func TestCardPreview_Empty(t *testing.T) {
	line1, line2 := cardPreview("", 80)
	if line1 != "" {
		t.Errorf("line1: got %q, want empty", line1)
	}
	if line2 != "" {
		t.Errorf("line2: got %q, want empty", line2)
	}
}

