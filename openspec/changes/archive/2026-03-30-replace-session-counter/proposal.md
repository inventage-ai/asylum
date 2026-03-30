## Why

The file-based session counter corrupts when asylum is killed by an unhandled signal (SIGHUP on terminal close, SIGKILL, or any crash). Once corrupted, a ghost count permanently prevents container cleanup on every subsequent normal exit — the counter increments and decrements by 1 each session, but the stale value never clears. The only recovery is `asylum cleanup`.

## What Changes

- Remove the file-based session counter system (`IncrementSessions`, `DecrementSessions`, `adjustCounter`, `SessionCount`, `sessionCounterPath`, and the `sessions` file)
- Add a runtime check that queries active exec sessions inside the container via `ps` at cleanup time — no persistent state to corrupt
- Add SIGHUP to the signal handler in `runDocker` for clean signal forwarding on terminal close
- Simplify the rebuild prompt to remove session count display

## Capabilities

### New Capabilities

None — this is a mechanism replacement, not a new capability.

### Modified Capabilities

- `container-exec`: The "Container cleanup after last session" requirement changes from counter-based to runtime-based session detection

## Impact

- `internal/container/container.go`: Remove ~80 lines of counter logic
- `internal/docker/docker.go`: Add `HasOtherSessions` function
- `cmd/asylum/main.go`: Simplify cleanup flow (remove increment/decrement, replace with runtime check), add SIGHUP to signal handler, simplify rebuild prompt
