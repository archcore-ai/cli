# Archcore

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/archcore-ai/cli)](https://github.com/archcore-ai/cli/releases)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/archcore-ai/cli/releases)

**Git-native architectural memory for AI coding agents**

Archcore keeps decisions, rules, plans, incidents, and other project context versioned in your repo so Claude Code, Cursor, Copilot, Gemini CLI, and other agents share the same context across sessions.

> Archcore brings context engineering into the codebase.

AI coding agents start every session with partial memory. Your architecture decisions, coding conventions, postmortems, and implementation plans are scattered across docs, chats, and tool-specific memory files. Archcore gives your repo a shared, structured context layer that agents can read and update.

## Why Archcore

- **Shared repo context** — keep architecture knowledge where code lives
- **Works across agents** — Claude Code, Cursor, Gemini CLI, GitHub Copilot, Codex CLI, OpenCode, Roo Code, and Cline
- **Structured documents** — ADRs, RFCs, Rules, Guides, Plans, PRDs, incidents, and more
- **Git-native** — local-first, version-controlled, reviewable with code
- **MCP-powered** — agents can read, create, update, and link documents in real time
- **Built for teams** — context survives sessions, teammates, and tool changes

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [How It Works](#how-it-works)
- [What Lives in `.archcore/`](#what-lives-in-archcore)
- [Why Not Just Instruction Files?](#why-not-just-instruction-files)
- [Try It](#try-it)
- [Document Types](#document-types)
- [Document Relations](#document-relations)
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

# Validate setup
archcore doctor
```

After `archcore init`, Archcore creates a `.archcore/` directory in your repo and installs integrations so your coding assistant can read and manage project context.

## Installation

### macOS / Linux

```bash
curl -fsSL https://archcore.ai/install.sh | bash
```

### Windows

Download `archcore.exe` from the [latest release](https://github.com/archcore-ai/cli/releases/latest) and add it to your `PATH`.

```powershell
# Example: move to a directory in your PATH
Move-Item archcore.exe C:\Users\$env:USERNAME\.local\bin\
```

### Windows (WSL)

Install [WSL](https://learn.microsoft.com/en-us/windows/wsl/install), then run inside it:

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

1. **Initialize your repo**  
   `archcore init` creates `.archcore/` and installs integrations for supported agents.

2. **Capture durable context**  
   Store architecture decisions, rules, plans, product docs, and incident learnings as structured Markdown files.

3. **Let agents reuse it**  
   Hooks and MCP let your coding agents read existing context and create or update documents during real work.

4. **Keep it in Git**  
   Review context changes like code, evolve them over time, and keep them portable across tools.

## What Lives in `.archcore/`

```text
.archcore/
├── settings.json
├── .sync-state.json
├── auth/
│   ├── jwt-strategy.adr.md
│   └── auth-redesign.prd.md
├── backend/
│   └── error-wrapping.rule.md
├── incidents/
│   └── connection-pool-exhaustion.cpat.md
└── notifications/
    └── notifications-implementation.plan.md
```

The structure is **free-form** — organize documents by domain, feature, team, or whatever fits your repo. Categories are virtual and inferred from the document type in the filename (`slug.type.md`).

Use `.archcore/` for:

- architecture decisions
- coding rules and conventions
- implementation plans
- product requirements
- incidents and postmortems
- reusable workflow knowledge

See the Archcore CLI repository itself for a working example: [`.archcore/` in this repo](https://github.com/archcore-ai/cli/tree/main/.archcore)

## Why Not Just Instruction Files?

Files like `CLAUDE.md`, `AGENTS.md`, or repository instructions are useful, but they break down when your team needs:

- more than one flat memory file
- structured document types like ADRs, rules, plans, and incidents
- reusable context across multiple AI tools
- versioned project knowledge that grows with the codebase
- relations between documents, like a plan implementing a PRD
- incident learnings and recurring workflows that agents can reuse later

Archcore complements agent-native memory by giving your repo a durable architectural memory layer.

## Try It

After `archcore init`, open your AI agent and start talking. The agent already knows your existing documents and has tools to create new ones.

> “We decided to use PostgreSQL instead of MongoDB for our primary database. Record this decision.”

Creates `infrastructure/use-postgres.adr.md` with context, decision, alternatives considered, and consequences.

> “We have a team convention: always wrap errors with context using fmt.Errorf and %w. Make this a rule.”

Creates `backend/error-wrapping.rule.md` with imperative guidance, rationale, and good/bad code examples.

> “Last week we had a connection pool exhaustion incident because idle connections weren't being recycled. Document this so we don't repeat it.”

Creates `incidents/connection-pool-exhaustion.cpat.md` with root-cause analysis and prevention steps.

> “I need a PRD for the user notifications feature — push, email digests, and in-app alerts.”

Creates `notifications/user-notifications.prd.md` with goals, user stories, requirements, and success metrics.

> “Create an implementation plan for the notifications PRD and link them together.”

Creates `notifications/notifications-implementation.plan.md`, then links it to the PRD with an `implements` relation.

## Document Types

Archcore organizes context into 3 layers of knowledge: Vision, Knowledge, and Experience.

| Type | Full Name | Category | Description |
|------|-----------|----------|-------------|
| `prd` | Product Requirements Document | Vision | Goals, user stories, acceptance criteria, and success metrics |
| `idea` | Idea | Vision | Lightweight capture of a product or technical idea for future exploration |
| `plan` | Plan | Vision | Phased task list with acceptance criteria and dependencies |
| `adr` | Architecture Decision Record | Knowledge | Captures a finalized technical decision with context, alternatives, and consequences |
| `rfc` | Request for Comments | Knowledge | Proposes a significant change open for team review and feedback |
| `rule` | Rule | Knowledge | Coding or process standard with imperative guidance and examples |
| `guide` | Guide | Knowledge | Step-by-step instructions for completing a specific task |
| `doc` | Document | Knowledge | Reference documentation, registries, and descriptive material |
| `task-type` | Task Type | Experience | Reusable checklist and workflow for a recurring task |
| `cpat` | Code Change Pattern | Experience | Root-cause analysis of a bug or incident with prevention steps |

Each document is a Markdown file with YAML frontmatter:

```markdown
---
title: "Use PostgreSQL for Primary Storage"
status: draft
---

## Context
...
```

Valid statuses: `draft`, `accepted`, and `rejected`.

## Document Relations

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
| `archcore update` | Update Archcore to the latest version |
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
| `sync` | Sync mode. Cloud and on-prem are coming soon. | `none` (local only), `cloud`, `on-prem` |
| `language` | Document language. Helps the agent generate documentation in the right language. | String, defaults to `en` |

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

```text
├── cmd/              # Cobra commands (init, doctor, config, validate, hooks, mcp, ...)
├── internal/
│   ├── agents/       # Supported AI agents with hooks/MCP capabilities
│   ├── api/          # HTTP client for archcore server
│   ├── config/       # Settings management and directory init
│   ├── display/      # Terminal output formatting (lipgloss)
│   ├── update/       # Self-update logic (version check, download, verify, replace)
│   ├── mcp/          # MCP stdio server implementation
│   └── sync/         # Sync logic
├── templates/        # Document type templates
├── install.sh        # Install script
└── .goreleaser.yaml  # Release configuration
```

## Links & License

- **Documentation:** [docs.archcore.ai](https://docs.archcore.ai)
- **Website:** [archcore.ai](https://archcore.ai)
- **Issues:** [github.com/archcore-ai/cli/issues](https://github.com/archcore-ai/cli/issues)
- **License:** [Apache 2.0](LICENSE)
