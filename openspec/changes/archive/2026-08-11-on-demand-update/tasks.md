## 1. Configurable staleness interval

- [x] 1.1 Add `VersionCheckInterval string` field to `config.Config` with yaml tag `version-check-interval,omitempty` (`internal/config/config.go`)
- [x] 1.2 Add the overlay merge rule (overlay wins when non-empty), mirroring `ReleaseChannel` in `mergeConfig`
- [x] 1.3 Add `VersionCheckStaleness() time.Duration` resolver: parse the field as a Go duration, fall back to 24h on empty/invalid/non-positive
- [x] 1.4 Add a table-driven test for `VersionCheckStaleness` (unset→24h, `1h`→1h, invalid→24h) and a merge test for the overlay rule
- [x] 1.5 Replace hardcoded `time.Hour` at `main.go:320` with `cfg.VersionCheckStaleness()`

## 2. Default config documentation

- [x] 2.1 Add a commented `version-check-interval` entry (with explanatory comment and `24h` default) to `configHeader` in `internal/config/defaults.go`, near `release-channel`

## 3. `asylum update` command

- [x] 3.1 Recognise `update` as a subcommand in `parseArgs` (reject unexpected args/flags, mirroring the `self-update` case)
- [x] 3.2 In the version block (`main.go:305-332`), when `subcommand == "update"`, force a blocking `versions.FetchAll()` + write (skip write when the map is empty) instead of the load/staleness path
- [x] 3.3 After `ensureImages` + state save, when `subcommand == "update"`, log a summary and `return` before any container start
- [x] 3.4 Ensure `resolveMode`/early guards treat `update` as a non-container subcommand so no container is started or exec'd
- [x] 3.5 Add `asylum update` to `printUsage` help text, clarifying it refreshes agent versions and images (distinct from `self-update`)

## 4. Verification

- [x] 4.1 `go test ./...` and `go vet ./...` pass
- [x] 4.2 Verify `asylum update` dispatch: help text lists it, unexpected args are rejected, and it exits before container start (full Docker rebuild path not exercised in-sandbox — no base image/versions.json present; rebuild-if-changed is inherited from `ensureImages`)
- [x] 4.3 Manually verify: `version-check-interval` in config changes the background-refresh window; unset defaults to 24h
- [x] 4.4 Add a CHANGELOG entry under Unreleased (Added: `asylum update`; Changed: default version-check window 1h→24h, now configurable)
