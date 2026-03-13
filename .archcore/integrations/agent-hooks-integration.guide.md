---
title: Integrating Archcore with AI Coding Agents
status: accepted
---

## Overview

Archcore integrates with AI coding agents via two mechanisms:

- **Hooks** — Lifecycle event interception (session start, stop, prompt submit) to inject context and detect documentation opportunities. Supported by Claude Code, Cursor, and Gemini CLI.
- **MCP** — Model Context Protocol server providing document management tools (`list_documents`, `get_document`, `create_document`, `update_document`). Supported by all agents except Cline (manual setup).

See [Supported AI Agents Registry](supported-ai-agents.rule.md) for the full agent list and capabilities.

## Quick Start

```bash
archcore init
```

`archcore init` automatically:

1. Creates the `.archcore/` directory structure
2. Detects installed agents by checking for marker directories (`.claude/`, `.cursor/`, `.gemini/`, etc.)
3. Falls back to Claude Code if no agents are detected
4. Installs hooks for agents that support them (Claude Code, Cursor, Gemini CLI)
5. Installs MCP config for all detected agents

Source: `cmd/init.go:126-138`

No further action is needed for most setups.

## Manual Installation

### Hooks Only

```bash
archcore hooks install              # auto-detect and install for all found agents
archcore hooks install --agent cursor  # install for a specific agent
```

Note: `archcore hooks install` also triggers MCP installation automatically.

### MCP Only

```bash
archcore mcp install                # auto-detect and install for all found agents
archcore mcp install --agent codex-cli  # install for a specific agent
```

## Auto-Detection

Archcore detects agents by checking for marker directories or files in the project root:

| Agent       | Marker                                         |
| ----------- | ---------------------------------------------- |
| Claude Code | `.claude/` directory                           |
| Cursor      | `.cursor/` directory                           |
| Gemini CLI  | `.gemini/` directory                           |
| OpenCode    | `opencode.json` file or `.opencode/` directory |
| Codex CLI   | `.codex/` directory                            |
| Roo Code    | `.roo/` directory                              |
| Cline       | `.clinerules/` directory                       |

If no markers are found, archcore defaults to installing for Claude Code.

Source: `internal/agents/agents.go:67-76` (`Detect` function), individual agent `DetectFn` in `internal/agents/*.go`

## What Gets Installed

### Hooks (3 agents)

| Agent       | Config File             | Events                                       |
| ----------- | ----------------------- | -------------------------------------------- |
| Claude Code | `.claude/settings.json` | `SessionStart`, `Stop`, `UserPromptSubmit`   |
| Cursor      | `.cursor/hooks.json`    | `sessionStart`, `stop`, `beforeSubmitPrompt` |
| Gemini CLI  | `.gemini/settings.json` | `SessionStart`, `BeforeAgent`                |

### MCP (6 agents, Cline is manual)

| Agent       | Config File             | Format                                             |
| ----------- | ----------------------- | -------------------------------------------------- |
| Claude Code | `.mcp.json`             | Standard `mcpServers` JSON                         |
| Cursor      | `.cursor/mcp.json`      | Standard `mcpServers` JSON                         |
| Gemini CLI  | `.gemini/settings.json` | Standard `mcpServers` JSON (shared with hooks)     |
| OpenCode    | `opencode.json`         | Custom `mcp` section with `type` + `command` array |
| Codex CLI   | `.codex/config.toml`    | TOML `[mcp_servers.archcore]` block                |
| Roo Code    | `.roo/mcp.json`         | Standard `mcpServers` JSON                         |

## Per-Agent Config Examples

### Claude Code — `.claude/settings.json`

```json
{
  "hooks": {
    "SessionStart": [
      { "matcher": "", "hooks": [{ "type": "command", "command": "archcore hooks claude-code session-start" }] }
    ],
    "Stop": [{ "matcher": "", "hooks": [{ "type": "command", "command": "archcore hooks claude-code stop" }] }],
    "UserPromptSubmit": [
      { "matcher": "", "hooks": [{ "type": "command", "command": "archcore hooks claude-code user-prompt-submit" }] }
    ]
  }
}
```

### Claude Code — `.mcp.json`

```json
{
  "mcpServers": {
    "archcore": { "command": "archcore", "args": ["mcp"] }
  }
}
```

### Cursor — `.cursor/hooks.json`

```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{ "command": "archcore hooks cursor session-start", "type": "command" }],
    "stop": [{ "command": "archcore hooks cursor stop", "type": "command" }],
    "beforeSubmitPrompt": [{ "command": "archcore hooks cursor before-submit-prompt", "type": "command" }]
  }
}
```

### Cursor — `.cursor/mcp.json`

```json
{
  "mcpServers": {
    "archcore": { "command": "archcore", "args": ["mcp"] }
  }
}
```

### Gemini CLI — `.gemini/settings.json`

```json
{
  "hooks": {
    "SessionStart": [
      { "matcher": "", "hooks": [{ "type": "command", "command": "archcore hooks gemini-cli session-start" }] }
    ],
    "BeforeAgent": [
      { "matcher": "", "hooks": [{ "type": "command", "command": "archcore hooks gemini-cli before-agent" }] }
    ]
  },
  "mcpServers": {
    "archcore": { "command": "archcore", "args": ["mcp"] }
  }
}
```

### OpenCode — `opencode.json`

```json
{
  "mcp": {
    "archcore": { "type": "local", "command": ["archcore", "mcp"] }
  }
}
```

### Codex CLI — `.codex/config.toml`

```toml
[mcp_servers.archcore]
command = "archcore"
args = ["mcp"]
```

### Roo Code — `.roo/mcp.json`

```json
{
  "mcpServers": {
    "archcore": { "command": "archcore", "args": ["mcp"] }
  }
}
```

### Cline — Manual Setup

Cline stores MCP config in VS Code `globalStorage`, not in project files. To add archcore:

1. Open Cline MCP settings in VS Code
2. Add an MCP server with command `archcore` and args `["mcp"]`

## Invalid Config Recovery

When archcore reads a config file that contains invalid JSON, it creates a `.bak` backup before proceeding with a fresh config. This prevents data loss while keeping the installation non-blocking.

Example: if `.cursor/hooks.json` is corrupted, archcore writes `.cursor/hooks.json.bak` and starts with an empty hooks config.

See [Backup Invalid Configs](backup-invalid-configs.adr.md) for the full decision record.

## Troubleshooting

### "`.archcore/` not found" error

Run `archcore init` first. All hooks and MCP commands require an initialized project.

### Agent not detected

Check that the agent's marker directory exists in your project root. You can also target a specific agent with `--agent`:

```bash
archcore hooks install --agent gemini-cli
archcore mcp install --agent opencode
```

### Hooks not firing

1. Verify the config file exists and contains archcore entries (see examples above)
2. Ensure `archcore` is on your `PATH`
3. Check the agent's logs for hook execution errors

### MCP tools not available

1. Verify MCP config file is correct for your agent
2. Restart the agent/IDE after installing
3. For Cline, ensure you added the server via the MCP settings UI

### Corrupted config after install

Check for a `.bak` file next to the config. Restore it and retry:

```bash
cp .cursor/hooks.json.bak .cursor/hooks.json
archcore hooks install --agent cursor
```
