# kt — KarpathyTalk CLI Specification

## Design Philosophy

The following principles are derived from the KarpathyTalk server design. **The CLI must follow them strictly.**

### 1. Everything Is a Post

The data model has only two entities: **User** and **Post**. Everything else is derived:

- Reply = a post with `parent_post_id`
- Quote = a post with `quote_of_id`
- Timeline = a query against `/api/posts`
- User profile = a query against `/api/posts?author=<u>`

The command design reflects this model: `timeline`, `user`, and `post` are different filtered views of the same collection, not separate object systems.

### 2. One Collection Interface, Orthogonal Filters

The server exposes a single list endpoint `/api/posts`, covering all scenarios through orthogonal query parameters:

```
author=<u>  has_parent=true|false  parent_post_id=<id>  before=<id>  limit=<n>
```

The CLI **should not invent new abstractions** for each scenario — it maps directly to this filter language.

### 3. Two Read Modes: JSON for Code, Markdown for Humans and Agents

The server distinguishes two output modes explicitly:
- **JSON** — structured fields, pagination, relationships; for programmatic consumption
- **Markdown** (with YAML frontmatter) — human-readable documents; for people and LLM agents

The `--json` / `--markdown` flags map directly to these modes. They are **not optional decorations** — they are core to the design. The default human output is a third mode: a terminal-friendly summary view.

### 4. Open Reads, Manual Writes

The platform has no public write API. The CLI is therefore a **purely read-only tool**. Authentication, write operations, and bot posting are out of scope and should not be accommodated.

### 5. content_markdown Is the Source of Truth

The server expands relative URLs to absolute URLs in `content_markdown` within JSON responses, making the content usable outside the platform. The CLI **uses `content_markdown` directly** for display and export, never `content_html`, ensuring portability.

### 6. Minimal Dependencies

The server uses only the Go standard library, goldmark, and SQLite. The CLI follows the same principle: **standard library first; every external dependency requires justification.**

### 7. Composability

The CLI is a Unix tool and must be composable via pipes:

```bash
kt timeline --json | jq '.posts[].author.username'
kt post 42 --markdown | llm "summarize this"
kt user karpathy --json | jq '.posts | length'
```

This requires `--json` output to be **strictly pure JSON** — no prompts or messages mixed into stdout (all such output goes to stderr).

---

## Overview

`kt` is a command-line read client for KarpathyTalk, written in Go. Its goal is to let developers and LLM agents browse KarpathyTalk content fluently from the terminal.

---

## Directory Structure

```
cmd/
  kt/
    main.go              ← entry point, top-level command dispatch
internal/
  client/
    client.go            ← HTTP client wrapping all API requests
    types.go             ← API response structs (mirroring server api.go)
  display/
    human.go             ← human-friendly formatted terminal output
    json.go              ← raw JSON passthrough output
    markdown.go          ← Markdown output
    render.go            ← Markdown rendering helpers (glamour, width)
  tui/
    model.go             ← root Bubble Tea model and state machine
    list.go              ← scrollable post list view
    detail.go            ← full post detail view with viewport
    load.go              ← async data-loading commands
    styles.go            ← TUI-specific lipgloss styles
```

> `internal/client` is fully decoupled from the server and depends only on the public API. It can be reused by future desktop or mobile clients.

---

## Phase 1 Features

### Commands

#### `kt timeline`

Browse the public timeline (latest root posts only).

```
kt timeline [flags]

Flags:
  --limit <n>      Posts per page; default 20, max 100
  --before <id>    Pagination cursor; load posts before this ID
  --json           Output raw JSON
  --markdown       Output post content as Markdown
```

---

#### `kt user <username>`

View a user's profile and their posts.

```
kt user <username> [flags]

Flags:
  --replies        Show only the user's replies; default shows root posts
  --limit <n>      Posts per page; default 20
  --before <id>    Pagination cursor
  --json           Output raw JSON
  --markdown       Output user profile Markdown (with YAML frontmatter, calls /user/<u>/md)
```

---

#### `kt post <id>`

View a single post and its direct replies (two API calls: the post itself + direct reply list).

> The server does not currently register a `/api/posts/{id}/thread` endpoint. Nested replies require additional recursive requests. The current version shows only the first level of direct replies.

```
kt post <id> [flags]

Flags:
  --limit <n>      Replies per page; default 20
  --json           Output raw JSON
  --markdown       Output post Markdown with YAML frontmatter (calls /posts/<id>/md)
  --raw            Output raw post Markdown body, no frontmatter (calls /posts/<id>/raw)
  --revision <n>   View a specific revision (use with --markdown / --raw)
```

#### `kt docs`

Fetch the platform API documentation as Markdown. Useful for giving LLM agents context about platform capabilities.

```
kt docs
```

---

### Global Flag

```
  --host <url>     API root URL; default https://karpathytalk.com
                   Supports local instances: --host http://localhost:8080
```

---

## Phase 2 Features

### Commands

#### `kt watch <username>`

Poll a user and print terminal notifications when new posts appear.

```
kt watch <username> [flags]

Flags:
  --interval <s>   Poll interval in seconds; default 60
```

#### `kt open <id>`

Open a post in the system default browser.

```
kt open <id>
```

#### `kt export <username>`

Export all posts by a user to local Markdown files.

```
kt export <username> [flags]

Flags:
  --out <dir>      Output directory; default ./<username>
```

---

## API Mapping

Registered public read-only endpoints on the server (`app.go` route table):

| Command | API Endpoint | Notes |
|---|---|---|
| `kt timeline` | `GET /api/posts?has_parent=false` | `limit` and `before` supported |
| `kt user <u>` | `GET /api/users/<u>` + `GET /api/posts?author=<u>&has_parent=false` | Two calls; `has_parent=false` filters out replies |
| `kt user <u> --replies` | `GET /api/posts?author=<u>&has_parent=true` | |
| `kt user <u> --markdown` | `GET /user/<u>/md` | Returns user profile Markdown with YAML frontmatter |
| `kt post <id>` | `GET /api/posts/<id>` + `GET /api/posts?parent_post_id=<id>` | Direct replies only; `/api/posts/{id}/thread` not registered |
| `kt post <id> --markdown` | `GET /posts/<id>/md` | Supports `?revision=N` |
| `kt post <id> --raw` | `GET /posts/<id>/raw` | Supports `?revision=N` |
| `kt docs` | `GET /docs.md` | |
| `kt export <u>` | Paginate `GET /api/posts?author=<u>&has_parent=false` + `GET /posts/<id>/md` per post | |

---

## Data Structures (`client/types.go`)

Directly mirror the server's `internal/app/api.go` JSON responses. Key structs:

```go
type User struct { ... }          // maps to apiUser (GET /api/users/{username})
type UserRef struct { ... }       // maps to apiUserRef (embedded in Post)
type Post struct { ... }          // maps to apiPost
type PostsResponse struct { ... } // maps to apiPostsQueryResponse (GET /api/posts, no User field)
type PostResponse struct { ... }  // maps to apiPostResponse (GET /api/posts/{id})
```

> **Note:** The server's `apiPostThreadResponse` (a combined Post + Replies structure) corresponds to the `/api/posts/{id}/thread` route, which **is not registered**. The client must not depend on this structure; `kt post` assembles post and reply data via two independent requests.
>
> Similarly, `apiPostListResponse` (a post list with an embedded User field) comes from the unregistered `/api/users/{username}/posts` endpoint and does not need to be declared in `types.go`.

---

## Suggested Development Order

1. `internal/client` — API wrapper layer (build and test first)
2. `internal/display` — formatted output layer
3. `cmd/kt/main.go` — command entry point, wiring the two layers together
4. Phase 1 commands in order: `timeline` → `user` → `post` → `docs`
5. Phase 2 features iterated on demand

---

## Dependencies

- Go standard library (`net/http`, `encoding/json`, `flag`)
- `github.com/charmbracelet/bubbletea` — terminal UI framework
- `github.com/charmbracelet/bubbles` — TUI components (viewport, etc.)
- `github.com/charmbracelet/glamour` — Markdown rendering for the terminal
- `github.com/charmbracelet/lipgloss` — terminal styling and layout
- `golang.org/x/term` — TTY detection

---

## Notes

- All API endpoints are public and read-only; no authentication required
- The `--host` flag supports pointing at a local instance for development and debugging
- `internal/client` has no dependency on any UI layer, keeping it reusable
