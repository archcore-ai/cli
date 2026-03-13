---
title: "Building the Archcore CLI"
status: accepted
---

## Overview

Guide for building, testing, and extending the Archcore CLI — a Go tool that manages a local `.archcore/` directory of structured documents and integrates with AI coding agents (Claude Code, Cursor, Gemini CLI, and others) via MCP and hooks.

See also: [Supported AI Agents Registry](supported-ai-agents.rule.md), [CLI Hooks Reference](cli-hooks-reference.doc.md), [Agent Integration Guide](agent-hooks-integration.guide.md).

## Prerequisites

- Go 1.24+, Git
- Familiarity with Go modules, packages, and testing

## Build & Test

```bash
go build -o archcore .     # Build binary
go test ./...              # Run all tests
go test ./cmd/ -run TestX  # Run a specific test
```

## Getting Started

```bash
./archcore init       # Interactive setup wizard (sync type, directories, settings)
./archcore doctor     # Health check: structure + settings + server connectivity
./archcore validate   # Structural checks only (naming, frontmatter, categories)
./archcore update     # Self-update to the latest release
```

`init` creates `.archcore/` with three subdirectories (`vision/`, `knowledge/`, `experience/`), each holding documents of specific types. Settings go in `.archcore/settings.json`. It also auto-detects AI agents and installs hooks + MCP config for all found agents.

## Settings and Configuration

Settings live in `.archcore/settings.json`. Use the `config` command to read and modify them:

```bash
./archcore config get sync        # Read a setting
./archcore config set language ru  # Set a setting
```

### Settings Fields

| Field          | Type   | Description                                                    |
|----------------|--------|----------------------------------------------------------------|
| `sync`         | string | Sync mode: `none`, `cloud`, or `on-prem`                      |
| `project_id`   | int    | Project ID for cloud/on-prem sync (optional)                   |
| `archcore_url` | string | Server URL for on-prem sync (required for on-prem)             |
| `language`     | string | Language for MCP-generated document content (default: `"en"`)  |

The `language` field controls the language the MCP server uses when generating document content (section headers, placeholders, descriptions). It is sync-independent — available in all modes. When not set, defaults to `"en"`. Uses `omitempty` so it only appears in settings.json when explicitly configured.

See [Sync Mode Field Validation](../sync/sync-mode-field-validation.rule.md) for per-sync-mode field rules, and [Optional Settings Omit Defaults](optional-settings-omit-defaults.rule.md) for the convention on optional fields.

## How to Add a New Setting

1. Add the field to `Settings` struct in `internal/config/config.go` with appropriate `json:"...,omitempty"` tag
2. Add the field name to `allowedFields` for the relevant sync types (all types if sync-independent)
3. Add to `requiredFields` only if it's mandatory for a sync type
4. Update `Validate()` with any constraints
5. Update `MarshalJSON()` — add the field to the per-sync-type struct literals
6. Update `UnmarshalJSON()` — add decoding and type validation
7. Add `"fieldname"` cases to `getSettingsValue()` and `setSettingsValue()` in `cmd/config.go`
8. Add tests in both `internal/config/config_test.go` and `cmd/config_cmd_test.go`

See [Optional Settings Omit Defaults](optional-settings-omit-defaults.rule.md) for conventions on optional fields.

## How to Add a New Command

1. Create `cmd/<name>.go` with a `newXxxCmd() *cobra.Command` function
2. Create `cmd/<name>_test.go` with table-driven tests
3. Register in `cmd/root.go` via `root.AddCommand(newXxxCmd())`
4. If the command needs the CLI version (like `update`), pass it from `NewRootCmd`: `newXxxCmd(cleaned)`
5. Keep cobra wiring minimal — extract logic into testable functions that accept a base directory

## How to Add a New Document Type

1. Add a `TypeXxx` constant in `templates/templates.go`
2. Add it to `categoryMap` with the correct category
3. Add it to `ValidTypes()` return slice
4. Create a `generateXxxTemplate()` function
5. Add the case to `GenerateTemplate()` switch
6. Update `typeDescriptions` in `cmd/hooks_common.go` for hook context

Document types map to categories:

| Category | Types |
|----------|-------|
| `knowledge/` | adr, rfc, rule, guide, doc |
| `vision/` | prd, idea, plan |
| `experience/` | task-type, cpat |

Files follow the naming convention: `<slug>.<type>.md` (e.g., `use-postgres.adr.md`).

## How to Add a New MCP Tool

The MCP server (`archcore mcp`) exposes tools for AI agents to manage documents. Currently: `list_documents`, `get_document`, `create_document`, `update_document`.

1. Create `internal/mcp/tools/<name>.go` returning `(mcp.Tool, server.ToolHandlerFunc)`
2. Create `internal/mcp/tools/<name>_test.go`
3. Register in `internal/mcp/server.go` via `s.AddTool()`
4. Use helpers from `common.go` (`ScanDocuments`, `ReadDocumentContent`, `ExtractDocType`, `splitDocument`)
5. Validate inputs and check path safety (no `..`, must start with `.archcore/`)

## How to Modify Hooks

Hooks intercept agent lifecycle events to inject documentation context and detect documentation opportunities. The implementation is split across multiple files with shared handler logic.

### File Structure

| File | Purpose |
|------|---------|
| `cmd/hooks.go` | Install command, Claude Code hooks config writer, `installHooksForAgent()` router |
| `cmd/hooks_claude_code.go` | Claude Code subcommand, `hookInput`/`hookOutput` structs, shared handler factories |
| `cmd/hooks_cursor.go` | Cursor subcommand, hooks config writer (`cursorHooksConfig` struct) |
| `cmd/hooks_gemini_cli.go` | Gemini CLI subcommand, hooks config writer |
| `cmd/hooks_common.go` | `buildSessionContext()`, keywords, type descriptions, instruction templates |

### Shared Handler Pattern

All agents share the same handler logic through factory functions in `cmd/hooks_claude_code.go`:

- `newSessionStartHookCmd(use, short)` — calls `handleSessionStart()` which uses `buildSessionContext()`
- `newStopHookCmd(use, short)` — calls `handleStop()` which uses `checkStopKeywords()`
- `newPromptHookCmd(use, short, eventName)` — calls `handleUserPromptSubmit()` which uses `checkPromptKeywords()`

Each agent registers its subcommands using these factories with agent-specific event names.

### Event Coverage by Agent

| Event | Claude Code | Cursor | Gemini CLI |
|-------|-------------|--------|------------|
| Session Start | `SessionStart` | `sessionStart` | `SessionStart` |
| Stop | `Stop` | `stop` | — |
| Prompt Submit | `UserPromptSubmit` | `beforeSubmitPrompt` | `BeforeAgent` |

### Modifying Keywords

Stop keywords and prompt keywords are defined in `cmd/hooks_common.go`:

- **Stop keywords** (`stopKeywords` slice, line 153) — add new phrases with a document type and suggested filename
- **Prompt keywords** (`promptKeywords` slice, line 185) — add new phrases; set `IsRegex: true` for patterns with regex metacharacters
- **Instruction templates** (constants `adrInstruction`, `planInstruction`, `cpatInstruction`, line 211) — modify or add new instruction text

See [CLI Hooks Reference](cli-hooks-reference.doc.md) for the full keyword tables.

## How to Add a New Agent

To add support for a new AI coding agent:

1. **Define the agent ID** — Add a new `AgentID` constant in `internal/agents/agents.go`

2. **Create the agent file** — Add `internal/agents/<name>.go` returning an `*Agent` struct:
   - `ID`, `DisplayName` — identity
   - `DetectFn` — check for the agent's marker directory/file
   - `MCPConfigPath` — return the path to the MCP config file
   - `WriteMCPConfig` — write the archcore MCP entry (use `WriteStandardMCPJSON` if the agent uses standard `mcpServers` format)
   - `WriteHooksConfig` — set to `nil` if hooks not supported
   - `ManualMCPInstallHint` — set if MCP must be installed manually (e.g., Cline)

3. **Register the agent** — Add the constructor call to the `all` slice in `internal/agents/agents.go`

4. **Add tests** — Create `internal/agents/<name>_test.go` covering detection, MCP config writing, and idempotency

5. **If hooks are supported:**
   - Create `cmd/hooks_<name>.go` with a `newHooksXxxCmd()` subcommand using the shared factories
   - Create `cmd/hooks_<name>_test.go`
   - Add a `runXxxHooksInstall()` function for writing the agent's hooks config
   - Register the subcommand in `cmd/hooks.go:newHooksCmd()`
   - Add the agent case to `installHooksForAgent()` in `cmd/hooks.go`

6. **Update documentation:**
   - Add the agent to [Supported AI Agents Registry](supported-ai-agents.rule.md)
   - Add config examples to [Agent Integration Guide](agent-hooks-integration.guide.md)

## Key Design Patterns

- **Sync modes drive validation** — `none`/`cloud`/`on-prem` each require/forbid different settings fields. Custom JSON marshaling in `internal/config/` enforces this.
- **Constructor functions** — each command is `newXxxCmd() *cobra.Command` with logic extracted into testable functions.
- **Version-aware commands** — commands needing the CLI version (like `update`) receive it as a parameter from `NewRootCmd`.
- **Interactive forms** — `charmbracelet/huh` for interactive input, with flag-based fallbacks.
- **Co-located tests** — every command and package has adjacent `_test.go` files using `t.TempDir()` and table-driven subtests.
- **Shared hook handlers** — all agents use the same `handleSessionStart`, `handleStop`, `handleUserPromptSubmit` functions via command factories, differing only in event names and config format.
- **Invalid config backup** — corrupted config files are backed up as `.bak` before being overwritten. See [Backup Invalid Configs](backup-invalid-configs.adr.md).

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/huh` | Interactive terminal forms |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/mark3labs/mcp-go` | MCP stdio server |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
| `github.com/wk8/go-ordered-map/v2` | Deterministic JSON key ordering (hooks config) |