## Why

The agent version cache-busting feature (`internal/versions`, `AgentInstall.VersionedSnippet`) shipped with correct core behavior but carries dead code, a sequential first-run fetch that can stall startup, and two robustness gaps that surface under concurrency and partial fetch failures. These are cheap to fix now while the code is fresh and uncovered by real usage.

## What Changes

- **Remove dead code**: delete `versions.FetchForAgent`, `agent.AssembleVersionedAgentSnippets`, and the unused `versions.AgentSource` / `SourceNpmAgent` / `SourceGitHubRelease` / `SourceGitHubTag` / `AgentSourceMap` symbols. All are referenced only by their own tests; production code uses `FetchAll` and `VersionedSnippet` directly.
- **Parallelize the blocking first-run fetch**: `FetchAll` currently queries six sources sequentially, each with a 60s client timeout, so a single slow endpoint stalls the very first `asylum` startup. Fetch concurrently with a bounded errgroup-style fan-out.
- **Make `versions.json` writes concurrency-safe**: `Write` uses a fixed `path + ".tmp"` temp file. Two concurrent `asylum` invocations (different projects, same user, one shared `~/.asylum/versions.json`) can clobber each other's temp file mid-write. Use a unique temp file (`os.CreateTemp`) before the atomic rename.
- **Retry agents missing from a partial fetch sooner**: a partial `FetchAll` (some agents errored) is still written, resetting the file mtime and suppressing any retry for 24h. Treat a version map missing tracked agents as stale so the next run re-attempts the missing ones.
- **Fix the `versions.Read` doc comment**: it claims an empty `VersionMap` is returned for a missing/corrupt file, but it returns `nil` (which `main` relies on to trigger the blocking fetch). Correct the comment to match behavior.
- **Reword the user-facing "pinning" language**: the CHANGELOG entry says versions are "pinned," but the mechanism only cache-busts the install layer on a new release — npm agents genuinely pin, while curl/script agents reinstall latest. Reword to describe cache-busting to avoid future confusion.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `agent-version`: the version-fetch refresh and write-durability requirements change — fetches run concurrently, writes are safe under concurrent invocation, and a version map missing tracked agents is considered stale so missing agents are retried before the 24h interval.

## Impact

- Code: `internal/versions/versions.go` (parallel fetch, dead-code removal), `internal/versions/file.go` (unique temp file, staleness-considers-missing-agents, doc fix), `internal/agent/install.go` (dead-code removal), and the corresponding `_test.go` files.
- Docs: `CHANGELOG.md` wording.
- No change to the Dockerfile assembly, ARG injection, or the on-disk `versions.json` format. No new dependencies (stdlib `sync`/`os` only). No config or CLI surface change.
