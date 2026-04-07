package tui

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"kt/internal/client"
)

// PostsLoadedMsg is sent when a list of posts is fetched.
type PostsLoadedMsg struct {
	Posts      []client.Post
	HasMore    bool
	NextCursor int64
	User       *client.User // non-nil only on initial user-mode load
}

// PostLoadedMsg is sent when a single post (with replies) is fetched.
type PostLoadedMsg struct {
	Post    client.Post
	Replies []client.Post
}

// ErrMsg is sent when an API call fails.
type ErrMsg struct{ Err error }

// LoadTimeline fetches the public timeline.
func LoadTimeline(c *client.Client, before int64, limit int) tea.Cmd {
	return func() tea.Msg {
		hasParent := false
		resp, err := c.GetPosts(client.PostsQuery{
			HasParent: &hasParent,
			Before:    before,
			Limit:     limit,
		})
		if err != nil {
			return ErrMsg{err}
		}
		return PostsLoadedMsg{Posts: resp.Posts, HasMore: resp.HasMore, NextCursor: resp.NextCursor}
	}
}

// LoadUserPosts fetches a user's profile and their posts concurrently.
func LoadUserPosts(c *client.Client, username string, before int64, limit int) tea.Cmd {
	return func() tea.Msg {
		var (
			user    *client.User
			userErr error
			resp    *client.PostsResponse
			respErr error
			wg      sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			user, userErr = c.GetUser(username)
		}()
		go func() {
			defer wg.Done()
			hasParent := false
			resp, respErr = c.GetPosts(client.PostsQuery{
				Author:    username,
				HasParent: &hasParent,
				Before:    before,
				Limit:     limit,
			})
		}()
		wg.Wait()
		if userErr != nil {
			return ErrMsg{userErr}
		}
		if respErr != nil {
			return ErrMsg{respErr}
		}
		return PostsLoadedMsg{Posts: resp.Posts, HasMore: resp.HasMore, NextCursor: resp.NextCursor, User: user}
	}
}

// LoadMoreUserPosts fetches additional posts for a user (no user profile refetch).
func LoadMoreUserPosts(c *client.Client, username string, before int64, limit int) tea.Cmd {
	return func() tea.Msg {
		hasParent := false
		resp, err := c.GetPosts(client.PostsQuery{
			Author:    username,
			HasParent: &hasParent,
			Before:    before,
			Limit:     limit,
		})
		if err != nil {
			return ErrMsg{err}
		}
		return PostsLoadedMsg{Posts: resp.Posts, HasMore: resp.HasMore, NextCursor: resp.NextCursor}
	}
}

// LoadPost fetches a single post and its direct replies concurrently.
func LoadPost(c *client.Client, id int64) tea.Cmd {
	return func() tea.Msg {
		var (
			post    *client.Post
			postErr error
			replies []client.Post
			repErr  error
			wg      sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			post, postErr = c.GetPost(id)
		}()
		go func() {
			defer wg.Done()
			hasParent := true
			resp, err := c.GetPosts(client.PostsQuery{
				ParentPostID: &id,
				HasParent:    &hasParent,
				Limit:        50,
			})
			if err != nil {
				repErr = err
				return
			}
			replies = resp.Posts
		}()
		wg.Wait()
		if postErr != nil {
			return ErrMsg{postErr}
		}
		if repErr != nil {
			return ErrMsg{repErr}
		}
		return PostLoadedMsg{Post: *post, Replies: replies}
	}
}
