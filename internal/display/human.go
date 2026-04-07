package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
)

var (
	styleID       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleAuthor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	styleMeta     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleAge      = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleLike     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleRepost   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	styleReplyC   = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	styleDim      = lipgloss.NewStyle().Faint(true).Italic(true)
	stylePaginate = lipgloss.NewStyle().Italic(true).Faint(true)
	styleSepColor = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleSubSep   = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
)

func sep(width int) string  { return styleSepColor.Render(strings.Repeat("─", width)) }
func subSep(width int) string { return styleSubSep.Render(strings.Repeat("╌", width)) }

// Sep returns a full-width post separator line at the given column width.
func Sep(width int) string { return sep(width) }

// SubSep returns a thin header/body divider line at the given column width.
func SubSep(width int) string { return subSep(width) }

// RenderHeader returns the styled one-line post header at the given column width.
// Stats are right-aligned.
func RenderHeader(p client.Post, width int) string {
	ago := timeAgo(p.CreatedAt.Time())

	left := fmt.Sprintf("%s  %s  %s",
		styleID.Render(fmt.Sprintf("#%d", p.ID)),
		styleAuthor.Render(p.Author.Username),
		styleMeta.Render(fmt.Sprintf("(%s)", p.Author.DisplayName)),
	)
	right := fmt.Sprintf("%s   %s  %s  %s",
		styleAge.Render(ago),
		styleLike.Render(fmt.Sprintf("♥ %d", p.LikeCount)),
		styleRepost.Render(fmt.Sprintf("↺ %d", p.RepostCount)),
		styleReplyC.Render(fmt.Sprintf("✦ %d", p.ReplyCount)),
	)

	pad := width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 2 {
		pad = 2
	}
	return left + strings.Repeat(" ", pad) + right
}

func previewLines(markdown string, maxLines int) (preview string, extra int) {
	lines := strings.Split(strings.TrimRight(markdown, "\n"), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n"), 0
	}
	return strings.Join(lines[:maxLines], "\n"), len(lines) - maxLines
}

// RenderSummaryCard returns the full card string (header + subSep + preview + sep).
// Used by the TUI list view and the non-TTY print fallback.
func RenderSummaryCard(p client.Post, width int) string {
	preview, extra := previewLines(StripMarkdownImages(p.ContentMarkdown), 3)
	rendered := strings.TrimLeft(RenderMarkdownWidth(preview, width-4), "\n")

	var sb strings.Builder
	sb.WriteString(RenderHeader(p, width))
	sb.WriteByte('\n')
	sb.WriteString(subSep(width))
	sb.WriteByte('\n')
	sb.WriteString(rendered)
	if extra > 0 {
		sb.WriteString(styleDim.Render(fmt.Sprintf("  [+%d more lines]", extra)))
		sb.WriteByte('\n')
		sb.WriteByte('\n')
	}
	sb.WriteString(sep(width))
	return sb.String()
}

// RenderFull returns the full post content string for the detail viewport.
func RenderFull(p client.Post, width int) string {
	var sb strings.Builder
	sb.WriteString(RenderHeader(p, width))
	sb.WriteByte('\n')
	sb.WriteString(subSep(width))
	sb.WriteByte('\n')

	if p.ParentPostID != nil {
		sb.WriteString(styleDim.Render(fmt.Sprintf("  ↩ reply to #%d", *p.ParentPostID)))
		sb.WriteByte('\n')
	}
	if p.QuoteOfID != nil {
		sb.WriteString(styleDim.Render(fmt.Sprintf("  ❝ quoting #%d", *p.QuoteOfID)))
		sb.WriteByte('\n')
	}
	if p.RevisionCount > 1 {
		sb.WriteString(styleDim.Render(fmt.Sprintf("  rev %d/%d", p.RevisionNumber, p.RevisionCount)))
		sb.WriteByte('\n')
	}

	rendered := strings.TrimLeft(RenderMarkdownWidth(StripMarkdownImages(p.ContentMarkdown), width-4), "\n")
	sb.WriteString(rendered)
	sb.WriteString(sep(width))
	return sb.String()
}

// PostSummary prints a single post in the human-friendly terminal format.
func PostSummary(p client.Post) {
	w := termWidth()
	fmt.Println(RenderSummaryCard(p, w))
	fmt.Println()
}

// PostFull prints a single post with full content.
func PostFull(p client.Post) {
	w := termWidth()
	fmt.Println(RenderFull(p, w))
	fmt.Println()
}

// UserProfile prints a user profile summary.
func UserProfile(u client.User) {
	inner := fmt.Sprintf("%s %s\n%s  %s  %s\n%s\n%s",
		styleAuthor.Render("@"+u.Username),
		styleMeta.Render(fmt.Sprintf("(%s)", u.DisplayName)),
		styleMeta.Render("Posts:")+styleID.Render(fmt.Sprintf(" %d", u.PostCount)),
		styleMeta.Render("Followers:")+styleID.Render(fmt.Sprintf(" %d", u.FollowerCount)),
		styleMeta.Render("Following:")+styleID.Render(fmt.Sprintf(" %d", u.FollowingCount)),
		styleMeta.Render("Profile: ")+u.ProfileURL,
		styleMeta.Render("GitHub:  ")+u.GitHubURL,
	)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		MaxWidth(termWidth() - 2).
		Render(inner)
	fmt.Println(box)
	fmt.Println()
}

// RenderUserProfile returns the user profile box as a string.
func RenderUserProfile(u client.User, width int) string {
	inner := fmt.Sprintf("%s %s\n%s  %s  %s\n%s\n%s",
		styleAuthor.Render("@"+u.Username),
		styleMeta.Render(fmt.Sprintf("(%s)", u.DisplayName)),
		styleMeta.Render("Posts:")+styleID.Render(fmt.Sprintf(" %d", u.PostCount)),
		styleMeta.Render("Followers:")+styleID.Render(fmt.Sprintf(" %d", u.FollowerCount)),
		styleMeta.Render("Following:")+styleID.Render(fmt.Sprintf(" %d", u.FollowingCount)),
		styleMeta.Render("Profile: ")+u.ProfileURL,
		styleMeta.Render("GitHub:  ")+u.GitHubURL,
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		MaxWidth(width - 2).
		Render(inner)
}

// Pagination prints a hint when more results are available.
func Pagination(nextCursor int64) {
	fmt.Println(stylePaginate.Render(fmt.Sprintf("↓ more results: --before %d", nextCursor)))
	fmt.Println()
}

// timeAgo returns a short human-readable relative time string.
func timeAgo(t time.Time) string {
	return timeAgoFrom(t, time.Now())
}

func timeAgoFrom(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
