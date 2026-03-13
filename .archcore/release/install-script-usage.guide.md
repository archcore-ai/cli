---
title: "Installing Archcore CLI via install.sh"
status: accepted
---

## Prerequisites

- `curl` — for downloading release artifacts
- `tar` — for extracting the archive
- `sha256sum` or `shasum` — for checksum verification (optional but recommended)
- A GitHub release must exist with `archcore_<os>_<arch>.tar.gz` and `checksums.txt` assets

## Steps

1. **Basic install (latest version)**

   ```bash
   curl -fsSL https://raw.githubusercontent.com/archcore-ai/cli/main/install.sh | bash
   ```

2. **Pin a specific version**

   ```bash
   ARCHCORE_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/archcore-ai/cli/main/install.sh | bash
   ```

3. **Custom install directory** (default: `~/.local/bin`)

   ```bash
   ARCHCORE_INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/archcore-ai/cli/main/install.sh | bash
   ```

4. **Authenticate for private repos or rate limits**

   ```bash
   GITHUB_TOKEN=ghp_xxx curl -fsSL https://raw.githubusercontent.com/archcore-ai/cli/main/install.sh | bash
   ```

## What the Script Does

1. Detects OS (`darwin`/`linux`) and architecture (`amd64`/`arm64`)
2. Resolves the latest version from GitHub API (or uses `ARCHCORE_VERSION`)
3. Downloads `archcore_<os>_<arch>.tar.gz` and `checksums.txt`
4. Verifies SHA-256 checksum
5. Extracts the binary and installs it atomically to the install directory
6. Checks if the install directory is in `$PATH` and prints shell-specific guidance if not

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `ARCHCORE_VERSION` | (latest) | Pin to a specific release tag (e.g. `v1.0.0`) |
| `ARCHCORE_INSTALL_DIR` | `~/.local/bin` | Override the install directory |
| `GITHUB_TOKEN` | (none) | GitHub token for authenticated API/download requests |

## Verification

```bash
archcore --version
```

Expected output: `archcore <version> (commit: <sha>)`

## Common Issues

- **"command not found" after install** — The install directory is not in your `$PATH`. The script prints instructions for your shell (bash/zsh/fish).
- **"Failed to fetch latest version"** — Network issue or GitHub API rate limit. Set `GITHUB_TOKEN` or use `ARCHCORE_VERSION` to skip the API call.
- **"Checksum verification failed"** — The download was corrupted. Retry the install.
- **"Unsupported operating system/architecture"** — Only `darwin`/`linux` on `amd64`/`arm64` are supported.