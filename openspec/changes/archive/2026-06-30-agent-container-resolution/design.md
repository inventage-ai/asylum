## Context

Asylum runs one detached container per project, named `asylum-<sha256(project_dir)[:N]>-<basename>`, and execs every session (agent/shell/run) into it. The container is removed when its last session exits (`HasOtherSessions`). The requested agent is always added to the install set (`main.go` ~239), so it is always baked into the freshly built image; switching the active agent changes the image hash, which `checkStaleContainer` already detects via image-ID mismatch.

The gap: when a project's container was built for one agent and the user requests another ad-hoc, the only paths are "exec in anyway and fail" or "tear down and rebuild the same name" (the shipped prototype), which kills the running agent's session.

## Goals / Non-Goals

**Goals:**
- Reuse the existing container when it already supports the requested agent.
- Otherwise run a *second* container for the same project, keyed by `hash(project + agents)`, without touching the first.
- Keep the primary container name byte-identical so existing containers, ports, and project data are untouched (no migration).
- Make "configure the agent properly" the path to a first-class container with ports.

**Non-Goals:**
- Port allocation for secondary containers (they forward none).
- Reworking container ephemerality.
- Teaching `asylum cleanup` to enumerate all of a project's containers.

## Decisions

- **Primary name unchanged; secondary = `hash(project + sorted_agents)`.** `ContainerName(projectDir)` becomes `ContainerName(projectDir, agents)`. With the default/configured agent set, the primary path must produce the exact current string (same hash width, same suffix) so nothing migrates. Secondaries use a distinct hash that folds in the sorted agent set.
- **First-come owns the primary name.** Whichever agent runs first (default or configured) takes `hash(project)`; incompatible ad-hoc agents spill to `hash(project+agents)`. Order-dependent but harmless — it only decides which container is portless.
- **Agent support is read from the `asylum.agents` label**, set from the image's baked agent set (sorted, comma-joined). Membership test = requested agent ∈ label.
- **Legacy containers (no label) are assumed to support the default agent (`claude`).** This preserves "normal operation stays as it is" — existing setups keep working without a surprise rebuild. A non-default agent against a legacy container spills to a secondary, leaving the legacy container intact.
- **Two-pass resolution in `main.go`:**
  1. `name = ContainerName(project, configuredAgents)`. If running and supports the active agent → exec (today's path).
  2. Else `name = ContainerName(project, [activeAgent + companions])`; repeat the running-container lookup; build/start there if absent.
- **Secondary containers skip port allocation.** `RunArgs` gains a "secondary" signal; when set it omits the ports kit's `-p` args and asylum does not call `ports.Allocate`. `internal/ports` is unchanged.
- **Remove the rebuild-in-place path.** Delete the `ContainerHasAgent` → `RemoveContainer` branch added by the prototype; keep the label emission and the `InspectLabels`/`ContainerHasAgent` helpers.

## Risks / Trade-offs

- **Secondaries have no forwarded ports.** Acceptable: they are for review/ad-hoc work. The escape hatch is to add the agent to config, which bakes it into the portful primary. Documented behavior, not a silent gap.
- **Orphaned secondaries.** If a secondary's process is killed externally without a clean last-session exit, `asylum cleanup` (which targets the primary name) won't sweep it. Accepted and deferred to a follow-up cleanup change; the common case self-cleans via the session counter.
- **Spec/code hash-width drift.** The current spec says `[:12]`, the code uses `[:6]`. This change does not touch the width; the primary string is preserved exactly as the code emits it today.
- **A newly-configured agent can spill to a portless secondary while the primary is concurrently live.** Support is decided from the *running* primary's `asylum.agents` label, not from current config. So if a claude session is open (primary labeled `claude`), you add `pi` to config, and you launch `pi` in another terminal, `pi` spills to a portless secondary even though it is now "properly configured". This is accepted, not fixed: the alternative — rebuilding the primary in place — would either kill the live claude session or, if the user declines the restart prompt, exec `pi` into a container without `pi` (the original bug). The case is narrow (requires a concurrent live primary) and self-correcting: once the claude session exits the primary is removed, and the next `pi` launch takes the portful primary. A future refinement could gate the spill on `CheckOtherSessions` so a primary that is running but idle (e.g. a leftover from a crash) is reclaimed with ports instead of spilled.
