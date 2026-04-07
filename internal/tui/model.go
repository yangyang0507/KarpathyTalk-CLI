package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/display"
)

type viewState int

const (
	viewLoading viewState = iota
	viewList
	viewDetail
	viewError
)

// Config is passed from main to decouple CLI flag parsing from the TUI.
type Config struct {
	Mode      string // "timeline" | "user" | "post"
	Username  string // for "user" mode
	PostID    int64  // for "post" mode
	Limit     int
	Client    *client.Client
}

// AppModel is the root Bubble Tea model.
type AppModel struct {
	state      viewState
	list       ListModel
	detail     DetailModel
	cfg        Config
	errMsg     string
	cachedUser *client.User // stored for user-mode header re-render on resize
}

func (m AppModel) Init() tea.Cmd {
	switch m.cfg.Mode {
	case "user":
		return LoadUserPosts(m.cfg.Client, m.cfg.Username, 0, m.cfg.Limit)
	case "post":
		return LoadPost(m.cfg.Client, m.cfg.PostID)
	default: // "timeline"
		return LoadTimeline(m.cfg.Client, 0, m.cfg.Limit)
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		// Propagate size to active sub-model.
		switch m.state {
		case viewList:
			updated, cmd := m.list.Update(msg)
			m.list = updated.(ListModel)
			// Re-render user profile box at the new terminal width.
			if m.cfg.Mode == "user" && m.list.userHeader != "" && m.cachedUser != nil {
				m.list.userHeader = display.RenderUserProfile(*m.cachedUser, msg.Width)
			}
			return m, cmd
		case viewDetail:
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		case viewLoading:
			// Store dimensions so they're ready when we transition.
			m.list.width, m.list.height = msg.Width, msg.Height
			m.detail.width, m.detail.height = msg.Width, msg.Height
		}
		return m, nil

	case PostsLoadedMsg:
		if m.state == viewLoading {
			// Initial load: set up list model.
			m.list = ListModel{
				posts:      msg.Posts,
				hasMore:    msg.HasMore,
				nextCursor: msg.NextCursor,
				apiClient:  m.cfg.Client,
				query: listQuery{
					mode:     m.cfg.Mode,
					username: m.cfg.Username,
					limit:    m.cfg.Limit,
				},
				width:  m.list.width,
				height: m.list.height,
			}
			if msg.User != nil {
				m.cachedUser = msg.User
				m.list.userHeader = display.RenderUserProfile(*msg.User, m.list.width)
			}
			m.state = viewList
			return m, nil
		}
		// Pagination append — pass to list.
		updated, cmd := m.list.Update(msg)
		m.list = updated.(ListModel)
		return m, cmd

	case PostLoadedMsg:
		m.detail = newDetailModel(msg.Post, msg.Replies)
		m.state = viewDetail
		// Use list dimensions — they're always updated by WindowSizeMsg regardless of prior state.
		initSize := tea.WindowSizeMsg{Width: m.list.width, Height: m.list.height}
		updated, cmd := m.detail.Update(initSize)
		m.detail = updated.(DetailModel)
		return m, cmd

	case selectPostMsg:
		// User pressed Enter on a list item: fetch full post + replies.
		m.state = viewLoading
		return m, LoadPost(m.cfg.Client, msg.post.ID)

	case backToListMsg:
		if m.cfg.Mode == "post" {
			// Entered directly in post mode — quit instead of going back.
			return m, tea.Quit
		}
		m.state = viewList
		return m, nil

	case ErrMsg:
		m.state = viewError
		m.errMsg = msg.Err.Error()
		return m, nil
	}

	// Delegate to active sub-model.
	switch m.state {
	case viewList:
		updated, cmd := m.list.Update(msg)
		m.list = updated.(ListModel)
		return m, cmd
	case viewDetail:
		updated, cmd := m.detail.Update(msg)
		m.detail = updated.(DetailModel)
		return m, cmd
	}
	return m, nil
}

func (m AppModel) View() string {
	switch m.state {
	case viewLoading:
		return styleLoading.Render("\n  loading…")
	case viewList:
		return m.list.View()
	case viewDetail:
		return m.detail.View()
	case viewError:
		return styleErrStyle.Render("\n  error: " + m.errMsg + "\n\n  press q to quit")
	}
	return ""
}

// Run starts the interactive TUI with an alternate screen.
func Run(cfg Config) error {
	if cfg.Limit <= 0 {
		cfg.Limit = 20
	}
	m := AppModel{
		state: viewLoading,
		cfg:   cfg,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
