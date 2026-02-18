# archcore CLI

Command-line tool for managing Archcore in any directory.

## Build

```bash
go build -o archcore .
```

## Commands

| Command           | Description                               |
| ----------------- | ----------------------------------------- |
| `archcore init`   | Interactive setup — create .archcore/ dir |
| `archcore status` | Show config and connection status         |
| `archcore config` | View or modify configuration              |
| `archcore doctor` | Check setup for issues                    |

## Quick start

```bash
# In any directory:
archcore init        # follow the prompts
archcore status      # verify setup
archcore doctor      # run diagnostics
```

## Configuration

Stored in `.archcore/settings.json`.

```bash
archcore config                              # show all
archcore config get sync                     # get a value
archcore config set sync cloud               # set a value
archcore config set archcore_url http://localhost:8080
```

Keys: `sync`, `project_id`, `archcore_url`.

Sync types: `none`, `cloud`, `on-prem`.

## Directory structure

```
.archcore/
├── settings.json
├── vision/
├── knowledge/
└── experience/
```
