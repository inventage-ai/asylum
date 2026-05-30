## Why

Cache directories (`~/.gradle`, `~/.m2`, `~/.npm`, `~/.cache/pip`) became unwritable for agents after commit `d27dd09` (2026-03-23) switched them from bind mounts (which preserved host ownership) to named Docker volumes (which Docker creates as `root:root`). The existing fix for shadow `node_modules` volumes (commit `1aea0f5`, 2026-04-05) chowns them to the host UID after container start, but the same fix was never extended to cache volumes — so agents hit `EACCES` on Maven, Gradle, pip, and npm cache writes.

## What Changes

- After `RunDetached` succeeds, asylum chowns every cache volume mount point to the host `UID:GID` (mirroring the existing node_modules chown loop in `cmd/asylum/main.go:302-308`).
- Behavior is unconditional: cache volumes are always chowned when they are mounted. There is no feature flag.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `container-assembly`: Add a requirement that cache volume mount points are chowned to the host UID/GID after container start.

## Impact

- **`cmd/asylum/main.go`**: Extend the existing post-`RunDetached` chown block to also iterate `cacheDirs` and chown each mount point as root.
- **No image, config, or kit changes.** The fix is purely runtime ownership normalization.
- **No migration needed.** Existing root-owned volumes are repaired on the next container start.
