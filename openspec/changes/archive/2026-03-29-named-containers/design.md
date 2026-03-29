## Context

Container names are currently `asylum-<sha256[:12]>` (e.g., `asylum-7a3f2b1c9e04`). The hash makes them unique but unreadable. The change appends the sanitized project directory basename so names become `asylum-<hash[:12]>-<project>` (e.g., `asylum-7a3f2b1c9e04-myapp`).

The hash prefix stays because it guarantees uniqueness. The project suffix is purely cosmetic.

## Goals / Non-Goals

**Goals:**
- Human-readable container names in `docker ps`, `docker volume ls`, `~/.asylum/projects/`
- Transparent migration of existing `~/.asylum/projects/<old-name>` dirs on first run
- Update `ports.json` container name entries during migration

**Non-Goals:**
- Migrating existing Docker volumes (can't rename Docker volumes; old ones get cleaned up on next `asylum cleanup`)
- Migrating running containers (user restarts anyway)

## Decisions

### 1. Name format

```
asylum-<sha256(projectDir)[:12]>-<sanitized-basename>
```

Sanitization reuses existing `safeHostname` logic: lowercase, replace non-`[a-z0-9-]` with `-`, trim leading/trailing hyphens, fallback to `project`. Extract the sanitization into a shared `sanitizeProject` helper used by both `ContainerName` and `safeHostname`.

Docker container names allow `[a-zA-Z0-9_.-]` and must start with `[a-zA-Z0-9]`. Our format satisfies this (starts with `asylum`, uses only lowercase alphanumeric and hyphens).

Length: the hash part is fixed at 20 chars (`asylum-` + 12 hex + `-`). Docker has no hard limit on container name length, but we truncate the project suffix to keep names reasonable — reuse the same 56-char limit from `safeHostname` (giving a max of ~76 chars).

### 2. Migration strategy

When asylum starts for a project, before anything else that needs the project dir:

1. Compute the old-format name: `asylum-<hash[:12]>` (just the hash, no suffix)
2. Compute the new-format name: `asylum-<hash[:12]>-<project>`
3. Check if `~/.asylum/projects/<old-name>/` exists
4. If it does and `~/.asylum/projects/<new-name>/` does not: rename the directory
5. Update `ports.json`: find entries where `Container == old-name` and update to new-name

This is a one-time operation per project. After migration, the old directory no longer exists.

Export an `OldContainerName(projectDir)` function so migration code can compute both formats. This keeps the old format accessible without duplicating the hash logic.

### 3. Ports migration

Add `ports.RenameContainer(oldName, newName)` that finds the entry matching `old-name` and updates its `Container` field. Uses the existing file-locking pattern.

### 4. Volume orphaning

Old volumes (prefixed with old container name) become orphaned after migration. They won't be found by `docker.ListVolumes(newName + "-")`. This is acceptable — they'll be cleaned up by `asylum cleanup --all` which lists all `asylum-*` volumes. No special migration needed for volumes.

## Risks / Trade-offs

**Old volumes orphaned until `cleanup --all`** → Acceptable. Volumes are small and don't cause issues. Users who care about disk space can run `cleanup --all`.

**Container name gets longer** → Acceptable. Docker has no practical limit. The suffix is truncated to 56 chars.
