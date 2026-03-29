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
asylum [flags]                Start default agent
asylum [flags] -- [args]      Start agent with extra args
asylum [flags] shell          Interactive zsh shell
asylum [flags] shell --admin  Admin shell with sudo notice
asylum [flags] run <cmd>      Run command in container
asylum cleanup                Remove Asylum images and cached data
asylum version [--short]      Show version
asylum ssh-init               Set up SSH keys for containers
asylum self-update [version]  Update to latest (or specific) version
```

### Flags

| Flag | Description |
|------|-------------|
| `-a`, `--agent` | Agent to use: `claude`, `gemini`, `codex` (default: `claude`) |
| `-p <port>` | Forward a port (repeatable, e.g. `-p 3000 -p 8080:80`) |
| `-v <volume>` | Mount a volume (repeatable, e.g. `-v ~/data:/data:ro`) |
| `-e KEY=VALUE` | Set environment variable (repeatable, last wins) |
| `--java <version>` | Select Java version (`17`, `21`, `25` pre-installed; others installed on demand). Auto-detected from `.tool-versions`. |
| `-n`, `--new` | Start a fresh session (skip auto-resume) |
| `--rebuild` | Force rebuild the Docker image |
| `--skip-onboarding` | Skip project onboarding tasks for this run |
| `--cleanup` | Alias for `cleanup` command |
| `--version` | Alias for `version` command |

Use `--` to pass extra flags to the agent (e.g. `asylum -- --verbose`). Use `run` to execute commands in the container (e.g. `asylum run python test.py`).

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

env:
  MY_API_KEY: "abc123"
  DEBUG: "true"

versions:
  java: "17"

tab-title: "🤖 {project}"    # Terminal tab title ({project}, {agent}, {mode})

features:
  session-name: true              # Name new Claude sessions after project dir
  allow-agent-terminal-title: true  # Let the agent set the terminal tab title

onboarding:
  npm: true                       # Auto-install Node.js deps on first container start

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

### Tab title

Asylum sets the terminal tab/window title before starting the container. The default is `🤖 projectname`. Customize it with `tab-title` using placeholders:

| Placeholder | Value |
|-------------|-------|
| `{project}` | Project directory basename |
| `{agent}` | Agent name (claude, gemini, codex) |
| `{mode}` | Mode (agent, shell, admin, run) |

By default, Claude Code is prevented from overriding this title. Set `allow-agent-terminal-title: true` in `features` to let it.

### Project onboarding

Asylum can automatically run setup tasks (like `npm install`) when a container is first created. Onboarding is opt-in per task:

```yaml
onboarding:
  npm: true    # Auto-install Node.js dependencies (detects lockfiles)
```

When enabled, asylum scans for lockfiles (`package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lock`), shows a prompt listing what it found, and runs the appropriate install command inside the container. State is tracked — it won't re-prompt unless a lockfile changes.

Skip onboarding for a single run with `--skip-onboarding`, or disable it globally with `features: { onboarding: false }`.

### Feature flags

Boolean flags in `features` control opt-in/opt-out behaviors:

| Flag | Default | Description |
|------|---------|-------------|
| `session-name` | `false` | Name new Claude sessions after the project directory |
| `allow-agent-terminal-title` | `false` | Let the agent set the terminal tab title (overrides asylum's `tab-title`) |
| `shadow-node-modules` | `true` | Shadow `node_modules` with named volumes to isolate host native binaries |
| `onboarding` | `true` | Enable the onboarding system (individual tasks still need opt-in via `onboarding` map) |

### Merge rules

- **Scalars** (agent, java version, tab-title): last value wins
- **Lists** (ports, volumes): concatenated across layers
- **Maps** (env, versions, features): merged per key, last value wins
- **Package lists** (apt, npm, pip, run): each sub-list concatenated independently

## How It Works

### Images

Asylum uses a two-tier image strategy:

1. **Base image** (`asylum:latest`) — full dev toolchain with all three agent CLIs, shared across projects
2. **Project image** (`asylum:proj-<hash>`) — adds packages from your config, only built when `packages` is set

Images auto-rebuild when the Dockerfile or your packages config changes (hash-based detection).

### What's in the container

- **Languages**: Python 3 + uv, Node.js LTS + fnm, Java 17/21/25 + mise
- **Build tools**: gcc, g++, make, cmake
- **Package managers**: npm, yarn, pnpm, pip, uv, poetry, maven, gradle
- **Dev tools**: git, vim, nano, tmux, htop, ripgrep, fd, jq, yq, direnv
- **Git tools**: gh (GitHub CLI), glab (GitLab CLI)
- **Docker**: Full engine with buildx and compose (Docker-in-Docker)
- **Shell**: zsh with oh-my-zsh
- **Agents**: Claude Code, Gemini CLI, Codex

### Mounts

Your project is mounted at its **real host path** inside the container (not `/workspace`). This preserves absolute paths and makes git worktrees work correctly.

| What | Host Path | Scope |
|------|-----------|-------|
| Project directory | `$PWD` | Per project |
| Git config | `~/.gitconfig` | Global (read-only) |
| SSH keys | `~/.asylum/ssh/` | Global |
| Agent config | `~/.asylum/agents/<agent>/` | Global (per agent) |
| Package caches (npm, pip, maven, gradle) | Named Docker volumes | Per project |
| Shell history | `~/.asylum/projects/<id>/history` | Per project |
| Direnv approvals | `~/.local/share/direnv/allow` | Global (read-only) |
| `.env` file | `$PWD/.env` | Per project (env vars only) |

On first run, agent config is **seeded** (one-time copy) from the host native config directory (e.g. `~/.claude` → `~/.asylum/agents/claude/`). After that, the asylum copy is independent — changes to the host config won't propagate.

### Shadow node_modules

On macOS, Node.js native binaries built on the host won't work inside the Linux container. Asylum automatically shadows `node_modules` directories with named Docker volumes so each platform has its own binaries. This is transparent — your source files are shared, only `node_modules` is isolated.

Disable with `features: { shadow-node-modules: false }`.

### Sessions

Multiple terminal sessions can share a single container. The first `asylum` invocation starts the container; subsequent ones exec into it. The container is automatically removed when the last session exits.

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

## Self-Update

```sh
asylum self-update          # Update to latest stable release
asylum self-update 0.4.0   # Install a specific version
asylum self-update --dev    # Update to latest dev build from main
asylum self-update --safe   # Emergency update (always dev, no checks)
```

To always track dev builds, set `release-channel: dev` in your config:

```yaml
# ~/.asylum/config.yaml
release-channel: dev
```

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
