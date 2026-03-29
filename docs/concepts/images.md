# Images

Asylum uses a two-tier image strategy to balance build time with per-project customization.

## Base Image

**Tag:** `asylum:latest`

The base image contains the full dev toolchain shared across all projects:

- **OS**: Debian Trixie
- **Languages**: Python 3 + uv, Node.js LTS + fnm, JDK 17/21/25 + mise
- **Build tools**: gcc, g++, make, cmake
- **Shell**: zsh, bash, oh-my-zsh
- **Dev tools**: git, vim, nano, tmux, htop, ripgrep, fd, jq, yq, direnv
- **Network**: curl, wget, openssh-client, netcat
- **Docker CLI** (not the engine — see the [Docker kit](../kits/docker.md) for Docker-in-Docker)
- **Agent CLIs**: Claude Code, Gemini CLI, Codex

The base image is built once (~5 minutes) and cached. It rebuilds automatically when the Dockerfile or kit snippets change (hash-based detection).

## Project Image

**Tag:** `asylum:proj-<hash>`

When your config includes [packages](../configuration/packages.md) (apt, npm, pip, or custom build commands), Asylum builds a project image on top of the base:

```
┌──────────────────────────────┐
│  Project Image               │
│  (apt packages, npm globals, │
│   pip tools, custom builds)  │
├──────────────────────────────┤
│  Base Image                  │
│  (languages, agents, tools)  │
└──────────────────────────────┘
```

If no packages are configured, the base image is used directly — no project image is built.

Project images are also cached and only rebuild when the package configuration changes.

## Rebuild Detection

Both tiers use hash-based detection:

- **Base image**: SHA-256 of the Dockerfile, kit snippets, agent install snippets, and host user identity (username, UID, GID, home directory). Stored as a Docker label (`asylum.hash`).
- **Project image**: SHA-256 of the packages config. Stored as a Docker label (`asylum.packages.hash`).

A base image rebuild invalidates all project images (they're built `FROM asylum:latest`).

## Forcing a Rebuild

```sh
asylum --rebuild
```

This forces both tiers to rebuild, ignoring the cached hashes.

## Cleanup

Remove all Asylum images with:

```sh
asylum cleanup
```

See [`cleanup`](../commands/cleanup.md) for details.
