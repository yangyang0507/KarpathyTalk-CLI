package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *client.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return client.New(srv.URL)
}

func TestLoadTimeline_PassesBeforeAndLimit(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/posts" {
			t.Fatalf("path: got %q, want %q", r.URL.Path, "/api/posts")
		}
		if got := r.URL.Query().Get("before"); got != "123" {
			t.Fatalf("before: got %q, want %q", got, "123")
		}
		if got := r.URL.Query().Get("limit"); got != "7" {
			t.Fatalf("limit: got %q, want %q", got, "7")
		}
		if got := r.URL.Query().Get("has_parent"); got != "false" {
			t.Fatalf("has_parent: got %q, want %q", got, "false")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(client.PostsResponse{})
	})

	msg := LoadTimeline(c, 123, 7)()
	if _, ok := msg.(PostsLoadedMsg); !ok {
		t.Fatalf("got %T, want PostsLoadedMsg", msg)
	}
}

func TestLoadUserPosts_PassesRepliesBeforeAndLimit(t *testing.T) {
	var postsSeen, userSeen bool
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/users/karpathy":
			userSeen = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(client.User{Username: "karpathy"})
		case "/api/posts":
			postsSeen = true
			if got := r.URL.Query().Get("author"); got != "karpathy" {
				t.Fatalf("author: got %q, want %q", got, "karpathy")
			}
			if got := r.URL.Query().Get("before"); got != "456" {
				t.Fatalf("before: got %q, want %q", got, "456")
			}
			if got := r.URL.Query().Get("limit"); got != "9" {
				t.Fatalf("limit: got %q, want %q", got, "9")
			}
			if got := r.URL.Query().Get("has_parent"); got != "true" {
				t.Fatalf("has_parent: got %q, want %q", got, "true")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(client.PostsResponse{})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	})

	msg := LoadUserPosts(c, "karpathy", 456, 9, true)()
	if _, ok := msg.(PostsLoadedMsg); !ok {
		t.Fatalf("got %T, want PostsLoadedMsg", msg)
	}
	if !userSeen || !postsSeen {
		t.Fatalf("expected both user and posts requests, got user=%v posts=%v", userSeen, postsSeen)
	}
}

func TestLoadMoreUserPosts_PassesRepliesFlag(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("has_parent"); got != "false" {
			t.Fatalf("has_parent: got %q, want %q", got, "false")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(client.PostsResponse{})
	})

	msg := LoadMoreUserPosts(c, "karpathy", 0, 20, false)()
	if _, ok := msg.(PostsLoadedMsg); !ok {
		t.Fatalf("got %T, want PostsLoadedMsg", msg)
	}
}

func TestLoadPost_PassesReplyLimit(t *testing.T) {
	var postSeen, repliesSeen bool
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/posts/42":
			postSeen = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(client.PostResponse{Post: client.Post{ID: 42}})
		case "/api/posts":
			repliesSeen = true
			if got := r.URL.Query().Get("parent_post_id"); got != "42" {
				t.Fatalf("parent_post_id: got %q, want %q", got, "42")
			}
			if got := r.URL.Query().Get("limit"); got != "13" {
				t.Fatalf("limit: got %q, want %q", got, "13")
			}
			if got := r.URL.Query().Get("has_parent"); got != "true" {
				t.Fatalf("has_parent: got %q, want %q", got, "true")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(client.PostsResponse{})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	})

	msg := LoadPost(c, 42, 13)()
	if _, ok := msg.(PostLoadedMsg); !ok {
		t.Fatalf("got %T, want PostLoadedMsg", msg)
	}
	if !postSeen || !repliesSeen {
		t.Fatalf("expected both post and replies requests, got post=%v replies=%v", postSeen, repliesSeen)
	}
}
