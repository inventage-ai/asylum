## Context

Asylum currently hardcodes three agent CLI installations (Claude Code, Gemini CLI, Codex) in `Dockerfile.core`. Every user gets all three regardless of which agent they use. The recently-introduced profile system solved the same problem for language toolchains — agent installation should follow the same pattern.

The existing `internal/agent` package handles runtime behavior (command generation, session detection, config directories). Agent *installation* is a separate concern: what Dockerfile instructions are needed, what profile dependencies exist, and what banner lines to show.

## Goals / Non-Goals

**Goals:**
- Configurable agent CLI installation via `agents` field in config YAML and `--agents` CLI flag
- nil (unspecified) defaults to `["claude"]`; explicit empty means none
- Agent install snippets assembled into the Dockerfile at build time, like profile snippets
- Agents can declare a dependency on a profile (e.g., Gemini needs node for npm)
- Dynamic welcome banner: only show version lines for installed agents

**Non-Goals:**
- Changing agent runtime behavior (command generation, session detection, config dirs)
- Adding new agent implementations beyond the existing three
- Agent-to-agent dependencies
- Dynamic agent installation at container start (install is always at image build time)

## Decisions

### 1. Agent install as a struct in the agent package, not a separate package

Agent install metadata (DockerSnippet, ProfileDeps, BannerLine) lives alongside the existing agent implementations in `internal/agent/`. Each agent file (claude.go, gemini.go, codex.go) already exists — adding install metadata there keeps everything about an agent in one place.

A new `AgentInstall` struct holds the build-time concerns. The existing `Agent` interface is unchanged.

```go
type AgentInstall struct {
    Name          string   // matches Agent.Name()
    DockerSnippet string   // Dockerfile RUN instructions
    ProfileDeps   []string // profile names this agent needs (e.g., ["node"])
    BannerLine    string   // shell command for welcome banner
}
```

**Alternative considered**: Separate `internal/agentinstall` package. Rejected because it creates an unnecessary package for 3 structs with no complex logic.

### 2. Install registry parallel to profile registry

A flat `var installs = map[string]*AgentInstall{}` in `internal/agent/install.go`, registered in init() from each agent file. `ResolveInstalls(names *[]string)` returns the active installs. When nil, defaults to `["claude"]` only — Gemini and Codex are opt-in. Empty means none.

### 3. Profile dependency validation at build time

When resolving agent installs, validate that each agent's `ProfileDeps` are satisfied by the active profile set. If Gemini requires `node` but node isn't active, emit a warning but proceed — the agent snippet will fail at Docker build time with a clear error. This is preferable to silently skipping the agent or adding complex auto-activation.

**Alternative considered**: Auto-activate missing profiles. Rejected because it creates surprising behavior — the user explicitly chose their profiles.

### 4. Snippet insertion order: core → profiles → agents → tail

Agent install snippets go after profile snippets and before the tail. This ensures language runtimes (fnm, Node.js) are available when agent CLIs are installed.

### 5. Config field: `agents: *[]string`

- `nil` (not specified): defaults to `["claude"]` — only Claude Code installed
- `[]` (empty list): no agents installed
- `["claude", "gemini", "codex"]`: all three installed
- Last-wins across config layers, same as profiles

**Differs from profiles' nil-means-all**: Profiles default to all for backwards compatibility (existing users expect all languages). Agents default to claude-only because Gemini and Codex add build time and most users don't need them. Users will be able to opt in via onboarding or config.

### 6. fnm + Node.js remains in core

Even though Gemini and Codex depend on Node.js, fnm + Node.js LTS stays in `Dockerfile.core`. It's lightweight infrastructure that multiple agents need. Moving it to the node profile would create a hard dependency from agents to profiles, which the user should be able to configure independently.

## Risks / Trade-offs

**Agent install fails if profile dep is missing** → Mitigated by emitting a warning at resolve time. The Docker build error will be clear ("npm: command not found") and the user can adjust their config.

**Four agents is small enough to not need this abstraction** → True today, but the pattern matches profiles and makes the system consistent. Adding a fifth agent would require only a new file, not touching Dockerfile.core. Opencode is the fourth agent, proving the pattern works.

**Config surface area grows** → One new field (`agents`) with familiar semantics. Minimal cognitive overhead since it mirrors `profiles`.
