## Context

The base image Dockerfile is assembled by concatenating snippets from multiple sources: core template, kit DockerSnippets, agent DockerSnippets, and a tail template. Currently, kits are ordered alphabetically by name (via `kit.Resolve` → `All()` → `slices.Sort`), and agents come after all kits. This means:

1. Adding a new kit (e.g., `rust`) that sorts alphabetically before existing kits (e.g., `ssh`) invalidates Docker layer cache for all subsequent layers — even though those layers haven't changed.
2. There's no consideration of build cost or change frequency — a heavy, stable layer (like Java) can end up after a lightweight, frequently-changing one.

The project already tracks known kits in `state.json` (`KnownKits` field) for the kit sync flow. We can extend this state to remember the previous Dockerfile source order.

## Goals / Non-Goals

**Goals:**
- Minimize Docker layer cache invalidation when kits are added or removed
- Order snippets by build cost (expensive first) and change frequency (stable first) using static priorities
- Ensure newly added snippet sources go last in the Dockerfile (most likely to change, least cached value)
- When a source is removed, re-sort only the trailing sources that came after it (preserve cached prefix)
- Keep the system simple — priorities are hardcoded per kit/agent, not user-configurable

**Non-Goals:**
- Ordering entrypoint snippets (entrypoint runs every container start, not cached by Docker)
- Full state-tracked ordering for project images (project images are small and rebuild frequently) — but agents before kits ordering still applies
- User-configurable priority values
- Merging or splitting Docker RUN instructions across kits (each kit's snippet remains a single block)

## Decisions

### Decision 1: Unified source identifiers

Each Dockerfile snippet source gets a string identifier used for tracking order in state:
- `"kit:<name>"` for kits (e.g., `"kit:java"`, `"kit:node"`)
- `"agent:<name>"` for agents (e.g., `"agent:claude"`, `"agent:codex"`)

**Why**: A unified namespace avoids collisions and makes the state format self-documenting. The prefix also lets us distinguish source types when needed.

**Alternatives considered**: Using separate `kit_order` and `agent_order` fields in state — rejected because the ordering algorithm is the same for both and they need to be interleaved relative to each other.

### Decision 2: Static priority field on Kit and AgentInstall

Add a `DockerPriority int` field to both `Kit` and `AgentInstall`. Lower values = earlier in the Dockerfile (more stable / more expensive). Default priority is 50 (middle). Priorities are set at registration time.

Suggested priority ranges:
- **5-9**: Agent CLIs (Claude, Codex, Gemini — installed once, change on asylum updates only)
- **10-19**: Heavy language runtimes that rarely change (Java, Python, Node.js)
- **30-39**: Medium-weight tools (Docker-in-Docker, shell extensions, browser)
- **40-49**: Lightweight tooling kits (GitHub CLI, cx, ast-grep, openspec)
- **50**: Default — unspecified kits land here

**Why**: A single integer is simple to reason about and set. The ranges give room for future kits without re-numbering.

**Alternatives considered**: 
- String-based tiers ("heavy"/"medium"/"light") — too coarse, can't distinguish within a tier.
- Float priorities — unnecessary precision, harder to read.

### Decision 3: Order algorithm with state preservation

The ordering algorithm for `assembleDockerfile`:

1. Collect all active sources (resolved kits + resolved agents) as `(identifier, priority, snippet)` tuples.
2. Load the previous source order from state (`DockerSourceOrder []string`).
3. Partition sources into:
   - **Retained**: sources present in both current and previous order
   - **New**: sources in current but not in previous order
4. Start with retained sources in their previous order.
5. If any previously-present source was **removed**: find the earliest position of any removed source. Everything from that position onward (in the retained list) gets re-sorted by priority (since the cache is already busted at that point).
6. Append new sources at the end, sorted by priority among themselves.
7. Save the resulting order to state.

**Why**: This preserves Docker layer cache maximally. Sources that were in the same relative order before keep their positions. Only the "tail" after a removal (which Docker would rebuild anyway) gets optimized. New sources go last because they're the least proven and most likely to change.

**Alternatives considered**:
- Always sort by priority (ignore previous order) — maximally "optimal" statically but busts cache on every kit addition/removal.
- Append-only (never re-sort) — misses the opportunity to optimize after a removal when the cache is already busted.

### Decision 4: State extension

Extend `config.State` with:

```go
type State struct {
    KnownKits        []string `json:"known_kits"`
    DockerSourceOrder []string `json:"docker_source_order,omitempty"`
}
```

The order is saved after every successful base image build (not after every assembly — only when the image is actually built, so the state reflects what Docker has cached).

**Why**: Saving only on successful build means the state always reflects the actual Docker cache state. If a build fails, the old order is preserved and the next attempt can retry.

### Decision 5: Ordering happens in `internal/image`, not `internal/kit`

The ordering logic lives in `internal/image/image.go` (or a new `internal/image/order.go`), not in the kit or agent packages. Kit/agent packages remain pure registries that expose their snippets and priorities.

**Why**: Ordering is a build-time concern. The kit package shouldn't know about Docker caching strategy. This keeps the kit package focused on registration and resolution.

## Risks / Trade-offs

**[Risk] Priority values become stale as kits evolve** → Priorities are set once at registration and rarely need changing. If a kit becomes heavier or lighter, updating its priority is a one-line change. No migration needed since the state tracks order, not priorities.

**[Risk] State file corruption or manual editing** → If `docker_source_order` contains unknown identifiers, they're ignored (filtered out during partition). If the field is missing or empty, all sources are treated as new and sorted by priority (clean slate).

**[Trade-off] First build after upgrade has no previous order** → Falls back to pure priority-based sorting, which is a reasonable default. Subsequent builds benefit from state tracking.

**[Trade-off] Re-sorting after removal may change order unnecessarily** → Only re-sorts the suffix after the removed source. The prefix (everything before the first removal point) is untouched. This is a deliberate choice: Docker invalidates everything after a changed layer anyway, so re-sorting the suffix is free in terms of cache.
