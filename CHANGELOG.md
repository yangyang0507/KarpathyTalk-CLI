# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-04-07

### Added
- Progressive help system: `kt help`, `kt help <command>`, and `kt <command> --help` all show command-specific flags and examples
- `--version` flag prints the installed version
- `kt docs` renders Markdown through Glamour when running in a terminal (TTY)
- Reply pagination hint in non-TTY mode when `HasMore` is true
- Unit tests for `internal/client`, `internal/display`, `internal/tui`, and `cmd/kt`
- Regression tests for TUI flag passthrough and `post --json` pagination fields
- Makefile with `build`, `install`, `release`, `clean` targets
- `install.sh` one-line installer for macOS / Linux
- GoReleaser config (`.goreleaser.yaml`) for cross-platform release builds
- GitHub Actions: CI workflow (build / vet / test on Ubuntu + macOS), release workflow, PR check workflow
- PR template, bug report and feature request issue templates
- `docs/RELEASE.md` release process documentation
- `CONTRIBUTING.md` contributor guide
- MIT License
- Screenshots in README (timeline, user, post views)

### Changed
- Module path renamed from `kt` to `github.com/yangyang0507/KarpathyTalk-CLI`
- `splitArgs` rewritten to accept a `*flag.FlagSet` and correctly keep flag-value pairs together — fixes `--limit 50`, `--before 230`, etc. being silently ignored
- TUI now respects all CLI flags: `--before`, `--limit`, `--replies`, and `--limit` (for replies) are passed through to `tui.Config` instead of using defaults
- TTY branch moved before the HTTP fetch in all `run*` functions — eliminates the wasted API call that was discarded anyway
- `kt post --json` output now includes `has_more` and `next_cursor` for replies
- HTTP client sends a `User-Agent: kt/<version>` header with every request
- `go mod tidy` run — all direct dependencies now correctly marked as direct
- `--revision` flag in `kt post` now exits with an error if used without `--markdown` or `--raw`
- `docs/CLI_SPEC.md` updated to reflect current directory structure and dependencies
- CI: added macOS runner, `go mod tidy` diff check, and Windows cross-compile verification in release workflow

### Fixed
- `splitArgs` broke every value-taking flag (`--limit`, `--before`, `--revision`) — values were classified as positional arguments and never parsed
- `install.sh` downloaded a bare binary URL that never existed; now correctly downloads and extracts the `.tar.gz` archive produced by GoReleaser
- `cardHeight` was `4` but each card occupies 5 terminal lines, causing the visible-card count to be off by ~25%
- User profile box in TUI user mode was not re-rendered on terminal resize
- `kt docs` showed raw Markdown in TTY instead of rendered output
- `.gitignore` rule `kt` matched `cmd/kt/` — renamed to `/kt` to scope it to the repo root

### Dependencies
- Module path: `kt` → `github.com/yangyang0507/KarpathyTalk-CLI`

## [0.2.0] - 2026-04-07

### Added
- Interactive TUI for `kt timeline`, `kt user`, and `kt post` when running in a terminal (TTY)
- Scrollable post list with card layout; navigate with `j`/`k` or arrow keys, open with `Enter`
- Scrollable detail view (post + replies) using `bubbles/viewport`
- Auto-load more posts when approaching the end of the list
- Right-aligned engagement stats in post headers (♥ likes, ↺ reposts, ✦ replies)
- Two-level separator hierarchy: `─` between posts, `╌` between header and body
- `StripMarkdownImages` helper — replaces `![alt](url)` with `[image: url]` to prevent URL overflow
- `RenderHeader`, `RenderSummaryCard`, `RenderFull`, `RenderUserProfile` as public string-returning functions in the display package
- README in English (`README.md`) and Chinese (`README.zh.md`)
- CLI specification document moved to `docs/CLI_SPEC.md` and translated to English

### Changed
- Non-TTY output (piped) retains the existing print-based display unchanged
- `renderMarkdown` refactored into `RenderMarkdownWidth(text, width)` accepting an explicit column width
- Post preview in list view uses plain-text stripping instead of full Markdown rendering for predictable card height

### Fixed
- `💬` wide emoji replaced with `✦` (single-width) to prevent header misalignment
- Glamour renderer's leading blank line trimmed to eliminate spurious double-separator appearance
- `.gitignore` rule `kt` scoped to `/kt` (repo root only) so `cmd/kt/` is no longer excluded from version control

### Dependencies
- Added `github.com/charmbracelet/bubbletea v1.3.10`
- Added `github.com/charmbracelet/bubbles v1.0.0`

## [0.1.0] - 2026-04-07

### Added
- `kt timeline` — browse the public timeline with `--limit`, `--before`, `--json`, `--markdown` flags
- `kt user <username>` — view user profile and posts with `--replies`, `--limit`, `--before`, `--json`, `--markdown` flags
- `kt post <id>` — view a single post and its direct replies with `--limit`, `--json`, `--markdown`, `--raw`, `--revision` flags
- `kt docs` — fetch platform API documentation as Markdown
- Global `--host` flag for pointing at a local instance
- Human-friendly terminal output with Lipgloss styling (colored IDs, authors, timestamps, stats)
- Markdown rendering via Glamour with terminal-width word wrapping
- JSON passthrough output for scripting and LLM pipelines
- TTY-aware output: structured display in terminal, plain text when piped
- Flexible argument order: flags and positional args can appear in any order
- `internal/client` package — HTTP client wrapping all KarpathyTalk API endpoints
- `internal/display` package — three output channels (human, JSON, Markdown)
- CLI specification document (`CLI_SPEC.md`) covering design philosophy and API mapping

[Unreleased]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/yangyang0507/KarpathyTalk-CLI/releases/tag/v0.1.0
