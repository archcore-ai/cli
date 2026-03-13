---
title: "How to Create a New Release"
status: accepted
---

## Prerequisites

- Push access to the `archcore-ai/cli` repository
- All changes merged to `main`
- Tests passing on `main`

## Steps

1. **Ensure you're on the latest main**

   ```bash
   git checkout main
   git pull origin main
   ```

2. **Choose a version following semver**

   - `v1.0.0` — first stable release or breaking changes
   - `v1.1.0` — new features, backwards compatible
   - `v1.1.1` — bug fixes only

3. **Create and push the tag**

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **Monitor the release workflow**

   The GitHub Actions workflow (`.github/workflows/release.yml`) triggers automatically. It will:
   - Run `go test ./...`
   - Build binaries for darwin/linux × amd64/arm64
   - Create a GitHub Release with archives and `checksums.txt`

   Watch progress at: `https://github.com/archcore-ai/cli/actions`

5. **Verify the release**

   ```bash
   # Check the GitHub Release page has all 5 assets (4 archives + checksums.txt)
   gh release view v1.0.0

   # Test the install script
   ARCHCORE_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/archcore-ai/cli/main/install.sh | bash

   # Verify the installed binary
   archcore --version
   # Expected: archcore 1.0.0 (commit: <sha>)
   ```

## Verification

- GitHub Release page shows 4 `.tar.gz` archives + `checksums.txt`
- `archcore --version` on installed binary shows correct version and commit
- Install script succeeds on a clean machine

## Common Issues

- **Workflow fails at test step** — Fix the tests on `main`, delete the tag (`git push origin :v1.0.0 && git tag -d v1.0.0`), then re-tag after fixing.
- **GoReleaser fails** — Check `.goreleaser.yaml` syntax. Run `goreleaser check` locally if available.
- **Wrong commit tagged** — Delete the remote tag, re-tag the correct commit, and push again.
- **install.sh can't find the release** — Ensure the tag follows the `v*` pattern (e.g. `v1.0.0`, not `1.0.0`).