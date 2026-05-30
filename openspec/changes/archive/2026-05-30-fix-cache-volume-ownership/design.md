## Context

Named Docker volumes are created with `root:root` ownership inside the Docker VM. When asylum mounts them at user-owned paths inside the container (e.g. `/home/simon/.gradle`), the mount point itself flips to `root:root` and the container user (a regular UID matched to the host) cannot write to it.

This was previously masked by two things:

1. Before commit `d27dd09` (2026-03-23) cache dirs were bind mounts from `~/.asylum/cache/<container>/` on the host. Bind mounts preserve host ownership, so the directories were already user-owned.
2. For volumes mounted onto a non-empty directory in the image, Docker seeds the volume with the image contents and their ownership on first mount. The `java/gradle` kit happens to pre-create `~/.gradle` during image build as the unprivileged user (via `mise install gradle`), so the gradle cache volume was sometimes user-owned by accident. The `java/maven` kit installs Maven as root via apt and never touches `~/.m2`, so the maven cache volume came up empty and root-owned.

This accidental behavior is brittle (depends on whether the kit was enabled when the volume was first created, and on what the kit's `RUN` lines happen to leave behind in `$HOME`). The existing node_modules fix (`cmd/asylum/main.go:302-308`, added in commit `1aea0f5`) already establishes the pattern for fixing this deterministically: `docker exec --user root chown $UID:$GID <mountpoint>` after `RunDetached`.

## Goals / Non-Goals

**Goals:**
- Cache volume mount points are always writable by the container user, regardless of which kit installed which tool as which user.
- Reuse the existing `docker exec` chown pattern — no new mechanisms.
- Idempotent: safe to run on every container start, including ones where the volume was already user-owned.

**Non-Goals:**
- Recursively chown volume contents. Only the mount point itself needs fixing; contents created inside the container are already created as the container user.
- Repair existing root-owned files inside long-lived volumes. The user can purge via `asylum --cleanup` if they want a clean slate, but the new chown only touches the top-level mountpoint.
- Cache volume ownership is not a kit concern. Kits don't opt in or out.

## Decisions

### Decision: Chown after `RunDetached`, not in the entrypoint

Mirroring the existing node_modules fix. `docker exec --user root chown` runs after `WaitReady` so the container's filesystem is mounted and the entrypoint has finished; we don't need to push extra logic into `entrypoint.core`.

**Alternatives considered:**
- *Chown inside `entrypoint.core`*: would require running the entrypoint as root and then dropping privileges. The current entrypoint runs as the user. Avoiding a privilege dance keeps the entrypoint simple.
- *Pre-create the dirs in `Dockerfile.core` so the volume gets seeded with user ownership*: relies on Docker's "non-empty mount point" seeding behavior, which is the brittle thing we're trying to escape from. Volumes that existed before such a Dockerfile change would still be root-owned.

### Decision: Chown only the mount point, not contents

Volume contents created by the agent inside the container are already created as the container user. Recursively chowning every volume on every start would scale poorly (gradle caches grow large) and would also stomp on intentional root-owned files (none today, but the boundary is cleaner).

### Decision: Unconditional, no feature flag

The behavior is correct in every case (chowning an already-user-owned dir to the same uid is a no-op). A feature flag would only add config surface for no win.

## Risks / Trade-offs

- **[Risk]** `docker exec` adds ~100ms to container start. → **Mitigation**: only runs on fresh container starts (gated by `freshContainer` block), not on every `asylum` invocation. Same gating the node_modules chown already uses.
- **[Risk]** A long-lived volume could have root-owned files *inside* it that the chown won't touch, causing partial failures (mountpoint writable, subdir not). → **Mitigation**: documented in `--cleanup`; in practice the container user already has write on its own caches once the mountpoint is fixed, since tools recreate the subdirs they need.
- **[Trade-off]** We don't migrate or rewrite existing data. Acceptable: cache contents are reproducible by re-downloading.
