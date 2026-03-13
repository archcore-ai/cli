# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Search Priority

When researching patterns, decisions, or conventions in this project, always search `.archcore/` documents FIRST (`list_documents` → `get_document`) before grepping the codebase or using external sources. See `.archcore/cli/archcore-first-search-priority.rule.md`.

## Build & Test Commands

```bash
# Build
go build -o archcore .

# Run all tests
go test ./...

# Run tests for a specific package
go test ./cmd/
go test ./internal/config/
go test ./internal/api/
go test ./templates/

# Run a single test
go test ./cmd/ -run TestConfigCmd

# Run tests with verbose output
go test -v ./...
```

## Architecture

This is a Go CLI tool (`archcore-cli`) for a context engineering platform. It manages a local `.archcore/` directory containing structured documents organized into three categories: `vision/`, `knowledge/`, and `experience/`.

### Package Structure

- **`cmd/`** — All cobra commands (`init`, `create`, `status`, `config`, `doctor`). Each command is in its own file with co-located tests.
- **`internal/config/`** — Settings management (`settings.json`), directory initialization, and field validation. Settings behavior varies by sync type (`none`, `cloud`, `on-prem`).
- **`internal/api/`** — HTTP client for the Archcore server (`/api/v1/status`, `/api/v1/projects`). Cloud URL: `https://app.archcore.ai`.
- **`internal/display/`** — Terminal output formatting using lipgloss (styled text, banners, status lines).
- **`templates/`** — 10 document template types (ADR, RFC, Rule, Guide, Doc, TaskType, CPAT, PRD, Idea, Plan), each mapped to a category directory.

### Key Design Patterns

- **Sync modes** drive validation: `cloud` requires `project_id`, `on-prem` requires `project_id` + `archcore_url`, `none` forbids both. This validation lives in `internal/config/config.go`.
- **Interactive CLI forms** use `charmbracelet/huh`. Commands support both interactive and flag-based usage.
- **Template types map to categories**: vision (ADR, PRD, Idea, Plan), knowledge (RFC, Rule, Guide, Doc), experience (TaskType, CPAT). This mapping is defined in `templates/templates.go`.
- **Tests use `t.TempDir()`** for isolated file system operations and `httptest.NewServer` for API client tests. Table-driven subtests with `t.Run()` throughout.
