## Why

The background version refresh currently fires whenever `versions.json` is older than one hour (`main.go` passes `time.Hour`), which hits the npm/GitHub APIs far more often than agent releases warrant and contradicts the `agent-version` spec, which already states a 24-hour staleness window. At the same time there is no way to force a version refresh and image rebuild on demand — users who want the newest agent CLI must wait for the staleness window to elapse or delete `versions.json` by hand.

## What Changes

- Change the default background-refresh staleness window from 1 hour to 24 hours, aligning the code with the existing `agent-version` spec.
- Make the staleness window configurable via a new top-level config field (`version-check-interval`, a Go duration string; default `24h`), resolved through the normal layered config merge.
- Add an `asylum update` subcommand that force-fetches all agent versions (bypassing the staleness check), writes `versions.json`, then rebuilds the base and project images if the new versions changed the image hash — without starting or exec'ing into a container.
- Document the new config field in the default config template.

## Capabilities

### New Capabilities
- `update-command`: The `asylum update` subcommand that performs an on-demand agent-version fetch followed by a conditional image rebuild.

### Modified Capabilities
- `agent-version`: The background-refresh staleness window becomes 24 hours by default and is configurable, rather than a hardcoded interval.
- `cli-dispatch`: A new `update` subcommand is recognised and dispatched.
- `config-defaults`: The default config template documents the new `version-check-interval` field.

## Impact

- `cmd/asylum/main.go`: replace the hardcoded `time.Hour` staleness argument with the configured interval; add `update` to argument parsing, dispatch, and usage text; add the on-demand update flow (force-fetch + `ensureImages`).
- `internal/config/config.go`: new `VersionCheckInterval` field, its merge rule, and a resolver returning a `time.Duration` with a 24h default.
- `internal/config/defaults.go`: document the field in the default config template.
- `internal/versions`: unchanged logic; `NeedsRefresh`/`IsStale` already accept the interval as a parameter.
- `CHANGELOG.md`: new Added/Changed entries.
