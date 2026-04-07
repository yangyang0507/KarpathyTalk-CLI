package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/display"
)

// DetailModel shows a single post with replies in a scrollable viewport.
type DetailModel struct {
	post    client.Post
	replies []client.Post
	vp      viewport.Model
	width   int
	height  int
	ready   bool
}

func newDetailModel(post client.Post, replies []client.Post) DetailModel {
	return DetailModel{post: post, replies: replies}
}

func (m DetailModel) Init() tea.Cmd { return nil }

func (m DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "backspace":
			return m, func() tea.Msg { return backToListMsg{} }
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		statusH := 1
		if !m.ready {
			m.vp = viewport.New(m.width, m.height-statusH)
			m.vp.SetContent(m.buildContent())
			m.ready = true
		} else {
			m.vp.Width = m.width
			m.vp.Height = m.height - statusH
			m.vp.SetContent(m.buildContent())
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m DetailModel) View() string {
	if !m.ready {
		return styleLoading.Render("  loading…")
	}

	pct := int(m.vp.ScrollPercent() * 100)
	status := styleStatus.Render(fmt.Sprintf(
		"  ↑↓/j/k scroll  esc/q back   %d%%", pct,
	))
	return m.vp.View() + "\n" + status
}

func (m DetailModel) buildContent() string {
	var sb strings.Builder
	sb.WriteString(display.RenderFull(m.post, m.width))

	if len(m.replies) > 0 {
		sb.WriteByte('\n')
		label := fmt.Sprintf("── %d direct repl", len(m.replies))
		if len(m.replies) == 1 {
			label += "y"
		} else {
			label += "ies"
		}
		label += " ──"
		sb.WriteString(styleStatus.Render(label))
		sb.WriteString("\n\n")

		for _, r := range m.replies {
			sb.WriteString(display.RenderSummaryCard(r, m.width))
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}
