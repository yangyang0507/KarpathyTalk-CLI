# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/yangyang0507/KarpathyTalk-CLI/releases/tag/v0.1.0
