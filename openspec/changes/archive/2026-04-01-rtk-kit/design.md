## Context

Asylum uses a kit system where each kit contributes Dockerfile snippets (build-time installation), entrypoint snippets (runtime setup), rules snippets (agent documentation), config snippets, and banner lines. RTK is a Rust binary that intercepts shell commands and compresses their output to reduce LLM token usage. For Claude Code, `rtk init -g` generates:
- `~/.claude/hooks/rtk-rewrite.sh` — a PreToolUse hook that rewrites Bash commands through RTK
- `~/.claude/RTK.md` — awareness doc teaching the agent about RTK meta-commands
- Patches `~/.claude/settings.json` to register the hook

Following the ast-grep kit pattern, RTK artifacts should be generated at build time and mounted into place at runtime.

## Goals / Non-Goals

**Goals:**
- Install RTK binary and generate its hook artifacts at Docker build time
- Mount hooks and RTK.md into the agent config directory at container start (like ast-grep skills)
- Register the RTK hook in settings.json at container start
- Document RTK usage in a rules snippet and docs page

**Non-Goals:**
- Exposing RTK configuration options (e.g., ultra-compact mode) via asylum config — users can set those via env vars or `.rtkrc` if needed
- Supporting per-project RTK settings in asylum config
- Adding RTK analytics/gain tracking to the asylum TUI
- Multi-agent RTK init (Gemini, Codex) — start with Claude Code only, extend later

## Decisions

### Install via curl install script
RTK provides a `curl -fsSL ... | sh` installer that detects OS/arch and downloads the correct binary. This mirrors how asylum installs other tools and avoids maintaining architecture-specific download URLs in the Dockerfile snippet. The alternative (cargo install) would require Rust toolchain in the image, which is too heavy.

### Build-time init, runtime mount (ast-grep pattern)
Run `rtk init -g` during `docker build` to generate the hook script and RTK.md, then move them to `/tmp/asylum-kit-rtk/`. At container start, the entrypoint mounts these into `~/.claude/hooks/` and `~/.claude/` using `mount --bind` (same as ast-grep). This keeps the entrypoint fast — no network calls or init logic at startup.

The settings.json hook registration is handled by a lightweight `jq` patch in the entrypoint (or inline script), since settings.json is mounted from the host and doesn't exist at build time.

### Claude Code only (for now)
RTK's hook mechanism is agent-specific. Start with Claude Code (the default and most common agent). Gemini/Codex support can be added later as separate entrypoint branches. This avoids over-engineering the first version.

### Opt-in tier, no dependencies
RTK is a standalone Rust binary with no runtime dependencies. Unlike ast-grep (which depends on the node kit for npm install), RTK has no kit dependencies. It's opt-in since not all users want command interception. `NeedsMount: true` since it uses `mount --bind` at runtime.

## Risks / Trade-offs

- **RTK install script fetches from GitHub at build time** → If GitHub is down, image build fails. Mitigation: this is the same risk as other curl-installed tools; the install is cached in the Docker layer.
- **settings.json patching at runtime** → Could conflict with user's existing hooks. Mitigation: the patch is additive (appends to the PreToolUse hooks array); existing hooks are preserved. If settings.json doesn't exist or has no hooks section, the patch creates it.
- **Claude-only initially** → Users of Gemini/Codex won't get RTK integration. Mitigation: RTK binary is still available on PATH; users can manually `rtk init` if needed. Full multi-agent support is a follow-up.
