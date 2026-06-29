## Why

Agents (Claude, Gemini, Codex, Copilot, Opencode, Pi) are currently installed via `RUN` commands in the base Docker image that always fetch the latest version. There is no pinning, so every new agent release automatically becomes the sandbox default without the user's knowledge or control. A local versions file enables automatic but deterministic version management — agents update on their own, but on a controlled cadence.

## What Changes

- A new local file `~/.asylum/versions.json` tracks the latest known version for each agent.
- On first run (when the file does not exist), a blocking fetch retrieves all agent versions before the image build proceeds.
- On subsequent runs, the file is loaded from disk (instant), the build proceeds with whatever is cached, and a background goroutine fetches updated versions every 24 hours.
- Fetch failures on background updates are silently ignored; stale versions remain valid until the next successful fetch.
- Each agent's Dockerfile `RUN` instruction gets an `ARG <AGENT>_VERSION` injected immediately before its command, and the command is modified to use that version (e.g. `@${GEMINI_VERSION}` for npm agents, `VERSION=${COPILOT_VERSION}` for curl agents).
- The ARGs are placed per-agent, not at the top of the Dockerfile, so changing one agent's version only invalidates that single RUN layer.

## Capabilities

### New Capabilities

- `agent-version`: Local version tracking file (`~/.asylum/versions.json`) with blocking first-run fetch, background refresh (24h), and per-agent ARG injection into Dockerfile agent install snippets.

### Modified Capabilities

- `agent-install`: Dockerfile snippets for npm agents modified to accept a version argument; curl-agent install scripts modified to accept a version parameter.
- `container-image`: Base image hash incorporates version ARGs so version changes trigger appropriate layer invalidation.

## Impact

- `internal/agent/install.go`: New `VersionedSnippet` field or transform logic; Dockerfile assembly passes versions to agent snippets.
- `internal/config/state.go` or new package: `versions.json` load/save with blocking and async fetch logic.
- `internal/image/image.go`: Base image hash now includes version data; version ARGs written to Dockerfile.
- `cmd/asylum/main.go`: Background version update goroutine spawned after container starts.
- New package: `internal/versions/` with fetchers for each agent source (GitHub API, npm registry).
- No config schema changes; purely additive.
