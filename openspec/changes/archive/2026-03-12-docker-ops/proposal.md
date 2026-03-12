## Why

Multiple packages need to shell out to Docker CLI for building images, inspecting labels, running containers, and pruning. A thin wrapper centralizes Docker CLI interaction.

## What Changes

- Create `internal/docker` package with functions for: build, inspect (get labels), run (assemble args), image prune, image remove
- All functions shell out to `docker` CLI via `os/exec`
- No unit tests — exercised through integration in image and container packages

## Capabilities

### New Capabilities
- `docker-cli-wrapper`: Thin wrapper around Docker CLI commands used by image and container packages

### Modified Capabilities

None.

## Impact

- Adds `internal/docker/docker.go`
- Used by `internal/image` and `internal/container`
