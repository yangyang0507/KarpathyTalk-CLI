# Release Process

This project uses [GoReleaser](https://goreleaser.com) and GitHub Actions to automate cross-platform builds and GitHub Releases.

## Overview

```
git tag v1.2.3 && git push origin v1.2.3
        │
        └─▶ GitHub Actions (.github/workflows/release.yml)
                │
                └─▶ GoReleaser (.goreleaser.yaml)
                        ├─ Cross-compile for 5 platforms
                        ├─ Package as .tar.gz / .zip
                        ├─ Generate checksums.txt (sha256)
                        └─ Publish GitHub Release with all assets
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html):

```
v<MAJOR>.<MINOR>.<PATCH>
```

| Increment | When |
|-----------|------|
| `MAJOR` | Breaking changes to CLI flags, commands, or output format |
| `MINOR` | New commands or flags; backwards-compatible additions |
| `PATCH` | Bug fixes and documentation updates |

Pre-release tags (e.g. `v1.0.0-beta.1`, `v2.0.0-rc.2`) are automatically marked as pre-releases on GitHub.

## Before Releasing

1. **Update `CHANGELOG.md`** — move items from `[Unreleased]` to a new versioned section:

   ```markdown
   ## [1.2.3] - 2026-04-07

   ### Added
   - ...

   ### Fixed
   - ...
   ```

   Update the comparison links at the bottom of the file:

   ```markdown
   [Unreleased]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v1.2.3...HEAD
   [1.2.3]: https://github.com/yangyang0507/KarpathyTalk-CLI/compare/v1.2.2...v1.2.3
   ```

2. **Verify the build locally:**

   ```bash
   make build        # smoke-test the current platform
   make release      # optional: verify all cross-compilation targets
   ```

3. **Commit the changelog update:**

   ```bash
   git add CHANGELOG.md
   git commit -m "chore: prepare release v1.2.3"
   git push origin main
   ```

## Tagging and Publishing

Push a version tag to trigger the release pipeline:

```bash
git tag v1.2.3
git push origin v1.2.3
```

The GitHub Actions release workflow starts automatically. Monitor progress at:

```
https://github.com/yangyang0507/KarpathyTalk-CLI/actions
```

## What GoReleaser Produces

| Asset | Platforms |
|-------|-----------|
| `kt-darwin-amd64.tar.gz` | macOS Intel |
| `kt-darwin-arm64.tar.gz` | macOS Apple Silicon |
| `kt-linux-amd64.tar.gz` | Linux x86-64 |
| `kt-linux-arm64.tar.gz` | Linux ARM64 |
| `kt-windows-amd64.zip` | Windows x86-64 |
| `checksums.txt` | sha256 hashes for all archives |

All binaries are statically linked (`CGO_ENABLED=0`) with no runtime dependencies.

Each archive includes the binary, `LICENSE`, and `README.md`.

## Verifying a Release

Download and verify the checksum before installing manually:

```bash
# Download binary and checksum file
curl -LO https://github.com/yangyang0507/KarpathyTalk-CLI/releases/download/v1.2.3/kt-darwin-arm64.tar.gz
curl -LO https://github.com/yangyang0507/KarpathyTalk-CLI/releases/download/v1.2.3/checksums.txt

# Verify
sha256sum --check --ignore-missing checksums.txt

# Extract and install
tar -xzf kt-darwin-arm64.tar.gz
mv kt /usr/local/bin/kt
```

## CI on Every Push

A separate workflow (`.github/workflows/ci.yml`) runs on every push to `main` and on all pull requests:

- `go build ./...` — ensures the project compiles
- `go vet ./...` — catches common correctness issues

Releases should only be tagged from commits where CI is green.

## Rolling Back a Release

GitHub does not allow re-pushing a tag. To fix a bad release:

1. Delete the tag locally and remotely:

   ```bash
   git tag -d v1.2.3
   git push origin :refs/tags/v1.2.3
   ```

2. Delete the GitHub Release from the web UI (mark it as a draft or delete it entirely).

3. Fix the issue, commit, then re-tag:

   ```bash
   git tag v1.2.4
   git push origin v1.2.4
   ```

   Prefer cutting a new patch release over reusing an existing tag.
