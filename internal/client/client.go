package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultHost = "https://karpathytalk.com"

// Client is the HTTP client for the KarpathyTalk API.
type Client struct {
	Host       string
	HTTPClient *http.Client
}

// New creates a new Client with the given host (e.g. "https://karpathytalk.com").
func New(host string) *Client {
	host = strings.TrimRight(host, "/")
	return &Client{
		Host: host,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PostsQuery holds query parameters for /api/posts.
type PostsQuery struct {
	Author       string
	HasParent    *bool
	ParentPostID *int64
	Before       int64
	Limit        int
}

// GetPosts calls GET /api/posts with the given filters.
func (c *Client) GetPosts(q PostsQuery) (*PostsResponse, error) {
	params := url.Values{}
	if q.Author != "" {
		params.Set("author", q.Author)
	}
	if q.HasParent != nil {
		if *q.HasParent {
			params.Set("has_parent", "true")
		} else {
			params.Set("has_parent", "false")
		}
	}
	if q.ParentPostID != nil {
		params.Set("parent_post_id", strconv.FormatInt(*q.ParentPostID, 10))
	}
	if q.Before > 0 {
		params.Set("before", strconv.FormatInt(q.Before, 10))
	}
	if q.Limit > 0 {
		params.Set("limit", strconv.Itoa(q.Limit))
	}

	endpoint := "/api/posts"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	var resp PostsResponse
	if err := c.getJSON(endpoint, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPost calls GET /api/posts/{id}.
func (c *Client) GetPost(id int64) (*Post, error) {
	var resp PostResponse
	if err := c.getJSON(fmt.Sprintf("/api/posts/%d", id), &resp); err != nil {
		return nil, err
	}
	return &resp.Post, nil
}

// GetUser calls GET /api/users/{username}.
func (c *Client) GetUser(username string) (*User, error) {
	var user User
	if err := c.getJSON(fmt.Sprintf("/api/users/%s", url.PathEscape(username)), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetPostMarkdown calls GET /posts/{id}/md (with optional ?revision=N).
func (c *Client) GetPostMarkdown(id int64, revision int) (string, error) {
	path := fmt.Sprintf("/posts/%d/md", id)
	if revision > 0 {
		path += "?revision=" + strconv.Itoa(revision)
	}
	return c.getText(path)
}

// GetPostRaw calls GET /posts/{id}/raw (with optional ?revision=N).
func (c *Client) GetPostRaw(id int64, revision int) (string, error) {
	path := fmt.Sprintf("/posts/%d/raw", id)
	if revision > 0 {
		path += "?revision=" + strconv.Itoa(revision)
	}
	return c.getText(path)
}

// GetUserMarkdown calls GET /user/{username}/md.
func (c *Client) GetUserMarkdown(username string) (string, error) {
	return c.getText(fmt.Sprintf("/user/%s/md", url.PathEscape(username)))
}

// GetDocs calls GET /docs.md.
func (c *Client) GetDocs() (string, error) {
	return c.getText("/docs.md")
}

// PostURL returns the canonical URL for a post.
func (c *Client) PostURL(id int64) string {
	return fmt.Sprintf("%s/posts/%d", c.Host, id)
}

// userAgent is sent with every request so the server can identify the client.
var userAgent = "kt/" + "dev" // overridden at link time via -ldflags

// SetVersion updates the User-Agent string. Called from main after version is set.
func SetVersion(v string) { userAgent = "kt/" + v }

func (c *Client) newRequest(path string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, c.Host+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

func (c *Client) getJSON(path string, out any) error {
	req, err := c.newRequest(path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

func (c *Client) getText(path string) (string, error) {
	req, err := c.newRequest(path)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return string(body), nil
}
