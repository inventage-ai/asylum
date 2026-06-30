## Context

The `internal/versions` package and `AgentInstall.VersionedSnippet` (in `internal/agent/install.go`) implement agent version cache-busting: on first run versions are fetched (blocking), persisted to `~/.asylum/versions.json`, and injected as per-agent `ARG <AGENT>_VERSION` declarations so a new upstream release changes the Dockerfile content and rebuilds only that agent's layer. The core mechanism is correct and tested. This change is a follow-up hardening pass — no behavioral redesign — addressing dead code, a sequential first-run fetch, and two robustness gaps found in review.

Current state of the relevant code:
- `versions.FetchAll` loops over six fetchers sequentially; `httpClient` has a 60s timeout.
- `versions.Write` writes to a fixed `path + ".tmp"` then renames.
- `versions.IsStale` returns true only when the file mtime is older than the passed duration.
- `versions.FetchForAgent`, `agent.AssembleVersionedAgentSnippets`, and the `AgentSource`/`AgentSourceMap` enum exist but are referenced only by tests.

## Goals / Non-Goals

**Goals:**
- Remove dead exported symbols and their tests.
- Bound first-run blocking time by parallelizing fetches.
- Eliminate the concurrent-write corruption window on the shared version file.
- Retry agents that failed a previous partial fetch before the 24h interval elapses.
- Align the `Read` doc comment and the CHANGELOG wording with actual behavior.

**Non-Goals:**
- Changing the `versions.json` on-disk format, the ARG-injection mechanism, or the per-agent install snippets.
- Adding retry/backoff logic to individual fetchers (failures remain best-effort).
- Adding new dependencies — stdlib `sync`/`os` only.
- Fixing or reinterpreting the copilot `VERSION=` env placement: it correctly busts the cache (the RUN references the ARG), which is the feature's actual goal.

## Decisions

**Parallelize `FetchAll` with `sync.WaitGroup` + a mutex-guarded map.** Each of the six fetchers runs in its own goroutine; results are collected into a `VersionMap` under a `sync.Mutex`. No new dependency (`golang.org/x/sync/errgroup` is unnecessary — there is no error to propagate, failures are dropped). Total wall-clock for the blocking first run drops from up to 6×timeout to ~1×timeout. Alternative considered: a worker pool with bounded concurrency — rejected as overkill for six fixed sources.

**Unique temp file via `os.CreateTemp`.** `Write` creates a temp file with `os.CreateTemp(dir, "versions-*.json")`, writes, `Close`s, then `os.Rename`s onto the target. This removes the fixed-name collision between concurrent invocations while keeping the atomic-rename guarantee. On any error before rename, the temp file is removed and the existing `versions.json` is untouched. Alternative considered: file locking (`flock`) — rejected as heavier and unnecessary, since rename is already atomic and last-writer-wins is acceptable for a cache.

**Staleness considers missing tracked agents.** Add a helper that reports stale when the loaded map lacks an entry for any name in the tracked-agent set (`fetchers` keys), independent of mtime. `main.go` already branches on `IsStale`; it will call the extended check so a partial prior fetch is retried on the next run rather than waiting 24h. The 24h mtime rule is preserved for the all-present case. Alternative considered: per-agent timestamps — rejected as more state than the problem warrants.

**Dead-code removal.** Delete `FetchForAgent`, `AssembleVersionedAgentSnippets`, `AgentSource`/`SourceNpmAgent`/`SourceGitHubRelease`/`SourceGitHubTag`/`AgentSourceMap`, and their tests. `go build ./...` and `go vet` confirm no production references. `AgentNames` is retained only if still used; otherwise removed too.

**Doc/wording fixes.** Correct the `Read` comment to state it returns `nil` for a missing/corrupt file (the value `main` keys on to trigger a blocking fetch). Reword the CHANGELOG entry from "pins it in the Dockerfile" to describe cache-busting on a new release.

## Risks / Trade-offs

- [Parallel fetch changes test timing/order] → Fetchers are independent and the result map is order-independent; existing httptest-based tests assert on map contents, not order. Add a concurrency-safety check only if cheap.
- [Missing-agent staleness could cause a fetch every run when an agent is permanently unfetchable (e.g. repo renamed)] → Acceptable: the background fetch is fire-and-forget and silently ignored on failure; it adds one best-effort network attempt per run, not a build stall. If this proves noisy, a future change can add a per-agent cooldown.
- [Removing exported symbols is technically an API change] → The package is internal (`internal/...`), so there are no external importers; safe to remove.

## Migration Plan

Purely internal refactor. No `versions.json` format change, no config migration, no rebuild semantics change. Existing files remain valid and are read unchanged. Rollback is a straight revert of the commit.

## Open Questions

None.
