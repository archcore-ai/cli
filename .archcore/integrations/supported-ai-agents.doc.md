---
title: Supported AI Agents Registry
status: accepted
---

## Summary

Archcore integrates with 8 AI coding agents. Each agent has a unique combination of hooks support (lifecycle event interception) and MCP support (document tool access). This document is the authoritative registry.

## Agent Registry

| Agent | ID | Hooks | MCP | Detection Marker | Link |
|-------|----|-------|-----|------------------|------|
| Claude Code | `claude-code` | Yes | Yes | `.claude/` dir | [docs.anthropic.com](https://docs.anthropic.com/en/docs/claude-code) |
| Cursor | `cursor` | Yes | Yes | `.cursor/` dir | [cursor.com](https://www.cursor.com/) |
| Gemini CLI | `gemini-cli` | Yes | Yes | `.gemini/` dir | [github.com/google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli) |
| OpenCode | `opencode` | No | Yes | `opencode.json` file or `.opencode/` dir | [opencode.ai](https://opencode.ai/) |
| Codex CLI | `codex-cli` | No | Yes | `.codex/` dir | [github.com/openai/codex](https://github.com/openai/codex) |
| Roo Code | `roo-code` | No | Yes | `.roo/` dir | [roocode.com](https://roocode.com/) |
| Cline | `cline` | No | Manual | `.clinerules/` dir | [cline.bot](https://cline.bot/) |
| GitHub Copilot | `copilot` | Yes | Yes | `.github/copilot-instructions.md` file | [github.com/features/copilot](https://github.com/features/copilot) |

## Integration Levels

### Full Integration (Hooks + MCP)

Agents: **Claude Code**, **Cursor**, **Gemini CLI**, **GitHub Copilot**

These agents support both lifecycle hooks (session start, stop, prompt submit) and MCP tool access. Archcore can automatically detect them, install hooks, and configure MCP.

### MCP Only

Agents: **OpenCode**, **Codex CLI**, **Roo Code**

These agents support MCP for document tool access but do not have a hooks mechanism compatible with archcore. They are auto-detected and receive MCP config during `archcore init` or `archcore mcp install`.

### Manual

Agent: **Cline**

Cline stores MCP config in VS Code `globalStorage`, not in a project-level file. Users must add the archcore MCP server manually via Cline's MCP settings UI. Archcore prints a hint when Cline is detected.

## Per-Agent Details

### Claude Code

- **Config paths:** `.claude/settings.json` (hooks), `.mcp.json` (MCP)
- **Hook events:** `SessionStart`, `Stop`, `UserPromptSubmit`
- **Hook commands:** `archcore hooks claude-code session-start|stop|user-prompt-submit`
- **MCP format:** Standard `mcpServers` JSON (`{"command": "archcore", "args": ["mcp"]}`)
- **Source:** `internal/agents/claude_code.go`, `cmd/hooks_claude_code.go`

### Cursor

- **Config paths:** `.cursor/hooks.json` (hooks), `.cursor/mcp.json` (MCP)
- **Hook events:** `sessionStart`, `stop`, `beforeSubmitPrompt`
- **Hook commands:** `archcore hooks cursor session-start|stop|before-submit-prompt`
- **MCP format:** Standard `mcpServers` JSON
- **Source:** `internal/agents/cursor.go`, `cmd/hooks_cursor.go`

### Gemini CLI

- **Config paths:** `.gemini/settings.json` (hooks and MCP, shared file)
- **Hook events:** `SessionStart`, `BeforeAgent`
- **Hook commands:** `archcore hooks gemini-cli session-start|before-agent`
- **MCP format:** Standard `mcpServers` JSON (inside same `settings.json`)
- **Note:** No `Stop` event — Gemini CLI uses `BeforeAgent` instead of `UserPromptSubmit`
- **Source:** `internal/agents/gemini_cli.go`, `cmd/hooks_gemini_cli.go`

### GitHub Copilot

- **Config paths:** `.github/hooks/archcore.json` (hooks), `.vscode/mcp.json` (MCP)
- **Hook events:** `sessionStart`
- **Hook commands:** `archcore hooks copilot session-start`
- **Hook format:** Uses `bash` field instead of `command` (`{"type": "command", "bash": "..."}`)
- **MCP format:** VS Code-style `servers` JSON with `"type": "stdio"` (`{"servers": {"archcore": {"type": "stdio", "command": "archcore", "args": ["mcp"]}}}`)
- **Detection:** `.github/copilot-instructions.md` file
- **Source:** `internal/agents/copilot.go`, `cmd/hooks_copilot.go`

### OpenCode

- **Config path:** `opencode.json` (MCP)
- **MCP format:** `{"mcp": {"archcore": {"type": "local", "command": ["archcore", "mcp"]}}}`
- **Note:** OpenCode uses a different MCP JSON structure with `type` and `command` as array
- **Source:** `internal/agents/opencode.go`

### Codex CLI

- **Config path:** `.codex/config.toml` (MCP)
- **MCP format:** TOML block `[mcp_servers.archcore]` with `command` and `args`
- **Note:** Only agent using TOML config format
- **Source:** `internal/agents/codex_cli.go`

### Roo Code

- **Config path:** `.roo/mcp.json` (MCP)
- **MCP format:** Standard `mcpServers` JSON
- **Note:** Roo Code only supports `onSave` hooks, not useful for lifecycle events
- **Source:** `internal/agents/roo_code.go`

### Cline

- **Config path:** VS Code `globalStorage` (not project-level)
- **MCP format:** Manual installation via Cline MCP settings UI
- **Hint shown:** "MCP config is stored in VS Code globalStorage — add manually via Cline MCP settings"
- **Source:** `internal/agents/cline.go`

## Adding a New Agent

1. **Define the ID** — Add a new `AgentID` constant in `internal/agents/agents.go`
2. **Create agent file** — Add `internal/agents/<name>.go` implementing the `Agent` struct with `DetectFn`, `MCPConfigPath`, `WriteMCPConfig`, and optionally `WriteHooksConfig`
3. **Register** — Add the agent constructor to the `all` slice in `internal/agents/agents.go`
4. **Add tests** — Create `internal/agents/<name>_test.go`
5. **If hooks supported** — Create `cmd/hooks_<name>.go` with event handlers and install logic; register the subcommand in `cmd/hooks.go:newHooksCmd()`; add the case to `installHooksForAgent()` in `cmd/hooks.go`
6. **Update this document** — Add the agent to the registry table above
7. **Cross-reference** — Update `agent-hooks-integration` guide and `building-the-cli` guide