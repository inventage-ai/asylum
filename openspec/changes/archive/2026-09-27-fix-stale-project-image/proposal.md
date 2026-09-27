# Proposal

## Why

A project image is rebuilt only when its own Dockerfile changes or when the same invocation rebuilt the base image. The base image is shared, so when one project's run rebuilds it, every other project with a project image keeps running on the old base. Kit changes, agent version bumps, and Asylum upgrades then silently never reach those projects. Enabling the global `dropshare` kit exposed this: the first project got the new command, and a second project's recreated container did not.

## What Changes

- The project image records the hash of the base image it was built on, and is rebuilt when the current base hash differs.
- The `baseRebuilt` flag passed from `EnsureBase` to `EnsureProject` goes away, because the recorded base hash covers it.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `image-build`: A project image is current only if it was built on the current base image.
- `stale-container-detection`: A base image change reaches every project's image, not only the project whose run rebuilt the base.

## Impact

- `internal/image/image.go`: `EnsureProject` reads the base image's `asylum.hash` label and stores it as `asylum.base.hash` on the project image. Its `baseRebuilt` parameter is removed.
- `cmd/asylum/main.go`: the call site drops the argument.
- Each project with a project image rebuilds it once on its first run after this change.
- `CLAUDE.md`: the "Two-tier images" note describes the new behaviour.
