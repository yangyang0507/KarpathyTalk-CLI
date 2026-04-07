# Contributing to kt

Thank you for your interest in contributing! This document covers everything you need to get started.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Commit Convention](#commit-convention)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Development Setup

**Prerequisites:** Go 1.26.1 or later.

```bash
# 1. Fork the repo on GitHub, then clone your fork
git clone https://github.com/<your-username>/KarpathyTalk-CLI.git
cd KarpathyTalk-CLI

# 2. Add the upstream remote
git remote add upstream https://github.com/yangyang0507/KarpathyTalk-CLI.git

# 3. Build and verify
make build
./dist/kt timeline --help
```

Available `make` targets:

```bash
make build     # compile for the current platform → dist/kt
make install   # install to $GOPATH/bin
make release   # cross-compile for all platforms → dist/
make clean     # remove dist/
```

To point the CLI at a local KarpathyTalk instance during development:

```bash
./dist/kt --host http://localhost:8080 timeline
```

## Project Structure

```
cmd/kt/main.go        entry point and CLI flag parsing
internal/client/      HTTP client and API types
internal/display/     output formatters (human, JSON, Markdown)
internal/tui/         interactive terminal UI (Bubble Tea)
docs/                 project documentation
```

See [`docs/CLI_SPEC.md`](docs/CLI_SPEC.md) for the design philosophy before making structural changes.

## Making Changes

1. **Sync with upstream** before starting:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Create a branch** named after the change type and a short description:

   ```bash
   git checkout -b feat/dark-mode
   git checkout -b fix/header-alignment
   ```

3. **Make your changes.** A few guidelines:
   - Keep changes focused — one concern per PR.
   - Follow existing code style. A `.golangci.yml` config is enforced in CI — run `golangci-lint run` locally before pushing.
   - Do not add dependencies without prior discussion in an issue.
   - Update `CHANGELOG.md` under `[Unreleased]` for any `feat` or `fix` changes.

4. **Verify before pushing:**

   ```bash
   go build ./...
   go vet ./...
   go test ./...
   make build
   ```

## Commit Convention

This project follows [Conventional Commits](https://www.conventionalcommits.org/). Every commit message must have the form:

```
<type>: <short description>
```

The description must start with a lowercase letter.

| Type | When to use |
|------|-------------|
| `feat` | A new command, flag, or user-visible behaviour |
| `fix` | A bug fix |
| `docs` | Documentation only |
| `refactor` | Code change with no behaviour change |
| `chore` | Dependencies, tooling, config |
| `ci` | CI/CD workflows |
| `build` | Makefile, build scripts |

**Examples:**

```
feat: add kt watch command for polling a user
fix: prevent URL overflow in list card preview
docs: update installation instructions
chore: upgrade bubbletea to v1.4.0
```

PR titles are validated automatically against this convention. Commits within a branch are not checked, but keeping them consistent makes the changelog cleaner.

## Pull Request Process

1. Push your branch and open a PR against `main`.
2. Fill in the PR template — summary, type, and testing notes.
3. Ensure all CI checks pass: `build`, `vet`, `test`, `lint` (golangci-lint), and PR title validation.
4. A maintainer will review and may request changes.
5. Once approved, the PR is squash-merged. The PR title becomes the commit message, so make sure it follows the commit convention above.

## Reporting Issues

Use the issue templates on GitHub:

- **Bug Report** — for something that isn't working as expected.
- **Feature Request** — for new commands, flags, or improvements.

For general questions, open a [Discussion](https://github.com/yangyang0507/KarpathyTalk-CLI/discussions) instead of an issue.
