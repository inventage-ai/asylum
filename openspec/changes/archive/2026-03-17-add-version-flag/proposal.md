## Why

There is no way for users to check which version of Asylum they have installed. A `--version` flag is a standard CLI convention that helps with troubleshooting and verifying updates. The `version` variable already exists (set via ldflags at build time), but is not exposed to the user.

## What Changes

- Add a `--version` flag that prints the current version and exits
- Wire it into the existing flag parsing and dispatch logic

## Capabilities

### New Capabilities

_None — this is a minor addition to the existing CLI dispatch capability._

### Modified Capabilities

- `cli-dispatch`: Add `--version` flag to flag parsing and dispatch

## Impact

- `cmd/asylum/main.go`: Add version flag parsing and dispatch (before container setup)
- No new dependencies, no breaking changes
