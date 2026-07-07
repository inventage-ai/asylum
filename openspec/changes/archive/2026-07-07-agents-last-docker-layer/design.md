## Context

The base image Dockerfile is assembled from: `Dockerfile.core` → ordered kit/agent snippets → `Dockerfile.tail`. The `dockerfile-ordering` capability (archived change `2026-04-01-dockerfile-instruction-ordering`) orders those middle snippets by a static `DockerPriority` and preserves the order across builds via `state.json`'s `docker_source_order`, so that adding a kit appends a new layer at the end rather than busting the cache of existing layers.

That design placed agents first (priorities 5-6) on the premise that agent CLIs "change on asylum updates only." The `2026-06-29-agent-version-pin` mechanism broke that premise: `AgentInstall.VersionedSnippet` now bakes a pinned `ARG <AGENT>_VERSION=<v>` into each agent's snippet, resolved from `versions.VersionMap` on every run. Agents are now the highest-churn sources, yet they occupy the lowest layers — so each version bump rebuilds every kit layer above them.

Node runtime lives in `Dockerfile.core` (fnm + Node LTS), not in the `node` kit, so npm-installed agents (gemini/codex/pi) do not depend on any kit layer. Placing agents after all kits is therefore dependency-safe.

## Goals / Non-Goals

**Goals:**
- Make agent snippets the last thing before the tail, structurally, so agent version bumps only rebuild the agent block + tail.
- Keep kit ordering (priority + state preservation + append-on-add) exactly as it is today.
- Migrate existing installs automatically with a single one-time rebuild.

**Non-Goals:**
- Changing kit priorities or the kit ordering algorithm.
- Per-agent cache optimization within the agent block (a bump to one agent still rebuilds agents listed after it — acceptable; agents are cheap relative to kits).
- Any user-facing config, CLI, or project-image behavior change.

## Decisions

### Decision 1: Agents leave the tracked-source set entirely

`collectSources` stops emitting `agent:*` sources; it returns kit sources only. `computeSourceOrder` and `docker_source_order` therefore track kits exclusively. `assembleDockerfile` gains an explicit agent-snippet block written after the ordered kit snippets and before the tail (using `agent.AssembleAgentSnippets` with versioning applied).

**Why:** This makes "agents are the volatile top layer" an invariant of the assembly step, not an emergent property of priority numbers that a future kit could accidentally out-rank. It also removes agents from the machinery that was pinning them in place.

**Alternatives considered:** Flip agent priorities to a high band (55-60) and keep them in `computeSourceOrder`. Rejected: the order-preservation logic would keep existing installs' agents at the front until an unrelated removal, requiring a separate migration trigger, and nothing structurally guarantees agents stay last.

### Decision 2: Drop `DockerPriority` from `AgentInstall`

With agents no longer sorted by priority, the field is dead. Remove it from the struct and from the six agent registrations that set it.

**Why:** Less code; removes a field that would otherwise imply agents participate in priority ordering when they don't. `DockerPriority` remains on `Kit`.

### Decision 3: Agent block internal order

Agents are emitted in the existing deterministic order (claude first, then the rest sorted by name — the current `ResolveInstalls` order). No priority needed.

**Why:** Order within the block only affects intra-agent cache churn, which is negligible. Deterministic output keeps `baseHash` stable across runs.

### Decision 4: `baseHash` continues to cover agent snippets and their position

`baseHash` already hashes agent banner lines and (via ordered snippets) agent content. It must continue to hash the assembled agent block so a version bump still changes the hash and triggers a rebuild. Since the agent block moves out of `orderedIDs`, hash it explicitly from the assembled agent snippets.

## Risks / Trade-offs

- **[Risk] One-time full rebuild on upgrade** → Existing `docker_source_order` lists `agent:*` at the front. When those IDs vanish from the active source set, `computeSourceOrder` treats them as removed at index 0, re-sorting the whole kit suffix; combined with the hash change this forces one full base rebuild. This is the intended, one-time migration cost; every subsequent agent bump is cheap. No explicit migration code needed.
- **[Risk] Stale `agent:*` entries left in `docker_source_order`** → They are already filtered as unknown identifiers by the existing "unknown identifiers ignored" logic and will not be re-persisted, since only kit IDs are saved after the next successful build.
- **[Trade-off] Intra-agent-block churn** → A bump to an agent listed before others rebuilds the later agents too. Accepted: agent installs are cheap (a curl/npm install) versus the kit layers we are protecting.

## Migration Plan

No migration code. On first run after upgrade, the hash mismatch + agent removal from the tracked set produces one full base rebuild that writes a kit-only `docker_source_order`. Rollback is a binary downgrade; the old binary re-reads the same state and rebuilds once more.

## Open Questions

_None._
