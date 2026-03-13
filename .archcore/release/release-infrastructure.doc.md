---
title: "Release Infrastructure Overview"
status: accepted
---

## Overview

The Archcore CLI uses a tag-driven release pipeline: pushing a `v*` tag triggers GitHub Actions, which runs tests and invokes GoReleaser to build cross-platform binaries and publish a GitHub Release.

## Content

### Components

| Component | File | Purpose |
|---|---|---|
| Version vars | `main.go` | `version` and `commit` vars with dev defaults |
| Cobra integration | `cmd/root.go` | `NewRootCmd(version, commit)` sets `Version` field and version template |
| GoReleaser config | `.goreleaser.yaml` | Defines build matrix, archive naming, checksums |
| GitHub Actions | `.github/workflows/release.yml` | Orchestrates test → build → publish on tag push |
| Install script | `install.sh` | End-user installer that downloads release artifacts |
| Self-update | `internal/update/update.go` | In-binary update: check latest version, download, verify checksum, atomic replace |
| Update command | `cmd/update.go` | `archcore update` — user-facing self-update command |

### Build Matrix

| OS | Architecture |
|---|---|
| darwin | amd64, arm64 |
| linux | amd64, arm64 |
| windows | amd64, arm64 |

All builds use `CGO_ENABLED=0` for static binaries and `-s -w` ldflags to strip debug info.

### Artifact Naming

Archives follow the pattern `archcore_<os>_<arch>.tar.gz` for darwin/linux and `archcore_<os>_<arch>.zip` for windows (e.g. `archcore_darwin_arm64.tar.gz`, `archcore_windows_amd64.zip`). This matches what `install.sh` and `archcore update` expect for unix platforms. Windows users download the zip directly from the GitHub release page.

A `checksums.txt` file with SHA-256 hashes is included in every release for verification.

### Version Format

- Dev builds: `archcore dev (commit: none)`
- Release builds: `archcore 1.2.3 (commit: abc1234)`

The version template is set via `SetVersionTemplate` on the cobra root command.

### Update Paths

Users can update the CLI via:

1. **`archcore update`** — self-update command that downloads and replaces the binary in-place
2. **Re-running install script** — `curl -fsSL https://archcore.ai/install.sh | bash`
3. **`go install`** — `go install github.com/archcore-ai/cli@latest`

### Secrets

Only `GITHUB_TOKEN` is required (automatically provided by GitHub Actions). No additional secrets, signing keys, or notarization credentials are needed.

## Examples

**Release artifact listing for v1.0.0:**

```
archcore_darwin_amd64.tar.gz
archcore_darwin_arm64.tar.gz
archcore_linux_amd64.tar.gz
archcore_linux_arm64.tar.gz
archcore_windows_amd64.zip
archcore_windows_arm64.zip
checksums.txt
```