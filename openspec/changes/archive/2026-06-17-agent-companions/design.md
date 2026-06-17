## Context

Asylum supports multiple coding agents (Claude, Codex, Gemini, Copilot, Opencode, Pi). At runtime exactly one is the "primary" — its config dir is mounted, its env vars are set, its `Command()` is launched (`internal/container/container.go:285`).

Recent work added a codex Claude Code plugin: while Claude is running, it shells out to the `codex` CLI inside the same container. That CLI needs `~/.codex` mounted (for auth/session state — writable) and `CODEX_HOME` set. Today neither happens unless codex itself is the primary agent.

The user has already settled on a YAML shape and four key decisions (see proposal). This design records the technical choices and the seams in the existing code.

## Goals / Non-Goals

**Goals:**
- Let a primary agent declare companions whose config dirs and env vars are present in the container at runtime.
- Reuse the existing `Agent` interface (no new methods) and the existing isolation machinery (`config.AgentIsolation` + `agent.ResolveConfigDir`).
- Fail loudly when a companion isn't installed, rather than silently degrading.

**Non-Goals:**
- Launching multiple agents in one container session. Only the primary's `Command()` runs.
- Cross-agent session/resume coupling. `HasSession()` and resume continue to consider only the primary.
- Changing `~/.agents` shared-mode semantics. That mount remains gated on the primary's isolation level being `shared`.
- Auto-install. Companions must already be installed via the project's agent install set.
- Sandbox rules / `asylum-reference.md` for companions. These are dropped into the primary's `ContainerConfigDir` only.

## Decisions

### D1: Configure companions on the primary, one-directional

```yaml
agents:
  claude:
    config: shared
    companions: [codex]
  codex:
    config: shared
```

**Rationale:** Matches the real asymmetry of the plugin (Claude plugin calls codex, not the reverse). Keeps the abstraction simple — no relationship table, just a list on each agent.

**Alternative considered:** `always-mount: true` on the companion. Rejected — loses targeting (you can't run Claude *without* codex), and couples two agents that may be independent in other projects.

### D2: Companion's own isolation setting resolves the host source

When mounting codex's config while Claude is primary:
- If `agents.codex.config: shared` → mount host `~/.codex`
- If `agents.codex.config: isolated` → mount `~/.asylum/agents/codex`
- If `agents.codex.config: project` → mount `~/.asylum/projects/<container>/codex-config`

Implementation reuses `agent.ResolveConfigDir(companion, isolationLevel, cname)` — the same helper used for the primary at `container.go:286`.

**Rationale:** Each agent's isolation is a property of the agent, not of who launches it. If the user said codex should use shared config, that's true whether codex is launched directly or invoked by a plugin from inside a Claude session.

**Alternative considered:** Inherit the primary's isolation. Rejected — would surprise users when companion writes (auth tokens, session state) land in unexpected places.

### D3: Companion mount is writable, just like the primary

Codex needs to write auth tokens and rotate session state in `~/.codex`. The primary's mount has no `:ro` (see `vol(hostConfigDir, containerConfigDir, "")` at `container.go:298`). Companions use the same default.

**Trade-off:** A misbehaving plugin could corrupt host config in `shared` mode. This is no worse than running codex as the primary in `shared` mode, which is already allowed.

### D4: Env var merge — primary wins on key collisions

`coreEnvVars` (`container.go:321`) currently iterates `opts.Agent.EnvVars()` only. Change: also iterate companions, but only add a key if not already set by the primary.

**Rationale:** Two agents are unlikely to share env var names, but if they did, the primary's launch contract must not be silently overridden.

**Open detail:** Log a warning when a collision occurs so misconfiguration is visible.

### D5: Companion validation at run assembly, not config load

The check "is this companion installed?" needs the resolved set of installed agents for the current project. That set lives in `AgentInstall` registrations gated by config/kits. The cleanest place is the same path that builds the run args — fail before `docker run`, with a message naming the offending agent and how to install it.

**Rationale:** Config load doesn't know about the install set; doing the check there would require duplicating the install resolution. Doing it later (e.g. inside the container) is too late.

### D6: Self-reference and cycles

- `agents.claude.companions: [claude]` → ignore the self entry (no error, no double-mount).
- Cycles across pairs are not possible in a single run (only the primary's companion list is read), but `agents.claude.companions: [codex, codex]` should be de-duplicated.

### D7b: Companion list is last-wins, with explicit-empty meaning "clear"

The `Companions` field is `*[]string`, not `[]string`. A nil pointer means "unset, inherit base"; a non-nil pointer (even to an empty slice) means "explicitly set, replace base". So a higher-precedence layer (e.g. `.asylum.local`) can clear an inherited companion list with `companions: []`.

**Rationale:** Companions mount writable config dirs (potentially exposing host auth tokens in shared mode). A project must be able to opt out of inherited companions. Concat semantics would make exposure append-only across layers and impossible to retract per project. Last-wins matches the spec ("merges using last-wins overlay semantics consistent with other agent-level fields") and the established convention for tri-state fields (`*bool` in `KitConfig`).

**Trade-off:** Users who want to add to an inherited list must re-list the inherited entries plus the additions. Acceptable: companion lists are short by design, and the safety property (explicit clearing) is more important than concat ergonomics.

### D7: One-directional semantics confirmed

Running `asylum codex` does **not** pull in claude's mounts because of `agents.claude.companions: [codex]`. Only the *primary's* companion list is consulted.

## Risks / Trade-offs

- **Risk:** Plugin author assumes companion is always present, but user has a project without `codex` installed and gets a hard error on `asylum claude`.
  - **Mitigation:** Error message names the missing agent and points at the install/kit config. Document the contract.
- **Risk:** Shared-mode writes from a buggy plugin corrupt host `~/.codex`.
  - **Mitigation:** No worse than running codex directly in shared mode. Users who want isolation can set `agents.codex.config: isolated`.
- **Risk:** Confusion about which agent's env vars win.
  - **Mitigation:** Documented (D4) and warned at runtime on collision.
- **Trade-off:** Companions don't get sandbox rules. If a companion ever needs `asylum-sandbox.md` (e.g. another Claude-style agent), this design defers that — it's not needed for the codex CLI plugin case.

## Migration Plan

Additive YAML field; no migration. Existing configs continue to work. Users who want the codex plugin add `companions: [codex]` to their claude agent config and ensure codex is in the agent install set.

Rollback: remove the `companions` field. No persistent state created.

## Open Questions

- Should there be a CLI override (`--with codex`) for one-off use, independent of YAML? Not required by the immediate use case; deferred unless asked.
- Should the first-run wizard surface companions? Probably no — too niche for the common path. Leave as YAML-only for now.
