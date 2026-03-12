## Why

Asylum's behavior is driven by layered YAML config files. The config system is a prerequisite for all runtime features — agent selection, volume mounts, port forwarding, and package management all depend on it.

## What Changes

- Create `internal/config` package with YAML config parsing
- Implement three-layer merge: `~/.asylum/config.yaml` → `$project/.asylum` → `$project/.asylum.local`
- Merge semantics: scalars last-wins, lists concatenated, maps-of-lists concatenated per sub-key
- Volume shorthand parsing with `~` expansion and `:ro`/`:rw` support
- Unit tests for merge logic and volume parsing

## Capabilities

### New Capabilities
- `config-loading`: Layered YAML config loading and merging with CLI flag overlay
- `volume-shorthand`: Volume path parsing with shorthand expansion and tilde resolution

### Modified Capabilities

None.

## Impact

- Adds `internal/config/config.go` and `internal/config/config_test.go`
- Used by container, image, and CLI packages in subsequent changes
