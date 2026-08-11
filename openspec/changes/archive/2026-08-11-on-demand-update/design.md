## Context

Agent versions are resolved from upstream (npm/GitHub) into `~/.asylum/versions.json` and baked into the base image as build ARGs. On startup:

- Missing/corrupt `versions.json` → blocking `versions.FetchAll()` then write (`main.go:312-319`).
- Otherwise → `versions.NeedsRefresh(path, vm, time.Hour)`; if stale, a background goroutine fetches and rewrites the file (`main.go:320-332`).

The staleness argument is hardcoded to `time.Hour`, even though the `agent-version` spec already specifies a 24-hour window. `NeedsRefresh`/`IsStale` already take the interval as a `time.Duration` parameter, so the interval is fully externalised at the `versions` layer — only the caller hardcodes it.

Subcommands `version`, `cleanup`, and `config` are dispatched early and return before container setup (`main.go:57-67`); `self-update` is dispatched after config load (`main.go:116-136`). Image assembly for a normal run happens in `ensureImages(...)` (`main.go:335`, defined at `main.go:1117`), which calls `image.EnsureBase` then `image.EnsureProject`.

## Goals / Non-Goals

**Goals:**
- Default background-refresh window is 24h, matching the spec.
- Window is configurable via a single top-level config field, resolved through the existing layered merge.
- `asylum update` force-refreshes versions and rebuilds images only if the fetched versions changed the image hash, then exits without starting a container.

**Non-Goals:**
- No change to how versions are fetched, written, or injected into the Dockerfile.
- `asylum update` does not update the asylum binary (that stays with `self-update`).
- No new flags on `update` (no `--dev`, no target version); it is argument-less.

## Decisions

### Config field: `version-check-interval`, Go duration string, default 24h
Add `VersionCheckInterval string` to `config.Config` (yaml `version-check-interval,omitempty`) plus a merge rule mirroring `ReleaseChannel` (`config.go:431` — overlay wins when non-empty). Add a resolver method:

```go
func (c Config) VersionCheckStaleness() time.Duration {
    if d, err := time.ParseDuration(c.VersionCheckInterval); err == nil && d > 0 {
        return d
    }
    return 24 * time.Hour
}
```

Rationale: a duration string (`24h`, `168h`, `1h`) is idiomatic Go and needs no unit convention doc. Invalid/empty/zero → 24h default, so a typo never breaks a run. Alternative (integer hours) was rejected as less expressive and requiring its own unit comment. `main.go:320` changes `time.Hour` → `cfg.VersionCheckStaleness()`.

### `asylum update` dispatch placement
`update` needs a fully-loaded config, assembled kits, agent installs, and state — the same inputs `ensureImages` already receives — so it cannot be an early-return dispatch like `version`/`config`. Instead, reuse the normal startup path up to and including `ensureImages`, but branch on `subcommand == "update"` to (a) force the fetch and (b) return immediately after `ensureImages`, before container start.

Concretely:
- Parsing: recognise `update` as a subcommand (`parseArgs`, near the `self-update` case ~`main.go:665`).
- The version block (`main.go:305-332`): when `subcommand == "update"`, replace the load/staleness logic with an unconditional blocking `versions.FetchAll()` + write (reusing the existing missing-file branch's fetch-and-write code), so the subsequent `ensureImages` sees the fresh version map.
- After `ensureImages` (and `SaveState`), if `subcommand == "update"`, log a short summary and `return` before `containerRunning`/container-start logic.

Rationale: rebuild-if-needed falls out for free — `ensureImages` → `EnsureBase` already rebuilds only when the assembled Dockerfile hash (which includes versions) changes. No new rebuild logic. Alternative (a standalone `runUpdate()` early-dispatch like `runConfig`) was rejected because it would duplicate config load, kit assembly, agent-install resolution, and the `ensureImages` call.

### Force-fetch semantics
"Force" means bypass `NeedsRefresh` and always `FetchAll()` synchronously (blocking, not the background goroutine) so images are rebuilt within the same invocation. If `FetchAll()` returns an empty map (all sources failed), skip the write (leaving `versions.json` intact) and let `ensureImages` run against the existing map — a no-op rebuild.

### Default config documentation
Add a commented `version-check-interval` line to `configHeader` in `defaults.go`, near `release-channel`, e.g.:

```
# How often to refresh cached agent versions in the background (Go duration).
# version-check-interval: 24h
```

## Risks / Trade-offs

- [Force-fetch is blocking, unlike the normal background refresh] → Acceptable: `update` is an explicit, interactive command where the user expects to wait; `FetchAll` is already concurrency-bounded by the slowest source.
- [`update` runs the full startup path (kit sync prompt, config writes, resume-migration prompt)] → Mitigation: gate those interactive/first-run steps that already special-case non-container subcommands the same way (e.g. the kit-sync block already skips for `self-update`); confirm `update` slots in without triggering a container. Keep `update` behind the same early guards used by other non-container subcommands where they exist.
- [Naming overlap between `update` and `self-update`] → Mitigation: usage text and CHANGELOG clearly state `update` = agent versions + image, `self-update` = the asylum binary.

## Migration Plan

Additive and backward compatible: no config migration needed (`version-check-interval` is optional, absent → 24h). Existing `versions.json` files are unaffected. Rollback is reverting the binary.

## Open Questions

None.
