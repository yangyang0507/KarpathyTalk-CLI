package tui

import (
	"regexp"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/display"
)

// cardHeight is the number of terminal lines each post card occupies.
// header line + 2 preview lines + separator line + inter-card newline = 5.
const cardHeight = 5

// internal message types
type selectPostMsg struct{ post client.Post }
type backToListMsg struct{}

// listQuery stores the original query parameters for pagination.
type listQuery struct {
	mode     string // "timeline" | "user"
	username string
	limit    int
}

// ListModel is the scrollable post list view.
type ListModel struct {
	posts      []client.Post
	cursor     int
	offset     int
	width      int
	height     int
	hasMore    bool
	nextCursor int64
	loading    bool
	err        error
	apiClient  *client.Client
	query      listQuery
	userHeader string // rendered user profile box for "user" mode
}

func (m ListModel) Init() tea.Cmd { return nil }

func (m ListModel) visibleCount() int {
	h := m.height - 2 // reserve 2 lines for status bar
	if m.userHeader != "" {
		// subtract lines taken by the user profile box
		h -= strings.Count(m.userHeader, "\n") + 2
	}
	n := h / cardHeight
	if n < 1 {
		n = 1
	}
	return n
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.posts)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if len(m.posts) > 0 {
				return m, func() tea.Msg { return selectPostMsg{m.posts[m.cursor]} }
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		// Keep cursor within visible window
		visible := m.visibleCount()
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
		if m.cursor >= m.offset+visible {
			m.offset = m.cursor - visible + 1
		}

		// Auto-load more when near the end
		if m.cursor >= len(m.posts)-3 && m.hasMore && !m.loading {
			m.loading = true
			return m, m.loadMoreCmd()
		}

	case PostsLoadedMsg:
		m.posts = append(m.posts, msg.Posts...)
		m.hasMore = msg.HasMore
		m.nextCursor = msg.NextCursor
		m.loading = false

	case ErrMsg:
		m.err = msg.Err
		m.loading = false
	}
	return m, nil
}

func (m ListModel) loadMoreCmd() tea.Cmd {
	switch m.query.mode {
	case "user":
		return LoadMoreUserPosts(m.apiClient, m.query.username, m.nextCursor, m.query.limit)
	default:
		return LoadTimeline(m.apiClient, m.nextCursor, m.query.limit)
	}
}

func (m ListModel) View() string {
	if m.width == 0 {
		return styleLoading.Render("  loading...")
	}
	if m.err != nil {
		return styleErrStyle.Render("  error: " + m.err.Error())
	}

	var sb strings.Builder

	// User profile header (user mode only)
	if m.userHeader != "" {
		sb.WriteString(m.userHeader)
		sb.WriteString("\n\n")
	}

	visible := m.visibleCount()
	end := m.offset + visible
	if end > len(m.posts) {
		end = len(m.posts)
	}

	for i := m.offset; i < end; i++ {
		sb.WriteString(m.renderCard(i))
		sb.WriteByte('\n')
	}

	// Status bar
	status := "  j/k ↑↓ navigate  Enter open  q quit"
	if m.loading {
		status += "   " + styleLoading.Render("[loading more…]")
	}
	sb.WriteString(styleStatus.Render(status))

	return sb.String()
}

func (m ListModel) renderCard(idx int) string {
	p := m.posts[idx]
	w := m.width

	header := display.RenderHeader(p, w)
	line1, line2 := cardPreview(display.StripMarkdownImages(p.ContentMarkdown), w)
	sepLine := styleSep.Render(strings.Repeat("─", w))

	selected := idx == m.cursor

	if selected {
		header = padToWidth(header, w)
		line1 = padToWidth(line1, w)
		line2 = padToWidth(line2, w)
		header = styleSelected.Render(header)
		line1 = styleSelected.Render(line1)
		line2 = styleSelected.Render(line2)
	} else {
		line1 = stylePreview.Render(line1)
		line2 = stylePreview.Render(line2)
	}

	return strings.Join([]string{header, line1, line2, sepLine}, "\n")
}

// padToWidth pads an ANSI string to the given visual width.
func padToWidth(s string, width int) string {
	current := lipgloss.Width(s)
	if current >= width {
		return s
	}
	return s + strings.Repeat(" ", width-current)
}

// cardPreview extracts two plain-text preview lines from markdown content.
func cardPreview(markdown string, width int) (string, string) {
	text := stripMarkdownSyntax(markdown)
	var nonEmpty []string
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			nonEmpty = append(nonEmpty, l)
		}
	}

	maxLen := width - 4
	if maxLen < 10 {
		maxLen = 10
	}

	get := func(i int) string {
		if i >= len(nonEmpty) {
			return ""
		}
		l := nonEmpty[i]
		if utf8.RuneCountInString(l) > maxLen {
			runes := []rune(l)
			return "  " + string(runes[:maxLen-3]) + "..."
		}
		return "  " + l
	}

	return get(0), get(1)
}

var (
	reBold    = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reItalic  = regexp.MustCompile(`\*(.+?)\*`)
	reCode    = regexp.MustCompile("`(.+?)`")
	reLink    = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	reHeading = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	reCodeBlk = regexp.MustCompile("(?s)```.*?```")
)

// stripMarkdownSyntax removes common markdown syntax for plain-text preview.
func stripMarkdownSyntax(text string) string {
	text = reCodeBlk.ReplaceAllString(text, "")
	text = reHeading.ReplaceAllString(text, "")
	text = reBold.ReplaceAllString(text, "$1")
	text = reItalic.ReplaceAllString(text, "$1")
	text = reCode.ReplaceAllString(text, "$1")
	text = reLink.ReplaceAllString(text, "$1")
	return text
}
