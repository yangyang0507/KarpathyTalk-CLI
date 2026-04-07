package client

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONTime_UnmarshalJSON(t *testing.T) {
	input := `"2024-01-15T10:30:00Z"`
	var jt JSONTime
	if err := json.Unmarshal([]byte(input), &jt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := jt.Time()
	want := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestJSONTime_UnmarshalJSON_Invalid(t *testing.T) {
	cases := []string{
		`"not-a-date"`,
		`12345`,
		`null`,
		`""`,
	}
	for _, c := range cases {
		var jt JSONTime
		if err := json.Unmarshal([]byte(c), &jt); err == nil {
			t.Errorf("expected error for input %s, got nil", c)
		}
	}
}

func TestJSONTime_MarshalJSON(t *testing.T) {
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	jt := JSONTime(ts)
	data, err := json.Marshal(jt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Roundtrip: unmarshal back and compare
	var jt2 JSONTime
	if err := json.Unmarshal(data, &jt2); err != nil {
		t.Fatalf("roundtrip unmarshal error: %v", err)
	}
	if !jt2.Time().Equal(ts) {
		t.Errorf("roundtrip got %v, want %v", jt2.Time(), ts)
	}
}

func TestPostsResponse_Unmarshal(t *testing.T) {
	raw := `{
		"posts": [
			{
				"id": 42,
				"url": "https://example.com/posts/42",
				"author": {
					"id": 1,
					"username": "alice",
					"display_name": "Alice",
					"avatar_url": "",
					"profile_url": "",
					"github_url": "",
					"created_at": "2024-01-01T00:00:00Z"
				},
				"content_markdown": "hello world",
				"created_at": "2024-01-15T10:30:00Z",
				"like_count": 3,
				"repost_count": 1,
				"reply_count": 2,
				"revision_id": 10,
				"revision_number": 1,
				"revision_count": 1,
				"revision_created_at": "2024-01-15T10:30:00Z"
			}
		],
		"has_more": true,
		"next_cursor": 41
	}`

	var resp PostsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(resp.Posts))
	}
	p := resp.Posts[0]
	if p.ID != 42 {
		t.Errorf("ID: got %d, want 42", p.ID)
	}
	if p.Author.Username != "alice" {
		t.Errorf("Author.Username: got %q, want %q", p.Author.Username, "alice")
	}
	if p.ContentMarkdown != "hello world" {
		t.Errorf("ContentMarkdown: got %q, want %q", p.ContentMarkdown, "hello world")
	}
	if p.LikeCount != 3 {
		t.Errorf("LikeCount: got %d, want 3", p.LikeCount)
	}
	if !resp.HasMore {
		t.Error("HasMore: got false, want true")
	}
	if resp.NextCursor != 41 {
		t.Errorf("NextCursor: got %d, want 41", resp.NextCursor)
	}
}

func TestPost_OptionalFields(t *testing.T) {
	parentID := int64(10)
	raw := `{
		"id": 99,
		"url": "",
		"author": {"id":1,"username":"bob","display_name":"Bob","avatar_url":"","profile_url":"","github_url":"","created_at":"2024-01-01T00:00:00Z"},
		"content_markdown": "reply",
		"created_at": "2024-01-15T10:30:00Z",
		"like_count": 0, "repost_count": 0, "reply_count": 0,
		"revision_id": 1, "revision_number": 1, "revision_count": 1,
		"revision_created_at": "2024-01-15T10:30:00Z",
		"parent_post_id": 10
	}`
	var p Post
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ParentPostID == nil {
		t.Fatal("ParentPostID: expected non-nil")
	}
	if *p.ParentPostID != parentID {
		t.Errorf("ParentPostID: got %d, want %d", *p.ParentPostID, parentID)
	}
	if p.QuoteOfID != nil {
		t.Error("QuoteOfID: expected nil")
	}
}
