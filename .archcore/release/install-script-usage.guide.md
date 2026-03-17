---
title: "Installing Archcore CLI via install.sh"
status: accepted
---

## Prerequisites

### macOS / Linux

- `curl` — for downloading release artifacts
- `tar` — for extracting the archive
- `sha256sum` or `shasum` — for checksum verification (optional but recommended)
- A GitHub release must exist with `archcore_<os>_<arch>.tar.gz` and `checksums.txt` assets

### Windows

- A browser or PowerShell to download `archcore.exe` from [GitHub Releases](https://github.com/archcore-ai/cli/releases/latest)

### Windows (WSL)

- [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) installed
- Same prerequisites as macOS / Linux (curl, tar, sha256sum)

## Installation Methods

### macOS / Linux (recommended)

1. **Basic install (latest version)**

   ```bash
   curl -fsSL https://archcore.ai/install.sh | sh
   ```

2. **Pin a specific version**

   ```bash
   ARCHCORE_VERSION=v1.0.0 curl -fsSL https://archcore.ai/install.sh | sh
   ```

3. **Custom install directory** (default: `~/.local/bin`)

   ```bash
   ARCHCORE_INSTALL_DIR=/usr/local/bin curl -fsSL https://archcore.ai/install.sh | sh
   ```

4. **Authenticate for private repos or rate limits**

   ```bash
   GITHUB_TOKEN=ghp_xxx curl -fsSL https://archcore.ai/install.sh | sh
   ```

### Windows

Download `archcore.exe` from the [latest release](https://github.com/archcore-ai/cli/releases/latest) and add it to your `PATH`.

```powershell
# Example: move to a directory in your PATH
Move-Item archcore.exe C:\Users\$env:USERNAME\.local\bin\
```

### Windows (WSL)

Install [WSL](https://learn.microsoft.com/en-us/windows/wsl/install), then run inside the WSL terminal:

```bash
curl -fsSL https://archcore.ai/install.sh | sh
```

This uses the same install script as macOS / Linux — WSL provides a full Linux environment.

## What the Script Does

1. Detects OS (`darwin`/`linux`) and architecture (`amd64`/`arm64`)
2. Resolves the latest version from GitHub API (or uses `ARCHCORE_VERSION`)
3. Downloads `archcore_<os>_<arch>.tar.gz` and `checksums.txt`
4. Verifies SHA-256 checksum
5. Extracts the binary and installs it atomically to the install directory
6. Checks if the install directory is in `$PATH` and prints shell-specific guidance if not

> **Note:** The install script supports macOS and Linux only. Windows users should download the `.exe` directly from GitHub Releases or use WSL.

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
- **"Unsupported operating system/architecture"** — Only `darwin`/`linux` on `amd64`/`arm64` are supported by the install script. Windows users should download `archcore.exe` directly from [GitHub Releases](https://github.com/archcore-ai/cli/releases/latest).