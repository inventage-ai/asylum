# Asylum

Docker sandbox for AI coding agents. Single binary, runs on Linux and macOS (ARM and x86).

Asylum wraps Docker to give [Claude Code](https://claude.ai), [Gemini CLI](https://github.com/google-gemini/gemini-cli), and [Codex](https://github.com/openai/codex) a full development environment with Python, Node.js, Java, and Docker-in-Docker — while keeping your host clean. Containers are ephemeral, but caches, auth, and history persist.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/inventage-ai/asylum/main/install.sh | sh
```

Or download a binary from the [releases page](https://github.com/inventage-ai/asylum/releases) and put it in your PATH.

**Requires**: [Docker](https://docs.docker.com/get-docker/) installed and running.

## Quick Start

```sh
cd your-project/

# Start Claude Code in YOLO mode (default)
asylum

# Start Gemini CLI
asylum -a gemini

# Start Codex
asylum -a codex

# Interactive shell (no agent)
asylum shell
```

On first run, Asylum builds a Docker image (~5 min) and seeds agent config from your host. Subsequent runs start in seconds.

## Usage

```
asylum                     Start default agent in YOLO mode
asylum -a gemini           Start Gemini CLI
asylum -a codex            Start Codex
asylum shell               Interactive zsh shell
asylum shell --admin       Shell with sudo notice
asylum ssh-init            Set up SSH keys for containers
asylum <cmd> [args...]     Run any command in the container
```

### Flags

| Flag | Description |
|------|-------------|
| `-a`, `--agent` | Agent to use: `claude`, `gemini`, `codex` (default: `claude`) |
| `-p <port>` | Forward a port (repeatable, e.g. `-p 3000 -p 8080:80`) |
| `-v <volume>` | Mount a volume (repeatable, e.g. `-v ~/data:/data:ro`) |
| `--java <version>` | Select Java version (`17`, `21`, `25`) |
| `-n`, `--new` | Start a fresh session (skip auto-resume) |
| `--rebuild` | Force rebuild the Docker image |
| `--cleanup` | Remove Asylum images and cached data |
| `--version` | Show version |

Flags not recognized by Asylum are passed through to the agent.

## Configuration

Asylum reads config from three YAML files, merged in order:

| Priority | File | Purpose |
|----------|------|---------|
| 1 (lowest) | `~/.asylum/config.yaml` | Global defaults |
| 2 | `.asylum` | Project config (commit this) |
| 3 (highest) | `.asylum.local` | Local overrides (gitignore this) |

CLI flags override all config files.

### Example `.asylum`

```yaml
agent: gemini

ports:
  - "3000"
  - "8080:80"

volumes:
  - ~/shared-data:/data:ro

versions:
  java: "17"

packages:
  apt:
    - libpq-dev
    - redis-tools
  npm:
    - "@anthropic-ai/claude-mcp-server-filesystem"
  pip:
    - pandas
    - numpy
  run:
    - "curl -fsSL https://deno.land/install.sh | sh"
```

### Merge rules

- **Scalars** (agent, java version): last value wins
- **Lists** (ports, volumes): concatenated across layers
- **Package lists** (apt, npm, pip, run): each sub-list concatenated independently

## How It Works

### Images

Asylum uses a two-tier image strategy:

1. **Base image** (`asylum:latest`) — full dev toolchain with all three agent CLIs, shared across projects
2. **Project image** (`asylum:proj-<hash>`) — adds packages from your config, only built when `packages` is set

Images auto-rebuild when the Dockerfile or your packages config changes (hash-based detection).

### What's in the container

- **Languages**: Python 3 + uv, Node.js LTS + NVM, Java 17/21/25 + SDKMAN
- **Build tools**: gcc, g++, make, cmake
- **Package managers**: npm, yarn, pnpm, pip, uv, poetry, maven, gradle
- **Dev tools**: git, vim, nano, tmux, htop, ripgrep, fd, jq, yq, direnv
- **Git tools**: gh (GitHub CLI), glab (GitLab CLI)
- **Docker**: Full engine with buildx and compose (Docker-in-Docker)
- **Shell**: zsh with oh-my-zsh
- **Agents**: Claude Code, Gemini CLI, Codex

### Mounts

Your project is mounted at its **real host path** inside the container (not `/workspace`). This preserves absolute paths and makes git worktrees work correctly.

Asylum also mounts:
- Git config (read-only)
- SSH keys from `~/.asylum/ssh/`
- Package caches (npm, pip, maven, gradle) — persisted per project
- Shell history — persisted per project
- Agent config from `~/.asylum/agents/<agent>/`
- `.env` file loaded automatically

### Sessions

Agents auto-resume previous sessions by default. Use `-n` to start fresh. Session detection is per-project (keyed by absolute path).

## SSH Setup

```sh
asylum ssh-init
```

Creates `~/.asylum/ssh/` with an Ed25519 key pair. Add the public key to GitHub/GitLab, or replace with your own keys. These are mounted into every container.

## Cleanup

```sh
asylum --cleanup
```

Removes Asylum Docker images and optionally clears caches. Agent config (`~/.asylum/agents/`) is preserved since it contains auth tokens.

## Building from Source

```sh
git clone https://github.com/inventage-ai/asylum.git
cd asylum
make build          # Build for current platform
make build-all      # Cross-compile all targets
make test           # Run tests
```

## License

MIT
