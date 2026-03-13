# Archcore

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/archcore-ai/cli)](https://github.com/archcore-ai/cli/releases)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/archcore-ai/cli/releases)

**System Context Platform — keeps humans and AI in sync with your system.**

AI coding agents start every session with amnesia. Your architectural decisions, coding standards, past incidents, and project context are scattered across wikis, Slack threads, and tribal knowledge. Each new session means re-explaining the same things.

Archcore fixes this. It creates a `.archcore/` directory in your repository — a structured, version-controlled knowledge base that AI agents read automatically:

- **10 document types** — ADRs, RFCs, Rules, Guides, Plans, and more — each with purpose-built templates
- **Integrates with 8 AI coding agents** — Claude Code, Cursor, Gemini CLI, GitHub Copilot, Codex CLI, OpenCode, Roo Code, and Cline
- **MCP server** for real-time context injection into any LLM-powered tool
- **Local-first and Git-friendly** — lives in your repo, versioned with your code, shared with your team
- **Cloud sync** for cross-project knowledge discovery (coming soon)

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [How It Works](#how-it-works)
- [Commands](#commands)
- [AI Agent Integration](#ai-agent-integration)
- [Configuration](#configuration)
- [Development](#development)
- [Links & License](#links--license)

## Quick Start

```bash
# Install
curl -fsSL https://archcore.ai/install.sh | bash

# Initialize in your project
cd your-project
archcore init

# Check your setup
archcore doctor
```

## Installation

### Install Script (recommended)

```bash
curl -fsSL https://archcore.ai/install.sh | bash
```

### Go Install

```bash
go install github.com/archcore-ai/cli@latest
```

### From Source

```bash
git clone https://github.com/archcore-ai/cli.git
cd cli
go build -o archcore .
```

**Supported platforms:** macOS, Linux, Windows — amd64 and arm64.

## How It Works

1. **Initialize** — `archcore init` creates a `.archcore/` directory and auto-installs MCP server config for your AI coding agent.
2. **Build context** — add documents through your AI agent (via MCP tools) or by hand — both work equally well.
3. **Stay in sync** — every agent session starts with your full project context loaded automatically.

```
.archcore/
├── settings.json
├── .sync-state.json
├── auth/
│   ├── jwt-strategy.adr.md
│   └── auth-redesign.prd.md
├── payments/
│   └── stripe.adr.md
└── infrastructure/
    └── migration.plan.md
```

The directory structure is **free-form** — organize documents by domain, feature, team, or any structure that fits your project. Categories are virtual, derived from the document type in the filename (`slug.type.md`).

*Archcore CLI is best example for how it works: https://github.com/archcore-ai/cli/tree/main/.archcore*

### Document Types

Arcor has 3 fundamental layers of knowledge: Vision, Knowledge, Experience.

| Type | Full Name | Category | Description |
|------|-----------|----------|-------------|
| `prd` | Product Requirements Document | Vision | Goals, user stories, acceptance criteria, and success metrics |
| `idea` | Idea | Vision | Lightweight capture of a product or technical idea for future exploration |
| `plan` | Plan | Vision | Phased task list with acceptance criteria and dependencies |
| `adr` | Architecture Decision Record | Knowledge | Captures a finalized technical decision with context, alternatives, and consequences |
| `rfc` | Request for Comments | Knowledge | Proposes a significant change open for team review and feedback |
| `rule` | Rule | Knowledge | Coding or process standard — imperative statements with good/bad examples |
| `guide` | Guide | Knowledge | Step-by-step how-to instructions for completing a specific task |
| `doc` | Document | Knowledge | Reference documentation — lookup tables, registries, descriptive material |
| `task-type` | Task Type | Experience | Recurring workflow pattern — reusable checklist and workflow for a common task |
| `cpat` | Code Change Patterns | Experience | Root-cause analysis of a bug or incident with prevention steps |

Each document is a Markdown file with YAML frontmatter:

```markdown
---
title: "Use PostgreSQL for Primary Storage"
status: draft
---

## Context
...
```

Valid statuses: `draft`, `accepted`, `rejected` for all types of documents.

### Document Relations

Documents can be linked with directed relations to other documents:

- **related** — general association
- **implements** — source implements what target specifies
- **extends** — source builds upon target
- **depends_on** — source requires target to proceed

Relations are stored in `.sync-state.json` and managed automatically by the AI agent through MCP tools.

## Commands

| Command | Description |
|---------|-------------|
| `archcore init` | Initialize `.archcore/` directory interactively |
| `archcore doctor` | Run diagnostic checks on your setup |
| `archcore validate` | Validate document structure and frontmatter |
| `archcore config` | View or modify settings |
| `archcore hooks install` | Install hooks for detected AI agents |
| `archcore update` | Update archcore to the latest version |
| `archcore mcp` | Run the MCP stdio server |
| `archcore mcp install` | Install MCP config for detected agents |

### Update

```bash
# Update to the latest version
archcore update
```

The command checks GitHub Releases for a newer version, downloads it, verifies the SHA-256 checksum, and atomically replaces the current binary.

### Examples

```bash
# Install integrations for a specific agent
archcore hooks install --agent cursor
archcore mcp install --agent gemini-cli
```

## AI Agent Integration

Archcore integrates with AI coding agents in two ways:

- **Hooks** inject context at session start, so the agent is aware of your `.archcore/` documents from the first message.
- **MCP** (Model Context Protocol) gives the agent tools to list, read, create, update, and link documents in real time.

### Supported Agents

| Agent | Hooks | MCP |
|-------|-------|-----|
| Claude Code | yes | yes |
| Cursor | yes | yes |
| Gemini CLI | yes | yes |
| GitHub Copilot | yes | yes |
| OpenCode | — | yes |
| Codex CLI | — | yes |
| Roo Code | — | yes |
| Cline | — | manual |

### Install Integrations

```bash
# Auto-detect agents in your project and install everything
archcore hooks install

# Or target a specific agent
archcore mcp install --agent opencode
```

## Configuration

Settings are stored in `.archcore/settings.json` and created during `archcore init`.

| Field | Description | Values |
|-------|-------------|--------|
| `sync` | Sync mode. Cloud and on-prem coming soon. | `none` (local only), `cloud`, `on-prem` |
| `language` | Documents language. Helps the agent understand in which language to generate documentation | String, defaults to `en` |

```bash
archcore config                              # show all settings
archcore config get <key>                    # get a specific value
archcore config set <key> <value>            # set a value
```

## Development

### Prerequisites

- Go 1.24+

### Build & Test

```bash
# Build
go build -o archcore .

# Run all tests
go test ./...

# Run a specific package
go test ./cmd/

# Run a single test
go test ./cmd/ -run TestConfigCmd
```

### Project Structure

```
├── cmd/              # Cobra commands (init, doctor, config, validate, hooks, mcp, ...)
├── internal/
│   ├── agents/       # 8 supported AI agents with hooks/MCP capabilities
│   ├── api/          # HTTP client for archcore server
│   ├── config/       # Settings management and directory init
│   ├── display/      # Terminal output formatting (lipgloss)
│   ├── update/       # Self-update logic (version check, download, verify, replace)
│   ├── mcp/          # MCP stdio server implementation
│   └── sync/         # Sync logic
├── templates/        # 10 document type templates
├── install.sh        # Install script
└── .goreleaser.yaml  # Release configuration
```

## Links & License

- **Website:** [archcore.ai](https://archcore.ai)
- **Issues:** [github.com/archcore-ai/cli/issues](https://github.com/archcore-ai/cli/issues)
- **License:** [Apache 2.0](LICENSE)
