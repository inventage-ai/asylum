## Why

Docker layers cache linearly: a change to any layer invalidates every layer below it. The base-image ordering system (`dockerfile-ordering`) was built on the assumption that agent CLIs are the most stable sources, so it assigns them the lowest priorities (5-6) and places them first — at the bottom of the layer stack. The new agent version-pinning mechanism (`VersionedSnippet` injecting `ARG <AGENT>_VERSION`, fed by `versions.VersionMap`) inverts that assumption: agents are now the *most* frequently-changing sources. Because they sit at the bottom, every routine agent version bump invalidates and rebuilds all of the expensive, rarely-changing kit layers above them (java, python, node, …).

## What Changes

- Remove agent installs from the state-tracked ordering algorithm (`computeSourceOrder` / `docker_source_order`). Ordering state now tracks kits only.
- Emit agent Dockerfile snippets as a contiguous block appended **after all kit snippets and before the tail**, making "agents are the volatile top layer" a structural invariant rather than a priority-number coincidence.
- Agent version bumps now rebuild only the agent block plus the cheap tail; kit layers stay cached.
- Migration is automatic: existing `state.json` files list `agent:*` IDs at the front of `docker_source_order`. Once agents leave the tracked source set, those IDs are seen as removed from the front, triggering a one-time re-sort of the kit suffix and a single full rebuild — stable thereafter.
- Remove the now-vestigial "Project image agents before kits" requirement (project images never install agents — `generateProjectDockerfile` takes no agents).
- Drop the `DockerPriority` field from `AgentInstall` (agents no longer participate in priority ordering); keep it on kits.
- Update ordering tests to assert agents-after-kits (replacing `TestOrderingAgentsBeforeKits`, `TestOrderingClaudeBeforeOtherAgents`, and the `claudeIdx < javaIdx` assertion).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `dockerfile-ordering`: agents are no longer part of the priority/state-tracked ordering; they are always appended after kit snippets and before the tail. The `DockerPriority`-on-agents and "project image agents before kits" requirements are removed.

## Impact

- Code: `internal/image/order.go` (drop agents from `collectSources`/`computeSourceOrder`), `internal/image/image.go` (`assembleDockerfile`, `baseHash`, `EnsureBase`), `internal/agent/install.go` (remove `DockerPriority`), agent files that set it (`claude.go`, `codex.go`, `gemini.go`, `copilot.go`, `opencode.go`, `pi.go`), and `internal/image/order_test.go` / `assembly_test.go`.
- Behavior: one-time full base-image rebuild on first run after upgrade; subsequent agent version bumps are cheap.
- No config, CLI, or user-facing surface changes. No dependency changes.
