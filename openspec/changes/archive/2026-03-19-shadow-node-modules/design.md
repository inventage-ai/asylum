## Context

Asylum mounts the host project directory into the container at its real path. Node.js projects with `node_modules` on the host contain platform-specific native binaries that don't work on Linux. This is the most common cause of "binary not found" or "exec format error" inside containers.

## Goals / Non-Goals

**Goals:**
- Host `node_modules` with native binaries are invisible inside the container.
- Dependencies installed in-container persist across container restarts.
- Works for monorepos with multiple `node_modules` directories.
- Can be disabled per-project.

**Non-Goals:**
- Auto-installing dependencies (handled by a separate feature).
- Sharing `node_modules` between different projects.

## Decisions

### 1. Named Docker volumes over anonymous volumes

Named volumes persist independently of the container lifecycle. Anonymous volumes are deleted when the container is removed, losing any installed dependencies. Named volumes also show up in `docker volume ls` for debugging.

Volume names follow a fixed-length pattern: `<container-name>-npm-<hash>` where `<hash>` is the first 11 hex chars of SHA-256 of the relative path from project root to `node_modules`. Example: `asylum-a1b2c3d4e5f6-npm-16b61a18f68`.

### 2. `--mount` syntax over `-v`

Using `--mount type=volume,src=<name>,dst=<path>` instead of `-v name:path` because Docker's `--mount` is stricter and doesn't silently create bind mounts if the source doesn't look like a path.

### 3. Walk with early exits for performance

`findNodeModulesDirs` walks the project tree looking for `package.json` files and returns the `node_modules` path next to each one — whether or not `node_modules` exists yet. This ensures fresh clones get shadow volumes before `npm install` runs. During the walk, it skips `.git`, `.venv`, `vendor`, `target`, `dist` — directories that never contain relevant `package.json` files. It also skips recursing into `node_modules` itself (packages inside `node_modules` have their own `package.json` but should not get separate shadow volumes).

### 4. Default-on with opt-out via `FeatureOff()`

The feature is useful for all Node.js projects and doesn't change behavior for non-Node projects (no `package.json` → no walk). The `FeatureOff()` config method checks if a feature is explicitly set to `false`, complementing the existing `Feature()` which checks for explicitly `true`.

## Risks / Trade-offs

- **First run is empty**: The shadow volume starts empty, so `npm install` must run inside the container. This is by design — a companion auto-install feature will handle this.
- **Disk usage**: Named volumes accumulate. They're cleaned up by `asylum --cleanup` (which prunes Docker resources) but not automatically on project removal.
- **Walk latency**: Large monorepos with deep directory trees add walk time. Mitigated by skipping common heavy directories and the `package.json` guard.
