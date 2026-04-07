package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestServer creates an httptest.Server and returns it along with a Client
// pointed at it. The handler func receives the request URL for assertions.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, New(srv.URL)
}

func jsonHandler(t *testing.T, v any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}
}

func textHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(body))
	}
}

// ── GetPosts ─────────────────────────────────────────────────────────────────

func TestGetPosts_Basic(t *testing.T) {
	want := PostsResponse{
		Posts:   []Post{{ID: 1, ContentMarkdown: "hello"}},
		HasMore: false,
	}
	_, c := newTestServer(t, jsonHandler(t, want))

	resp, err := c.GetPosts(PostsQuery{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(resp.Posts))
	}
	if resp.Posts[0].ID != 1 {
		t.Errorf("post ID: got %d, want 1", resp.Posts[0].ID)
	}
}

func TestGetPosts_QueryParams(t *testing.T) {
	var gotQuery url.Values
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{}})
	})
	_ = srv

	hasParent := false
	_, err := c.GetPosts(PostsQuery{
		Author:    "karpathy",
		HasParent: &hasParent,
		Before:    100,
		Limit:     5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := map[string]string{
		"author":     "karpathy",
		"has_parent": "false",
		"before":     "100",
		"limit":      "5",
	}
	for k, want := range checks {
		if got := gotQuery.Get(k); got != want {
			t.Errorf("query param %q: got %q, want %q", k, got, want)
		}
	}
}

func TestGetPosts_HasParentTrue(t *testing.T) {
	var gotQuery url.Values
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{}})
	})

	hasParent := true
	parentID := int64(42)
	c.GetPosts(PostsQuery{HasParent: &hasParent, ParentPostID: &parentID})

	if got := gotQuery.Get("has_parent"); got != "true" {
		t.Errorf("has_parent: got %q, want %q", got, "true")
	}
	if got := gotQuery.Get("parent_post_id"); got != "42" {
		t.Errorf("parent_post_id: got %q, want %q", got, "42")
	}
}

// ── GetPost ───────────────────────────────────────────────────────────────────

func TestGetPost(t *testing.T) {
	want := PostResponse{Post: Post{ID: 7, ContentMarkdown: "test post"}}
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(want)
	})

	post, err := c.GetPost(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.ID != 7 {
		t.Errorf("ID: got %d, want 7", post.ID)
	}
	if gotPath != "/api/posts/7" {
		t.Errorf("path: got %q, want %q", gotPath, "/api/posts/7")
	}
}

// ── GetUser ───────────────────────────────────────────────────────────────────

func TestGetUser(t *testing.T) {
	want := User{ID: 1, Username: "karpathy", DisplayName: "Andrej"}
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(want)
	})

	user, err := c.GetUser("karpathy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "karpathy" {
		t.Errorf("Username: got %q, want %q", user.Username, "karpathy")
	}
	if gotPath != "/api/users/karpathy" {
		t.Errorf("path: got %q, want %q", gotPath, "/api/users/karpathy")
	}
}

// ── Text endpoints ────────────────────────────────────────────────────────────

func TestGetPostMarkdown(t *testing.T) {
	body := "# Hello\nsome content"
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		textHandler(body)(w, r)
	})

	got, err := c.GetPostMarkdown(42, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != body {
		t.Errorf("body: got %q, want %q", got, body)
	}
	if gotPath != "/posts/42/md" {
		t.Errorf("path: got %q, want %q", gotPath, "/posts/42/md")
	}
}

func TestGetPostMarkdown_WithRevision(t *testing.T) {
	var gotQuery url.Values
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		textHandler("content")(w, r)
	})

	c.GetPostMarkdown(42, 3)
	if got := gotQuery.Get("revision"); got != "3" {
		t.Errorf("revision: got %q, want %q", got, "3")
	}
}

func TestGetPostRaw(t *testing.T) {
	body := "raw content"
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		textHandler(body)(w, r)
	})

	got, err := c.GetPostRaw(5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != body {
		t.Errorf("body: got %q, want %q", got, body)
	}
	if gotPath != "/posts/5/raw" {
		t.Errorf("path: got %q, want %q", gotPath, "/posts/5/raw")
	}
}

func TestGetDocs(t *testing.T) {
	body := "# Docs\nsome docs"
	_, c := newTestServer(t, textHandler(body))

	got, err := c.GetDocs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != body {
		t.Errorf("body: got %q, want %q", got, body)
	}
}

func TestGetUserMarkdown(t *testing.T) {
	body := "---\nusername: alice\n---\nprofile"
	var gotPath string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		textHandler(body)(w, r)
	})

	got, err := c.GetUserMarkdown("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != body {
		t.Errorf("body mismatch")
	}
	if gotPath != "/user/alice/md" {
		t.Errorf("path: got %q, want %q", gotPath, "/user/alice/md")
	}
}

// ── Error handling ────────────────────────────────────────────────────────────

func TestClient_NotFound(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 page not found", http.StatusNotFound)
	})

	_, err := c.GetPost(999)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestClient_ServerError(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	_, err := c.GetPosts(PostsQuery{})
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	_, c := newTestServer(t, textHandler("not json"))

	_, err := c.GetPosts(PostsQuery{})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestNew_TrimsTrailingSlash(t *testing.T) {
	c := New("https://example.com/")
	if c.Host != "https://example.com" {
		t.Errorf("Host: got %q, want %q", c.Host, "https://example.com")
	}
}
