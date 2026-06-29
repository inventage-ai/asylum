## Context

Agent install instructions live in `internal/agent/install.go` as static `DockerSnippet` strings (e.g. `RUN npm install -g @google/gemini-cli`). These are assembled by `image.EnsureBase` into the base Dockerfile via `AssembleAgentSnippets`. The base image hash (in `baseHash`) determines whether a rebuild is needed. Agent versions are always "latest" — there is no pinning mechanism.

## Goals / Non-Goals

**Goals:**
- Track the latest version of every agent in a local JSON file at `~/.asylum/versions.json`.
- Fetch versions on first run (blocking) so the very first build has real versions.
- Fetch versions in the background on subsequent runs (24h interval); do not block startup.
- Inject `ARG <AGENT>_VERSION` right before each agent's install command in the Dockerfile.
- Modify agent install commands to use their version ARG (npm `@${VERSION}`, curl agents' `VERSION` env or argument).
- Preserve Docker layer cache: each agent gets its own ARG so changing one agent's version only invalidates that agent's RUN layer.

**Non-Goals:**
- Kit/tool version pinning (deferred to a future change).
- Manual version override via config (user can edit versions.json directly if needed).
- Pinning to specific agents only (all agents are versioned together).
- Rollback — if a new version's Dockerfile command fails, the user must either wait for a fix or edit versions.json to a known good version.
- Retry logic — background fetch is best-effort; failures are silently ignored.

## Decisions

**Local file at `~/.asylum/versions.json`.** A simple JSON file, same location as the existing `state.json`. No new dependencies — just `encoding/json` from stdlib.

Format:
```json
{
  "claude": "2.1.195",
  "gemini": "0.8.0",
  "codex": "0.1.0",
  "copilot": "1.0.65",
  "opencode": "0.0.55",
  "pi": "0.13.0"
}
```
Values are stored exactly as returned by each source. GitHub fetchers strip the `v` prefix from tag names; npm fetchers return the version as-is (npm versions have no prefix).

**Blocking fetch only when the file does not exist.** First run: fetch all versions, write file, proceed with build. Subsequent runs: load from disk, proceed with build immediately, spawn a goroutine for background updates.

```
First run:                     Subsequent run:
  load versions.json (missing)   load versions.json (present)
  blocking fetch all             background goroutine
  write versions.json            if stale: fetch all
  continue build                 update versions.json
  continue build                 (failures ignored)
```

**Background goroutine spawned after container starts.** The version update is fire-and-forget. If it fails, the next run uses cached values. If it succeeds, the next run picks up new versions. The goroutine runs at most once per container startup — it does not loop.

**ARG injected right before each agent's RUN.** Not at the top of the Dockerfile. This preserves layer cache:

```dockerfile
# ...kit layers, fnm, mise, etc. (cached, unchanged)

ARG CLAUDE_VERSION=2.1.195
RUN curl -fsSL https://claude.ai/install.sh | bash -s -- ${CLAUDE_VERSION}
# ^ Claude's layer only — Gemini change doesn't invalidate this

ARG GEMINI_VERSION=0.8.0
RUN npm install -g @google/gemini-cli@${GEMINI_VERSION}
# ^ Gemini's layer only — Claude change doesn't invalidate this

# ... rest of Dockerfile (cached, unchanged)
```

Each agent's ARG is scoped to exactly one RUN instruction.

**Fetch sources per agent:**

| Agent | Source | API call |
|-------|--------|----------|
| Claude | GitHub tags | `GET /repos/anthropics/claude-code/tags` — first non-pre-release tag |
| Gemini | npm registry | `GET https://registry.npmjs.org/@google/gemini-cli/latest` |
| Codex | npm registry | `GET https://registry.npmjs.org/@openai/codex/latest` |
| Copilot | GitHub releases | `GET /repos/github/copilot-cli/releases/latest` |
| Opencode | GitHub releases | `GET /repos/opencode-ai/opencode/releases/latest` |
| Pi | npm registry | `GET https://registry.npmjs.org/@earendil-works/pi-coding-agent/latest` |

Each fetcher returns a plain string version. npm fetchers return the `version` field from the `/latest` response. GitHub fetchers return the `tag_name` stripped of the `v` prefix.

**Dockerfile snippet modification:** The existing `AgentInstall.DockerSnippet` is static text. A new field `VersionedSnippet` (or a transform function) receives the version map and returns the Dockerfile lines with ARG + modified RUN:

```go
// For npm agents:
// Before: RUN npm install -g @google/gemini-cli
// After:  ARG GEMINI_VERSION=0.8.0\nRUN npm install -g @google/gemini-cli@${GEMINI_VERSION}

// For Claude:
// Before: RUN curl -fsSL https://claude.ai/install.sh | bash
// After:  ARG CLAUDE_VERSION=2.1.195\nRUN curl -fsSL https://claude.ai/install.sh | bash -s -- ${CLAUDE_VERSION}

// For Copilot:
// Before: RUN curl -fsSL https://gh.io/copilot-install | bash
// After:  ARG COPILOT_VERSION=1.0.65\nRUN VERSION=${COPILOT_VERSION} curl -fsSL https://gh.io/copilot-install | bash

// For Opencode:
// Before: RUN curl -fsSL https://opencode.ai/install | bash
// After:  ARG OPENCODE_VERSION=0.0.55\nRUN curl -fsSL https://opencode.ai/install | bash -s -- --version ${OPENCODE_VERSION}
```

**VersionedSnippet function.** Instead of adding a new struct field, a transform function `AssembleVersionedAgentSnippets(installs []*AgentInstall, versions map[string]string) string` replaces `AssembleAgentSnippets`. It receives the version map and returns the combined snippet text with ARG lines injected. This keeps the change localized to the image assembly layer — `AgentInstall` structs remain unchanged.

**Base image hash includes version map.** `baseHash()` must incorporate the version map so that a version change triggers a rebuild. Since the version ARGs are part of the Dockerfile content, the existing hash approach (hashing the assembled Dockerfile) already covers this — if versioned snippets differ from non-versioned ones, the hash changes and rebuilds.

**No kit versioning.** Kit versions (mise, fnm, uv, etc.) are managed separately through kit config (`kits.python.versions`, etc.) or the core Dockerfile. This change only covers agent versions.

## Risks / Trade-offs

- **Agent install scripts may not accept version arguments reliably.** The install.sh scripts for Claude, Copilot, and Opencode all accept version parameters, but this is a contract with upstream projects. If an install script changes its CLI, the Dockerfile generation needs to adapt.
- **npm registry rate limits.** Multiple users running asylum concurrently could hit npm's rate limit. Mitigation: background fetch with 24h interval means each user fetches at most once per day, and failures are silently ignored (existing versions still work).
- **Build failure with bad version.** If a version in versions.json doesn't work (e.g. registry deleted a tag, install script broke), the Docker build will fail at that agent's RUN layer. The error is clear and actionable — the user can edit versions.json to pin to an older version.
- **Stale versions on offline users.** Users without internet access keep their cached versions. This is acceptable — the alternative (blocking and failing on first run without internet) is worse.

## Migration Plan

Purely additive. Existing installs without `versions.json` get a blocking fetch on their first run after this change. Users who already have `asylum` installed will see one build with versioned ARGs on first run, then cached builds thereafter. No config migration needed.

## Open Questions

None — all decisions resolved during exploration.
