## Why

Most modern browsers block access to ports at or above 10000 by default (they're on the "restricted ports" list), which breaks the primary use case of forwarded ports: letting the user open a dev server running in the container from the host browser. Moving the default base port below this restricted range makes forwarded services work out of the box.

## What Changes

- Lower the default base port for automatic allocation from `10000` to `7001`.
- When allocating ports for a project whose existing allocation starts at or above `10000`, discard the stale range and assign a new one from the lower range. The old entry is removed so its ports become available again.
- Update the `port-allocation` capability's scenarios to reflect the new base port and the reassignment behavior.
- Update in-container reference material / sandbox rules examples that mention the `10000+` range.

No user-facing config changes; the `count` option and ports kit API are unchanged.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `port-allocation`: base port changes from 10000 to 7001; add reassignment rule for projects with pre-existing ranges ≥ 10000.

## Impact

- `internal/ports/ports.go` — `BasePort` constant and `Allocate` logic (detect + drop stale high ranges before reusing).
- `internal/ports/ports_test.go` — update existing tests and add coverage for reassignment.
- Documentation / reference: `assets/asylum-reference.md` and any README/docs that mention port `10000+` as the forwarded range.
- Existing users: on next session, projects currently holding a ≥10000 range will be migrated to a new range. Any hardcoded URLs users bookmarked against the old ports will need to be updated — acceptable, since those URLs don't work in browsers anyway.
