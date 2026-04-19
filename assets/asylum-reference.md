# Asylum Reference

Asylum is an agent-agnostic Docker sandbox for AI coding agents. This document describes how it works so you can help the user troubleshoot issues, configure their environment, and understand the sandbox.

Changelog: https://github.com/inventage-ai/asylum/blob/main/CHANGELOG.md
Repository: https://github.com/inventage-ai/asylum

## Container Lifecycle

1. **First run**: Asylum builds a base Docker image with all tools, then a project-specific image on top (for custom packages, Java versions, etc.)
2. **Container start**: A detached container is started with the project directory bind-mounted at its real host path. The entrypoint configures git, SSH, direnv, and language managers, then sleeps.
3. **Agent exec**: Asylum runs `docker exec` to start the agent (or shell) inside the running container.
4. **Multiple sessions**: Additional `asylum` invocations attach to the same container. The container stays running until all sessions exit.
5. **Cleanup**: When the last session exits, the container is removed. Named volumes (caches, node_modules shadows) persist.

Container names are deterministic: `asylum-<sha256(project_dir)[:12]>`. This means the same project directory always maps to the same container name.

## Configuration

Asylum uses layered YAML configuration. Each layer overrides the previous:

1. `~/.asylum/config.yaml` — global defaults (created on first run)
2. `.tool-versions` — Java version from asdf/mise format (overrides global, overridden by project)
3. `<project>/.asylum` — project-specific config (commit to repo)
4. `<project>/.asylum.local` — local overrides (gitignore this)
5. CLI flags — highest priority

### Merge Rules

- **Kits**: last-wins (project config replaces global kit map entirely)
- **Ports, Volumes**: concatenated (all layers contribute)
- **Env vars**: merged (later layers override specific keys)
- **Agent, Java version**: last-wins

### Config Options

```yaml
version: "0.2"

# Agent to start by default
agent: claude  # claude, gemini, codex, opencode

# Release channel for self-update
release-channel: stable  # stable, dev

# Agent CLIs to install in the container image
agents:
  claude:
  # gemini:
  # codex:

# Kits configure language toolchains and tools
kits:
  docker:                    # Docker-in-Docker support

  java:
    versions: [17, 21, 25]   # JDK versions to install
    default-version: 21      # Default JDK version
  python:
    # packages: [ansible]    # Extra tools installed via uv
  node:
    shadow-node-modules: true # Shadow node_modules with Docker volumes
    # packages: [turbo]      # Extra npm packages installed globally

  # github:                  # GitHub CLI (gh) (on by default)
  # openspec:                # OpenSpec CLI (on by default)
  # shell:                   # Custom Dockerfile/entrypoint steps (on by default)

  # Disable a default-on kit:
  # github:
  #   disabled: true

  # System packages via apt:
  # apt:
  #   packages: [imagemagick, ffmpeg]

  # Custom build commands:
  # shell:
  #   build:
  #     - "curl -fsSL https://example.com/install.sh | sh"

# Port forwarding (host:container or just port for same on both sides)
# ports:
#   - "3000"
#   - "8080:80"

# Additional volume mounts
# volumes:
#   - ~/shared-data:/data
#   - ~/.aws

# Environment variables
# env:
#   GITHUB_TOKEN: ghp_xxx
#   NODE_ENV: development
```

## Kits

Kits are modular bundles of tools, language runtimes, and configuration. A kit is active when its key is present in the `kits` config map.

### Available Kits

| Kit | Description | Default |
|-----|-------------|---------|
| `docker` | Docker Engine (Docker-in-Docker, privileged mode) | No |
| `java` | JDK via mise (17, 21, 25) | No |
| `java/maven` | Maven build tool | No (activated with java) |
| `java/gradle` | Gradle via mise | No (activated with java) |
| `python` | Python tools via uv (black, ruff, mypy, pytest, poetry, etc.) | No |
| `python/uv` | Auto-creates .venv for Python projects | No (activated with python) |
| `node` | Node.js global packages (typescript, eslint, prettier, etc.) | No |
| `node/npm` | npm caching and onboarding | No (activated with node) |
| `node/pnpm` | pnpm package manager | No (activated with node) |
| `node/yarn` | yarn package manager | No (activated with node) |
| `ports` | Automatic port forwarding for web services | Yes |
| `github` | GitHub CLI (gh) | Yes |
| `openspec` | OpenSpec CLI | Yes |
| `shell` | oh-my-zsh, tmux, direnv hooks | Yes |

### Kit Hierarchy

Top-level kits like `java` automatically activate all their sub-kits (`java/maven`, `java/gradle`). To activate only a specific sub-kit, use the full path: `java/maven`.

### Default-On Kits

Some kits (`ports`, `github`, `openspec`, `shell`) are active by default even without explicit config. To disable them:

```yaml
kits:
  github:
    disabled: true
```

## Installing Additional Tools

To install additional tools in the container, prefer system packages over shell build commands. Packages are cached by Docker's layer system and install faster on rebuilds.

**Preferred: apt packages**

```yaml
kits:
  apt:
    packages: [golang, postgresql-client, redis-tools]
```

**Also available: kit-specific packages**

```yaml
kits:
  node:
    packages: [turbo, nx]     # npm global packages
  python:
    packages: [ansible, aws-cli]  # uv tools
```

**Last resort: custom build commands**

Use `shell.build` only for tools that aren't available as packages:

```yaml
kits:
  shell:
    build:
      - "curl -fsSL https://example.com/install.sh | sh"
```

Build commands run during `docker build` (cached in the image layer), not on every container start.

## Port Forwarding

### Automatic Ports (ports kit)

The `ports` kit (default-on) automatically allocates a range of high host ports per project and forwards them into the container. This allows agents to start web servers without manual port configuration.

- **Default count**: 5 ports per project (configurable via `kits: { ports: { count: 10 } }`)
- **Starting port**: 7001, allocated sequentially (below the browser-restricted 10000+ range)
- **Persistence**: Port assignments are stored in `~/.asylum/ports.json` and reused across container restarts
- **No collisions**: Each project gets a unique, non-overlapping range

The allocated ports appear in the sandbox rules file so the agent knows which ports are available.

To disable automatic port allocation:

```yaml
kits:
  ports:
    disabled: true
```

### Manual Ports

In addition to (or instead of) automatic ports, you can manually forward specific ports:

```yaml
ports:
  - "3000"        # same port on host and container
  - "8080:80"     # host:container
```

Or via CLI: `asylum -p 3000 -p 8080:80`

Manual ports and automatic ports work independently — both are forwarded.

## Volume Mounting

The project directory is bind-mounted at its real host path (not `/workspace`), preserving absolute paths between host and container.

### Volume Shorthand

In config files:
- `/path` — mounts at the same path inside the container
- `/host:/container` — explicit mapping
- `/host:/container:ro` — read-only
- `~/path` — expands `~` to home directory

### Shadow Node Modules

When the `node` kit has `shadow-node-modules: true` (default), each `node_modules` directory is overlaid with a named Docker volume. This prevents host OS-specific binaries from being visible inside the container and persists `npm install` across container restarts.

## Image Management

Asylum uses a two-tier image system:

1. **Base image** (`asylum:latest`): OS, core tools, language managers, kit tools, agent CLIs. Shared across all projects. Rebuilt when kits or agents change.
2. **Project image** (`asylum:proj-<hash>`): Custom packages, specific Java version. Built on top of base. Rebuilt when project-specific config changes.

Images are hash-tagged. If the assembled content hasn't changed, the existing image is reused. A base image rebuild cascades to all project images.

To force a rebuild: `asylum --rebuild`

## Self-Update

Asylum can update itself:

```
asylum self-update              # Update to latest stable release
asylum self-update --dev        # Update to latest dev build
asylum self-update 0.4.0        # Install a specific version
```

## SSH

SSH is managed by an always-on kit. On first container start, an ed25519 key pair is generated automatically at `~/.asylum/ssh/` and mounted into `~/.ssh/`. The host's `~/.ssh/known_hosts` is mounted read-write if it exists.

Configure isolation in `.asylum` or `~/.asylum/config.yaml`:

```yaml
kits:
  ssh:
    isolation: isolated   # default — generated keys in ~/.asylum/ssh/
    # isolation: shared   # mount host ~/.ssh/ directly (read-write)
    # isolation: project  # per-project keys in ~/.asylum/projects/<container>/ssh/
```

## Troubleshooting

### Container won't start
- Check Docker is running: `docker info`
- Force rebuild: `asylum --rebuild`
- Check logs: `docker logs asylum-<hash>`

### Tools not found
- The container may be using a cached image. Run `asylum --rebuild` to rebuild.
- Check if the kit providing the tool is active in your config.

### Permission issues
- The container user is `claude` with passwordless sudo.
- Mounted directories inherit host permissions. If files are read-only, check host permissions.

### Git issues inside container
- All mounted directories are marked as safe (`safe.directory = *`).
- On Docker Desktop (linuxkit), `core.fileMode` is set to false to handle permission differences.
- Host `.gitconfig` is copied into the container at startup.

### Node.js issues
- If `npm install` produces platform-specific errors, the shadow node_modules volume may contain stale data from a previous architecture. Run `asylum --rebuild` to clear it.

### Checking for updates
If you encounter a bug, check the changelog to see if it's been fixed in a newer version:
https://github.com/inventage-ai/asylum/blob/main/CHANGELOG.md

Update with: `asylum self-update`
