package main

import (
	"testing"

	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
)

func TestPostJSONOutput_IncludesReplyPagination(t *testing.T) {
	post := &client.Post{ID: 42}
	repliesResp := &client.PostsResponse{
		Posts:      []client.Post{{ID: 1001}},
		HasMore:    true,
		NextCursor: 1001,
	}

	got := postJSONOutput(post, repliesResp)

	if got["post"] != post {
		t.Fatalf("post: got %#v, want %#v", got["post"], post)
	}
	replies, ok := got["replies"].([]client.Post)
	if !ok {
		t.Fatalf("replies type: got %T", got["replies"])
	}
	if len(replies) != 1 || replies[0].ID != 1001 {
		t.Fatalf("replies: got %#v", replies)
	}
	if got["has_more"] != true {
		t.Fatalf("has_more: got %#v, want true", got["has_more"])
	}
	if got["next_cursor"] != int64(1001) {
		t.Fatalf("next_cursor: got %#v, want 1001", got["next_cursor"])
	}
}
