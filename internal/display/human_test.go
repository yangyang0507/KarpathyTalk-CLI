package display

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"kt/internal/client"
)

// ── timeAgoFrom ───────────────────────────────────────────────────────────────

func TestTimeAgoFrom(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1m ago"},
		{"45 minutes ago", now.Add(-45 * time.Minute), "45m ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1h ago"},
		{"23 hours ago", now.Add(-23 * time.Hour), "23h ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1d ago"},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), "6d ago"},
		{"8 days ago (date)", now.Add(-8 * 24 * time.Hour), "2024-06-07"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := timeAgoFrom(tc.t, now)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// ── RenderHeader ──────────────────────────────────────────────────────────────

func makePost(id int64, username, displayName string) client.Post {
	return client.Post{
		ID: id,
		Author: client.UserRef{
			Username:    username,
			DisplayName: displayName,
		},
		LikeCount:   5,
		RepostCount: 2,
		ReplyCount:  3,
	}
}

func TestRenderHeader_FitsWidth(t *testing.T) {
	p := makePost(42, "alice", "Alice")
	for _, width := range []int{80, 100, 120} {
		got := RenderHeader(p, width)
		visWidth := lipgloss.Width(got)
		if visWidth > width {
			t.Errorf("width=%d: rendered width %d exceeds limit", width, visWidth)
		}
	}
}

func TestRenderHeader_ContainsMetadata(t *testing.T) {
	p := makePost(231, "ajay99511", "Ajay Elika")
	got := RenderHeader(p, 120)

	// Strip ANSI codes to check plain text content
	plain := lipgloss.NewStyle().Render(got) // just to use the import; check raw
	_ = plain

	checks := []string{"#231", "ajay99511", "Ajay Elika", "♥", "↺", "✦"}
	for _, sub := range checks {
		if !strings.Contains(got, sub) {
			t.Errorf("RenderHeader missing %q in output", sub)
		}
	}
}

// ── RenderSummaryCard ─────────────────────────────────────────────────────────

func TestRenderSummaryCard_ContainsSeparators(t *testing.T) {
	p := makePost(1, "bob", "Bob")
	p.ContentMarkdown = "hello world"
	got := RenderSummaryCard(p, 80)

	if !strings.Contains(got, "─") {
		t.Error("RenderSummaryCard: missing post separator (─)")
	}
	if !strings.Contains(got, "╌") {
		t.Error("RenderSummaryCard: missing sub-separator (╌)")
	}
}

func TestRenderSummaryCard_TruncatesLongContent(t *testing.T) {
	p := makePost(2, "carol", "Carol")
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line of content"
	}
	p.ContentMarkdown = strings.Join(lines, "\n")

	got := RenderSummaryCard(p, 80)
	if !strings.Contains(got, "more lines") {
		t.Error("expected '[+N more lines]' indicator for long content")
	}
}

func TestRenderSummaryCard_ShortContent_NoTruncation(t *testing.T) {
	p := makePost(3, "dave", "Dave")
	p.ContentMarkdown = "short post"
	got := RenderSummaryCard(p, 80)
	if strings.Contains(got, "more lines") {
		t.Error("unexpected truncation indicator for short content")
	}
}

// ── RenderFull ────────────────────────────────────────────────────────────────

func TestRenderFull_WithParentPost(t *testing.T) {
	p := makePost(10, "eve", "Eve")
	parentID := int64(5)
	p.ParentPostID = &parentID
	p.ContentMarkdown = "reply content"

	got := RenderFull(p, 80)
	if !strings.Contains(got, "↩ reply to #5") {
		t.Errorf("RenderFull: expected reply indicator, got:\n%s", got)
	}
}

func TestRenderFull_WithQuote(t *testing.T) {
	p := makePost(11, "frank", "Frank")
	quoteID := int64(3)
	p.QuoteOfID = &quoteID
	p.ContentMarkdown = "quoting content"

	got := RenderFull(p, 80)
	if !strings.Contains(got, "❝ quoting #3") {
		t.Errorf("RenderFull: expected quote indicator, got:\n%s", got)
	}
}

func TestRenderFull_WithRevision(t *testing.T) {
	p := makePost(12, "grace", "Grace")
	p.RevisionNumber = 2
	p.RevisionCount = 3
	p.ContentMarkdown = "revised content"

	got := RenderFull(p, 80)
	if !strings.Contains(got, "rev 2/3") {
		t.Errorf("RenderFull: expected revision indicator, got:\n%s", got)
	}
}

func TestRenderFull_StripsImages(t *testing.T) {
	p := makePost(13, "henry", "Henry")
	p.ContentMarkdown = "see this ![photo](https://example.com/photo.jpg)"

	got := RenderFull(p, 80)
	if strings.Contains(got, "![photo]") {
		t.Error("RenderFull: raw image markdown should be stripped")
	}
	if !strings.Contains(got, "image:") {
		t.Error("RenderFull: expected [image: ...] replacement")
	}
}
