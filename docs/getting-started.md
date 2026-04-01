# Getting Started

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running
- An AI coding agent installed on your host (Claude Code, Gemini CLI, or Codex) — Asylum seeds its config from your host installation

## Install Asylum

```sh
curl -fsSL https://raw.githubusercontent.com/inventage-ai/asylum/main/install.sh | sh
```

This detects your OS and architecture and downloads the correct binary.

## First Run

```sh
cd your-project/
asylum
```

On the very first run, Asylum:

1. **Builds the base image** (`asylum:latest`) — a full dev toolchain with all three agent CLIs, language runtimes, and tools. This takes ~5 minutes and is cached for all projects.
2. **Seeds agent config** — copies your host agent configuration (e.g., `~/.claude`) to `~/.asylum/agents/claude/`. This is a one-time copy; the asylum copy is independent after seeding.
3. **Starts a container** — mounts your project directory, caches, and agent config, then launches the agent.

Subsequent runs start in seconds since the image is cached.

## What Happens Inside

When the container starts, you get:

- Your project mounted at its **real host path** (not `/workspace`) — absolute paths and git worktrees work correctly
- Agent config, SSH keys, git config, and shell history mounted from `~/.asylum/`
- Package caches (npm, pip, maven, gradle) persisted in named Docker volumes
- A full dev environment: Python 3 + uv, Node.js LTS + fnm, Java 17/21/25 + mise, build tools, and more

## Choosing an Agent

```sh
asylum              # Claude Code (default)
asylum -a gemini    # Gemini CLI
asylum -a codex     # Codex
```

Each agent runs in YOLO mode by default (auto-approve all actions). Agent config is stored separately per agent in `~/.asylum/agents/<agent>/`.

## Project Configuration

Create a `.asylum` file in your project root to customize the environment:

```yaml
agent: gemini

kits:
  java:
    versions: ["17"]
  node:
    packages:
      - "@anthropic-ai/claude-mcp-server-filesystem"

ports:
  - "3000"
  - "8080:80"

env:
  DEBUG: "true"
```

You can also use `asylum config` to interactively toggle kits, credentials, and isolation mode. See [Configuration](configuration/index.md) for all options, and [Kits](kits/index.md) for available language and tool kits.

## SSH

SSH keys are managed automatically by the always-on [SSH kit](kits/ssh.md). On first container start, an Ed25519 key pair is generated at `~/.asylum/ssh/` and mounted into `~/.ssh/`. Add the printed public key to GitHub/GitLab.

You can configure SSH isolation in your config:

```yaml
kits:
  ssh:
    isolation: isolated   # default — generated keys in ~/.asylum/ssh/
    # isolation: shared   # mount host ~/.ssh/ directly
    # isolation: project  # per-project keys
```

## Next Steps

- [Commands](commands/index.md) — all available commands and flags
- [Configuration](configuration/index.md) — layered YAML config system
- [Kits](kits/index.md) — language and tool kits
- [Concepts](concepts/index.md) — how images, mounts, and sessions work
