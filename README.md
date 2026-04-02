# Asylum

**Docker sandbox for AI coding agents.** Run Claude Code in a full dev environment — your host stays clean, your tools stay fast. Experimental support for Gemini CLI and Codex.

[![CI](https://github.com/inventage-ai/asylum/actions/workflows/ci.yml/badge.svg)](https://github.com/inventage-ai/asylum/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/inventage-ai/asylum)](https://github.com/inventage-ai/asylum/releases)
[![Docs](https://img.shields.io/badge/docs-asylum.inventage.ai-blue)](https://asylum.inventage.ai/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Asylum wraps Docker to give your AI coding agent a fully-equipped Linux environment with Python, Node.js, Java, Docker-in-Docker, and more. Containers are ephemeral, but caches, auth, shell history, and agent config persist across sessions. A single Go binary, cross-compiled for Linux and macOS (ARM and x86).

**[Read the documentation &rarr;](https://asylum.inventage.ai/latest/)**

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/inventage-ai/asylum/main/install.sh | sh
```

Or download a binary from the [releases page](https://github.com/inventage-ai/asylum/releases) (you'll need to `chmod +x` it before use).

**Requires:** [Docker](https://docs.docker.com/get-docker/) installed and running.

## Quick Start

```sh
cd your-project/

asylum              # Start Claude Code (default)
asylum -a gemini    # Start Gemini CLI
asylum -a codex     # Start Codex
asylum shell        # Interactive shell (no agent)
```

On first run, Asylum builds a Docker image (~5 min) and seeds agent config from your host. Subsequent runs start in seconds.

## Why Asylum?

- **Built for Claude Code** — first-class support, with experimental Gemini CLI and Codex support
- **Docker-in-Docker** — build and run containers inside the sandbox
- **Real host paths** — project mounted at its actual path, so absolute paths and git worktrees work
- **Persistent caches** — npm, pip, Maven, Gradle caches survive container restarts
- **Kit system** — modular language/tool bundles you can enable per-project
- **Layered config** — global defaults, project overrides, local tweaks, CLI flags
- **Credential scoping** — mount only the host credentials each project needs (read-only)
- **Single binary** — no runtime dependencies beyond Docker

## What's Included

| Category | Tools |
|----------|-------|
| **Python** | Python 3, [uv](https://github.com/astral-sh/uv), black, ruff, mypy, pytest, poetry |
| **Node.js** | LTS via [fnm](https://github.com/Schniz/fnm), TypeScript, eslint, prettier, npm/pnpm/yarn |
| **Java** | JDK 17/21/25 via [mise](https://mise.jdx.dev/), Maven, Gradle |
| **Docker** | Docker Engine, buildx, compose (Docker-in-Docker) |
| **Dev tools** | git, vim, nano, tmux, htop, ripgrep, fd, jq, yq, direnv, gcc, make, cmake |
| **Shell** | zsh with oh-my-zsh |
| **GitHub** | GitHub CLI (`gh`) |

Everything is modular via [Kits](https://asylum.inventage.ai/latest/kits/) — enable what you need, disable what you don't.

## Configuration

Layered YAML: `~/.asylum/config.yaml` (global) &rarr; `.asylum` (project) &rarr; `.asylum.local` (local) &rarr; CLI flags.

```yaml
agent: gemini

kits:
  java:
    default-version: "17"
    credentials: auto
  docker: {}
  node:
    onboarding: true
    packages:
      - tsx

ports:
  - "3000"

env:
  DEBUG: "true"
```

Full reference: [Configuration documentation](https://asylum.inventage.ai/latest/configuration/)

## Commands

| Command | Description |
|---------|-------------|
| `asylum` | Start the default agent |
| `asylum shell` | Interactive zsh shell |
| `asylum run <cmd>` | Run a command in the container |
| `asylum config` | Configure kits, credentials, and isolation |
| `asylum cleanup` | Clean up current project (`--all` for everything) |
| `asylum version` | Show version |
| `asylum self-update` | Update to latest version |

Full reference: [Commands documentation](https://asylum.inventage.ai/latest/commands/)

## Kits

| Kit | Description | Default |
|-----|-------------|---------|
| [`node`](https://asylum.inventage.ai/latest/kits/node/) | Node.js LTS + npm/pnpm/yarn | Always on |
| [`python`](https://asylum.inventage.ai/latest/kits/python/) | Python 3 + uv | On |
| [`java`](https://asylum.inventage.ai/latest/kits/java/) | JDK 17/21/25 + Maven/Gradle | On |
| [`docker`](https://asylum.inventage.ai/latest/kits/docker/) | Docker-in-Docker | On |
| [`github`](https://asylum.inventage.ai/latest/kits/github/) | GitHub CLI | On |
| [`shell`](https://asylum.inventage.ai/latest/kits/shell/) | oh-my-zsh, tmux, direnv | Always on |
| [`ports`](https://asylum.inventage.ai/latest/kits/ports/) | Automatic port forwarding | Always on |
| [`openspec`](https://asylum.inventage.ai/latest/kits/openspec/) | [OpenSpec](https://openspec.dev) CLI | On |
| [`ast-grep`](https://asylum.inventage.ai/latest/kits/ast-grep/) | AST-based code search (`sg`) | Opt-in |
| [`agent-browser`](https://asylum.inventage.ai/latest/kits/agent-browser/) | Browser automation via agent-browser | Opt-in |
| [`cx`](https://asylum.inventage.ai/latest/kits/cx/) | Semantic code navigation | Opt-in |
| [`apt`](https://asylum.inventage.ai/latest/kits/apt/) | Extra system packages | Opt-in |

See the [full kits reference](https://asylum.inventage.ai/latest/kits/) for configuration details.

## How Does It Compare?

| | Asylum | [Claudebox](https://github.com/RchGrav/claudebox) | [AgentBox](https://github.com/fletchgqc/agentbox) | [Safehouse](https://github.com/eugene1g/agent-safehouse) |
|---|---|---|---|---|
| **Approach** | Docker container | Docker container | Docker/Podman container | macOS kernel sandbox |
| **Agents** | Claude (primary), Gemini, Codex (experimental) | Claude only | Claude, OpenCode | 13+ agents |
| **Platform** | Linux, macOS | Linux, macOS | Linux, macOS (Bash 4+) | macOS only |
| **Distribution** | Single Go binary | Self-extracting installer | Git clone + alias | Homebrew / shell script |
| **Docker-in-Docker** | Yes | No | No | N/A |
| **Config** | Layered YAML | INI profiles | `.env` files | `.sb` policy files |
| **Mount strategy** | Real host path | `/workspace` | Host path | N/A (host-native) |
| **Extensibility** | Kit system | Profiles | Single image | Policy profiles |

## Building from Source

```sh
git clone https://github.com/inventage-ai/asylum.git
cd asylum
make build
```
