## Why

Dockerfile instructions are currently ordered alphabetically by kit name, which means adding or removing a kit can invalidate Docker's layer cache for unrelated kits. A deliberate ordering — placing stable, expensive layers first and volatile layers last — would significantly reduce rebuild times. Additionally, newly added kits should always go last (most likely to change) while preserving the order of existing kits to avoid unnecessary cache busts.

## What Changes

- Add a static priority system to Dockerfile snippet sources (core, agents, kits) so that each source has a well-defined position based on expected build cost and change frequency.
- Track the previous build's snippet source order in `state.json` so that existing sources retain their position and only new sources are appended at the end.
- When a source is removed, re-sort the sources that followed it according to static priority (reclaiming optimal order for the suffix without disturbing the prefix).
- Apply ordering to both base image and project image Dockerfile assembly.
- Entrypoint assembly is not affected (entrypoint runs every container start, not cached by Docker layers).

## Capabilities

### New Capabilities
- `dockerfile-ordering`: Static priority assignment for Dockerfile snippet sources and state-aware ordering that preserves layer cache across kit additions/removals.

### Modified Capabilities

## Impact

- `internal/image/image.go` — `assembleDockerfile` gains ordering logic instead of concatenating in resolution order.
- `internal/kit/kit.go` — Kits gain a static `DockerPriority` field (or similar) used for tie-breaking and initial ordering.
- `internal/agent/install.go` — Agent installs gain a priority field for ordering relative to kits.
- `internal/config/state.go` — `State` struct extended with a `DockerfileSourceOrder` (or similar) field tracking previous build's source order.
- `internal/image/hash.go` (or equivalent) — Image hash computation must remain content-based so reordering alone triggers a rebuild when layer order actually changes.
- No new dependencies. No breaking changes to user-facing config.
