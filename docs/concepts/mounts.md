# Mounts

Asylum mounts your project, config, caches, and tools into the container. Your project is mounted at its **real host path** — not `/workspace` — so absolute paths and git worktrees work correctly.

## Default Mounts

| What | Host Path | Container Path | Mode |
|------|-----------|----------------|------|
| Project directory | `$PWD` | `$PWD` | Read-write |
| Git config | `~/.gitconfig` | `/tmp/host_gitconfig` | Read-only |
| SSH keys | Depends on [SSH isolation](../kits/ssh.md) | `~/.ssh/` | Read-only (keys), Read-write (known_hosts) |
| Agent config | Depends on [isolation mode](isolation.md) | Agent-specific path | Read-write |
| Shell history | `~/.asylum/projects/<id>/history/` | `~/.shell_history/` | Read-write |
| Claude temp root (Claude sessions only) | `~/.asylum/projects/<id>/tmp/` | Same path | Read-write |
| Direnv approvals | `~/.local/share/direnv/allow` | Same path | Read-only |
| `.env` file | `$PWD/.env` | Loaded as `--env-file` | — |

Claude Code keeps its scratchpad and background task output under `CLAUDE_CODE_TMPDIR`, which Asylum points at the temp root. The files survive container restarts, and any path Claude prints opens unchanged on the host. Asylum never prunes the directory. `asylum cleanup` removes it with the rest of the project data. Other tools keep using the container's local `/tmp`.

## Cache Volumes

Package caches are stored in named Docker volumes (not host bind mounts), scoped per project:

| Cache | Container Path | Kit |
|-------|---------------|-----|
| npm | `~/.npm` | node/npm |
| pip | `~/.cache/pip` | python/uv |
| Maven | `~/.m2` | java/maven |
| Gradle | `~/.gradle` | java/gradle |

Named volumes persist across container restarts and are shared between sessions on the same project.

## Shadow node_modules

When the Node.js kit is active (with `shadow-node-modules: true`, the default), Asylum creates a named Docker volume for each `node_modules` directory found in the project. This isolates Linux-built native binaries from macOS host binaries.

See [Node.js Kit — Shadow node_modules](../kits/node.md#shadow-node_modules) for details.

## Git Worktrees

If the project directory is a git worktree, Asylum detects and mounts both the worktree's `.git` file and the main repository's `.git` directory. This ensures git operations work correctly inside the container.

## Custom Volumes

Add extra volumes via config or CLI:

```yaml
# .asylum
volumes:
  - ~/shared-data:/data:ro
  - /tmp/exchange:/tmp/exchange
```

```sh
asylum -v ~/data:/data:ro
```

Tilde (`~`) is expanded. Supported options: `ro`, `rw`, `z`, `Z`, `shared`, `cached`, `delegated`, and others.

## Host IP Access

The host machine is accessible from inside the container via `host.docker.internal`.
