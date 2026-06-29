## 1. Version package

- [x] 1.1 Create `internal/versions/` package with a `VersionMap` type (`map[string]string`) and fetchers for each agent
- [x] 1.2 Implement npm fetcher: `fetchNpmVersion(packageName string) (string, error)` using `GET https://registry.npmjs.org/<pkg>/latest`
- [x] 1.3 Implement GitHub release fetcher: `fetchGitHubRelease(repo string) (string, error)` using `GET /repos/<repo>/releases/latest`, stripping the `v` prefix
- [x] 1.4 Implement GitHub tags fetcher: `fetchGitHubTags(repo string) (string, error)` using `GET /repos/<repo>/tags`, returning the first non-pre-release tag, stripping the `v` prefix
- [x] 1.5 Implement agent-specific fetchers that dispatch to the right source:
  - Claude → GitHub tags for `anthropics/claude-code`
  - Gemini → npm for `@google/gemini-cli`
  - Codex → npm for `@openai/codex`
  - Copilot → GitHub releases for `github/copilot-cli`
  - Opencode → GitHub releases for `opencode-ai/opencode`
  - Pi → npm for `@earendil-works/pi-coding-agent`
- [x] 1.6 Implement `FetchAll() (VersionMap, error)` that calls all agent fetchers and returns the map
- [x] 1.7 Implement `FetchForAgent(agent string, installs []*agent.AgentInstall) (VersionMap, error)` — only fetch agents that are in the active install list
- [x] 1.8 Add unit tests for npm fetcher (mock HTTP), GitHub fetcher (mock HTTP), tags fetcher filtering pre-releases

## 2. Version file I/O

- [x] 2.1 In `internal/versions/`: `Read(path string) (VersionMap, error)` — reads and parses `versions.json`; returns empty map if missing or corrupt
- [x] 2.2 In `internal/versions/`: `Write(path string, vm VersionMap) error` — marshals to JSON, writes atomically (write to temp, rename)
- [x] 2.3 Add unit tests: round-trip, missing file returns empty map, corrupt file returns empty map, file written atomically

## 3. Dockerfile snippet versioning

- [x] 3.1 In `internal/agent/install.go`: Add `VersionedSnippet(versions map[string]string) string` method to `AgentInstall` struct
  - For npm agents: inject `ARG <UPPER_NAME>_VERSION=<ver>\nRUN ...@${<UPPER_NAME>_VERSION}`
  - For Claude: inject `ARG CLAUDE_VERSION=<ver>\nRUN ...bash -s -- ${CLAUDE_VERSION}`
  - For Copilot: inject `ARG COPILOT_VERSION=<ver>\nRUN VERSION=${COPILOT_VERSION} curl...`
  - For Opencode: inject `ARG OPENCODE_VERSION=<ver>\nRUN ...bash -s -- --version ${OPENCODE_VERSION}`
  - For agents without version support (Echo): return the original snippet unchanged
- [x] 3.2 In `internal/agent/install.go`: Add `AssembleVersionedAgentSnippets(installs []*AgentInstall, versions VersionMap) string`
  - Calls `VersionedSnippet` on each install and concatenates
- [x] 3.3 In `internal/image/image.go`: Replace calls to `AssembleAgentSnippets` with `AssembleVersionedAgentSnippets` passing the version map
- [x] 3.4 Add unit tests for versioned snippet generation for each agent type

## 4. Base image assembly integration

- [x] 4.1 In `internal/image/image.go`: Update `EnsureBase` signature to accept a `VersionMap` parameter
- [x] 4.2 In `internal/image/image.go`: Pass the version map to `AssembleVersionedAgentSnippets`
- [x] 4.3 The base hash already hashes the assembled Dockerfile content — no changes needed since version ARGs are part of the Dockerfile
- [x] 4.4 In `cmd/asylum/main.go`: Pass the version map to `image.EnsureBase`

## 5. CLI integration and background update

- [x] 5.1 In `cmd/asylum/main.go`: After resolving agents and kits but before `image.EnsureBase`, load versions.json
  - If file missing/corrupt: call `versions.FetchAll()`, save, proceed (blocking)
  - If file valid: proceed immediately, spawn goroutine for background update
- [x] 5.2 Background goroutine: check `time.Now() - modTime(versions.json) > 24h`, if true call `versions.FetchAll()` and save
- [x] 5.3 Log background fetch failures at warn level (non-blocking)
- [x] 5.4 The version map flows through to `image.EnsureBase` so the Dockerfile gets versioned ARGs

## 6. Verify

- [x] 6.1 `go test ./...` and `go vet ./...` pass
- [ ] 6.2 Manual test: first run blocks on version fetch, generates Dockerfile with ARG lines
- [ ] 6.3 Manual test: second run loads cached versions, background fetch fires, build reuses cache
- [ ] 6.4 Manual test: change version in versions.json → base rebuilds
- [x] 6.5 Add a CHANGELOG.md entry under Unreleased (Added)
