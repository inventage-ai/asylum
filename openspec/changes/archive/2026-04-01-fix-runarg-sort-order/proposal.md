## Why

The `ResolveArgs` sort produces docker args where `run` and `-d` appear after option flags like `--cap-add`. Docker receives `docker --cap-add SYS_ADMIN ... run -d` instead of `docker run -d ...`, failing with "unknown flag: --cap-add". This blocks container startup entirely.

## What Changes

- Remove `run` and `-d` from the RunArg pipeline — they are the docker subcommand, not options subject to dedup or override
- Prepend `["run", "-d"]` as a fixed prefix when flattening resolved args to `[]string`
- Clean up `booleanFlags` map (remove `-d` since it no longer goes through dedup)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `runarg-pipeline`: The pipeline no longer includes `run`/`-d` as RunArg entries; they are prepended as a structural prefix during flattening
- `container-assembly`: `RunArgs()` stops emitting `run`/`-d` as core RunArgs

## Impact

- `internal/container/container.go` — RunArgs function, FlattenArgs or flat slice assembly
- `internal/container/runarg.go` — `booleanFlags` map
- `internal/container/runarg_test.go` — tests referencing `run`/`-d` RunArgs
- `internal/container/container_test.go` — tests checking flattened arg output
