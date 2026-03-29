# Asylum

Docker sandbox for AI coding agents. Single Go binary, cross-compiled for Linux and macOS (ARM and x86).

Asylum wraps Docker to give [Claude Code](https://claude.ai), [Gemini CLI](https://github.com/google-gemini/gemini-cli), and [Codex](https://github.com/openai/codex) a full development environment — while keeping your host clean. Containers are ephemeral, but caches, auth, and history persist across sessions.

## How Does It Compare?

| | Asylum | [Claudebox](https://github.com/RchGrav/claudebox) | [AgentBox](https://github.com/fletchgqc/agentbox) | [Safehouse](https://github.com/eugene1g/agent-safehouse) |
|---|---|---|---|---|
| **Approach** | Docker container | Docker container | Docker/Podman container | macOS kernel sandbox |
| **Agents** | Claude, Gemini, Codex | Claude only | Claude, OpenCode | 13+ agents |
| **Platform** | Linux, macOS | Linux, macOS | Linux, macOS (Bash 4+) | macOS only |
| **Distribution** | Single Go binary | Self-extracting installer | Git clone + alias | Homebrew / shell script |
| **Languages** | Python, Node.js, Java | 15+ profiles | Python, Node.js, Java | Host environment |
| **Docker-in-Docker** | Yes | No | No | N/A |
| **Config** | Layered YAML | INI profiles | `.env` files | `.sb` policy files |
| **Mount strategy** | Real host path | `/workspace` | Host path | N/A (host-native) |
| **Extensibility** | Kit system | Profiles | Single image | Policy profiles |

Asylum is the only option with Docker-in-Docker support, multi-agent support in a single tool, and a modular kit system. Projects are mounted at their real host path (not `/workspace`), so absolute paths and git worktrees work correctly.

## What's Included

**Languages & Runtimes**

- Python 3 with [uv](https://github.com/astral-sh/uv), black, ruff, mypy, pytest, poetry
- Node.js LTS with [fnm](https://github.com/Schniz/fnm), TypeScript, eslint, prettier
- Java 17/21/25 with [mise](https://mise.jdx.dev/), Maven, Gradle

**Tools**

- Docker Engine with buildx and compose (Docker-in-Docker)
- GitHub CLI (`gh`)
- Build tools: gcc, g++, make, cmake
- Dev tools: git, vim, nano, tmux, htop, ripgrep, fd, jq, yq, direnv
- Shell: zsh with oh-my-zsh

**Plugins**

- [OpenSpec](https://openspec.dev) CLI for structured change management

Everything is modular — see [Kits](https://asylum.inventage.ai/kits/) for the full list.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/inventage-ai/asylum/main/install.sh | sh
```

Or download a binary from the [releases page](https://github.com/inventage-ai/asylum/releases).

**Requires**: [Docker](https://docs.docker.com/get-docker/) installed and running.

## Quick Start

```sh
cd your-project/

asylum              # Start Claude Code (default)
asylum -a gemini    # Start Gemini CLI
asylum -a codex     # Start Codex
asylum shell        # Interactive shell (no agent)
```

On first run, Asylum builds a Docker image (~5 min) and seeds agent config from your host. Subsequent runs start in seconds.

## Commands

| Command | Description |
|---------|-------------|
| `asylum` | Start the default agent |
| `asylum shell` | Interactive zsh shell |
| `asylum run <cmd>` | Run a command in the container |
| `asylum cleanup` | Remove images and cached data |
| `asylum version` | Show version |
| `asylum ssh-init` | Set up SSH keys |
| `asylum self-update` | Update to latest version |

Full reference: [Commands documentation](https://asylum.inventage.ai/commands/)

## Configuration

Layered YAML config: `~/.asylum/config.yaml` (global) → `.asylum` (project) → `.asylum.local` (local overrides) → CLI flags.

```yaml
agent: gemini

kits:
  java:
    default-version: "17"
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

Full reference: [Configuration documentation](https://asylum.inventage.ai/configuration/)

## Kits

Kits are modular bundles for languages and tools. Enable them in your config:

| Kit | Description | Default |
|-----|-------------|---------|
| `node` | Node.js LTS + npm/pnpm/yarn | Off |
| `python` | Python 3 + uv | Off |
| `java` | JDK 17/21/25 + Maven/Gradle | Off |
| `docker` | Docker-in-Docker | Off |
| `github` | GitHub CLI | **On** |
| `openspec` | OpenSpec CLI | **On** |
| `shell` | oh-my-zsh, tmux, direnv | **On** |
| `apt` | Extra system packages | Off |

Full reference: [Kits documentation](https://asylum.inventage.ai/kits/)

## Building from Source

```sh
git clone https://github.com/inventage-ai/asylum.git
cd asylum
make build
```

## License

MIT
