## 1. Remove dead code

- [x] 1.1 Delete `FetchForAgent` from `internal/versions/versions.go` and its test `TestFetchForAgent` in `versions_test.go`
- [x] 1.2 Delete the `AgentSource` type, the `SourceNpmAgent`/`SourceGitHubRelease`/`SourceGitHubTag` constants, and `AgentSourceMap` from `internal/versions/versions.go`
- [x] 1.3 Delete `AssembleVersionedAgentSnippets` from `internal/agent/install.go` and its test `TestAssembleVersionedAgentSnippets` in `install_version_test.go`
- [x] 1.4 Run `go build ./...` and `go vet ./...` to confirm no remaining references

## 2. Parallelize the version fetch

- [x] 2.1 Rewrite `FetchAll` in `internal/versions/versions.go` to run all fetchers concurrently with `sync.WaitGroup`, collecting results into the `VersionMap` under a `sync.Mutex`; keep failed fetches omitted
- [x] 2.2 Confirm existing `FetchAll`/fetcher tests still pass (assertions are on map contents, not order)

## 3. Concurrency-safe writes

- [x] 3.1 Change `Write` in `internal/versions/file.go` to use `os.CreateTemp(dir, "versions-*.json")` for the temp file, then atomic `os.Rename`; remove the temp file on any pre-rename error
- [x] 3.2 Verify the existing `Write`/round-trip test passes and add a check that a failed write leaves an existing file intact

## 4. Retry missing agents before the interval

- [x] 4.1 Add a staleness check (extend `IsStale` or add a sibling helper) that treats the version map as stale when it is missing an entry for any tracked agent (`fetchers` keys via `AgentNames`), in addition to the 24h mtime rule
- [x] 4.2 Update the staleness branch in `cmd/asylum/main.go` to use the extended check so a partial prior fetch is retried on the next run
- [x] 4.3 Add a unit test covering: file fresh + all agents present → not stale; file fresh + missing agent → stale; file old → stale

## 5. Documentation fixes

- [x] 5.1 Fix the `Read` doc comment in `internal/versions/file.go` to state it returns `nil` (not an empty map) for a missing or corrupt file
- [x] 5.2 Reword the `CHANGELOG.md` "Unreleased" entry to describe cache-busting the install layer on a new release rather than "pinning"

## 6. Verification

- [x] 6.1 Run `go test ./internal/versions/... ./internal/agent/... ./internal/image/...`
- [x] 6.2 Run `go build ./...` and `go vet ./...`
