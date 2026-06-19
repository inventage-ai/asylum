# Design: Copilot agent

## Context

Asylum currently supports multiple agents (claude, gemini, codex, opencode). Each agent implements the Agent interface in `internal/agent/agent.go` and registers an install snippet via `RegisterInstall`. The `claude` agent provides session resume via `--continue` and supports kit-provided skills via `--add-dir`.

GitHub Copilot CLI (`copilot`) is a terminal-native coding agent with features that overlap and extend those of Claude: LSP server integration, MCP servers, experimental flags, and GitHub auth. The goal is full feature parity for basic workflows: starting sessions, resuming sessions where feasible, LSP config, and MCP server wiring.

## Goals / Non-Goals

**Goals:**
- Add `internal/agent/copilot.go` implementing Agent interface
- Provide an install snippet registered with `RegisterInstall`
- Implement HasSession based on `~/.copilot` layout (spike to confirm exact file layout)
- Provide Command(resume, extraArgs, opts) supporting common flags: `--experimental`, `--model`, and LSP-related env/flags
- Document onboarding steps for GH_TOKEN and repo-level `.github/lsp.json`

**Non-Goals:**
- Re-implement Copilot features or replace their auth flow. Use host token mounting or interactive login.
- Implement deep MCP server orchestration beyond exposing a hook or mount point — define a spec for MCP integration first.

## Decisions

1. Agent structure
   - Add `agents["copilot"] = Copilot{}` in init() and a `RegisterInstall` with a DockerSnippet that uses the official install script (curl | bash) and verifies `copilot --version`.
   - Mirror `claude.go` for environment and Command wrapping (use wrapZsh). Return `wrapZsh("copilot [flags]")`.

2. Session detection (HasSession)
   - Conservative approach: treat presence of any files in `~/.copilot` as potential session state and return true only if a known session artifact exists.
   - Spike to confirm: likely layout includes `sessions`, `chats`, or `state` files. If copilot stores chats per-project, implement logic similar to Claude/Gemini that inspects per-project directories.

3. Resume behavior
   - If Copilot exposes a CLI flag (e.g., `--continue` or `resume`) use it. If not, fallback to starting copilot and let user select previous sessions. Document this limitation.

4. LSP & repo config
   - Support mounting repo `.github/lsp.json` automatically (already mounted). Document how to configure LSP and add a note in the install snippet or onboarding banner.

5. Auth
   - Prefer mounting GH_TOKEN through Asylum's credential system. Document interactive login caveats (browser/device flow may require host interaction).

6. MCP server support
   - Create a spec `openspec/specs/copilot-mcp/spec.md` describing minimal plumbing (env var, mount path, flag to point to MCP server). Implementation deferred after spec.

## Risks / Trade-offs

- [Unknown] Copilot session storage layout → Spike to confirm. Mitigation: default to conservative detection and add spec to refine.
- [Auth UX] Interactive login may fail inside containers without browser access. Mitigation: recommend GH_TOKEN or mounting host auth files.
- [MCP complexity] Full MCP orchestration can be complex; implement minimal support first and iterate via spec.

## Open Questions

- Where exactly does Copilot persist per-project sessions? (spike)
- Does Copilot support a CLI resume flag? (spike)
- Which env vars or CLI flags control MCP server endpoints or additional skill dirs?

Next: create tasks.md with a breakdown for implementation and spikes.
