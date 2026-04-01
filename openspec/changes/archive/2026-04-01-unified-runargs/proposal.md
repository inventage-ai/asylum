## Why

Container assembly in `container.RunArgs()` is procedural and scatters kit-specific logic across `main.go` and `container.go` — four places check for specific kit names (`ports`, `docker`, `title`) and run hardcoded behavior. This design broke the ports kit: `cfg.KitActive("ports")` only checks the config map, but TierAlwaysOn kits don't appear there, so port allocation is silently skipped and containers have no port mappings. Rather than patching the check, we should fix the architecture so kits declare their own container-time behavior.

## What Changes

- Introduce a `RunArg` type that pairs a docker flag+value with a source label and priority
- Add a `ContainerFunc` field to the Kit struct — each kit produces `[]RunArg` tuples at container creation time
- Build a collection/dedup/validation pipeline that merges args from all sources (core, kits, user config) with conflict detection
- Move port allocation into the ports kit's `ContainerFunc` (self-contained)
- Move `--privileged` and `ASYLUM_DOCKER=1` into the docker kit's `ContainerFunc`
- Add `--debug` flag that prints every docker run argument with its source before launching
- **BREAKING**: Remove port release system (`ports.Release`, `ports.ReleaseContainer`) — port allocations are permanent per project
- Remove title kit (exec-time `--name` feature broken for same reason, out of scope to fix)
- Remove all `cfg.KitActive()` checks in container assembly path

## Capabilities

### New Capabilities
- `runarg-pipeline`: Unified docker run argument collection, deduplication, conflict detection, and debug output

### Modified Capabilities
- `container-assembly`: RunArgs rewritten to use the RunArg pipeline instead of procedural arg building
- `port-allocation`: Remove release/cleanup functions; allocation called from ports kit's ContainerFunc instead of main.go

## Impact

- `internal/container/container.go` — major refactor of `RunArgs()`, remove `appendPorts`, `appendEnvVars`
- `internal/kit/kit.go` — add `ContainerFunc` field and `ContainerOpts` type
- `internal/kit/ports.go` — add `ContainerFunc` that calls `ports.Allocate()` and returns `-p` args
- `internal/kit/docker.go` — add `ContainerFunc` returning `--privileged` and env var
- `internal/kit/title.go` — deleted
- `internal/ports/ports.go` — remove `Release`, `ReleaseContainer`, `release` functions
- `cmd/asylum/main.go` — remove port allocation block, remove `ReleaseContainer` callsites, add `--debug` flag handling
