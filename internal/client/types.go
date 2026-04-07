package client

import (
	"encoding/json"
	"time"
)

// JSONTime wraps time.Time to support RFC3339 JSON serialization.
type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(time.RFC3339))
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	*t = JSONTime(parsed)
	return nil
}

func (t JSONTime) Time() time.Time { return time.Time(t) }

// User corresponds to apiUser (GET /api/users/{username}).
type User struct {
	ID             int64    `json:"id"`
	Username       string   `json:"username"`
	DisplayName    string   `json:"display_name"`
	AvatarURL      string   `json:"avatar_url"`
	GitHubURL      string   `json:"github_url"`
	ProfileURL     string   `json:"profile_url"`
	FeedURL        string   `json:"feed_url"`
	CreatedAt      JSONTime `json:"created_at"`
	FollowerCount  int      `json:"follower_count"`
	FollowingCount int      `json:"following_count"`
	PostCount      int      `json:"post_count"`
}

// UserRef is the embedded author reference inside a Post.
type UserRef struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	AvatarURL   string   `json:"avatar_url"`
	ProfileURL  string   `json:"profile_url"`
	GitHubURL   string   `json:"github_url"`
	CreatedAt   JSONTime `json:"created_at"`
}

// PostRef is a lightweight reference to another post.
type PostRef struct {
	ID             int64   `json:"id"`
	URL            string  `json:"url"`
	Author         UserRef `json:"author"`
	RevisionNumber int     `json:"revision_number"`
	RevisionCount  int     `json:"revision_count"`
}

// Post corresponds to apiPost.
type Post struct {
	ID                   int64    `json:"id"`
	URL                  string   `json:"url"`
	Author               UserRef  `json:"author"`
	ContentMarkdown      string   `json:"content_markdown"`
	CreatedAt            JSONTime `json:"created_at"`
	EditedAt             *JSONTime `json:"edited_at,omitempty"`
	LikeCount            int      `json:"like_count"`
	RepostCount          int      `json:"repost_count"`
	ReplyCount           int      `json:"reply_count"`
	RevisionID           int64    `json:"revision_id"`
	RevisionNumber       int      `json:"revision_number"`
	RevisionCount        int      `json:"revision_count"`
	RevisionCreatedAt    JSONTime `json:"revision_created_at"`
	ParentPostID         *int64   `json:"parent_post_id,omitempty"`
	ParentPostRevisionID *int64   `json:"parent_post_revision_id,omitempty"`
	QuoteOfID            *int64   `json:"quote_of_id,omitempty"`
	QuoteOfRevisionID    *int64   `json:"quote_of_revision_id,omitempty"`
	Depth                int      `json:"depth,omitempty"`
	ParentPost           *PostRef `json:"parent_post,omitempty"`
	QuotedPost           *PostRef `json:"quoted_post,omitempty"`
}

// PostsResponse corresponds to apiPostsQueryResponse (GET /api/posts).
type PostsResponse struct {
	Posts      []Post `json:"posts"`
	HasMore    bool   `json:"has_more"`
	NextCursor int64  `json:"next_cursor,omitempty"`
}

// PostResponse corresponds to apiPostResponse (GET /api/posts/{id}).
type PostResponse struct {
	Post Post `json:"post"`
}
