## 1. Spikes and discovery

- [x] 1.1 Spike: sessions live in `~/.copilot/session-state/<session-id>/` (conversation history, tool calls, files touched) plus a SQLite session store; the config root is controlled by `COPILOT_HOME` (defaults to `~/.copilot`).
- [x] 1.2 Spike: copilot supports an explicit `--resume` CLI flag (with a `/resume` slash command available in interactive mode); selecting a previous session restores its saved context.
- [x] 1.3 Spike: auth env precedence is `COPILOT_GITHUB_TOKEN` → `GH_TOKEN` → `GITHUB_TOKEN`. Config dir is overridden via `COPILOT_HOME`. MCP servers are configured in `~/.copilot/mcp-config.json` and managed via `copilot mcp` subcommands; GitHub's MCP server is built-in.

## 2. Agent implementation

- [x] 2.1 Add `internal/agent/copilot.go` implementing Agent interface (Name, Binary, NativeConfigDir, ContainerConfigDir, AsylumConfigDir, EnvVars, HasSession, Command).
- [x] 2.2 Register install snippet via `RegisterInstall` with DockerSnippet and BannerLine. Document KitDeps if required.
- [x] 2.3 Add unit tests mirroring patterns in `internal/agent/agent_test.go` and `install_test.go` for the copilot agent.

## 3. Session & resume

- [x] 3.1 `HasSession` checks an Asylum-owned per-project marker (`<configDir>/asylum-projects/<encoded>/.has_session`) written by `WriteMarker` after a successful first launch in that project. Mirrors codex's pattern. Plain `session-state/` content is NOT used because copilot's `--resume` picker is global and would otherwise leak unrelated project context.
- [x] 3.2 `Command` emits `copilot --resume <args>` when `HasSession` is true, otherwise plain `copilot <args>`. Falls back to fresh start when no session is present (no special fallback path needed).

## 4. Auth & onboarding

- [x] 4.1 Update onboarding docs to include GH_TOKEN guidance and interactive login caveats.
- [x] 4.2 GH_TOKEN flows via the existing `github` kit credential mount: the kit writes `~/.config/gh/hosts.yml` from the host's `gh auth token`, and the copilot agent's launch wrapper exports `GH_TOKEN="$(gh auth token)"` at startup. No new credential mechanism required.

## 5. MCP & LSP support

- [x] 5.1 Add spec `openspec/specs/copilot-mcp/spec.md` describing minimal MCP plumbing and required env/flags.
- [x] 5.2 Document LSP configuration mapping (`~/.copilot/lsp-config.json` and repo `.github/lsp.json`) in docs and copilot agent README.

## 6. Integration & verification

- [x] 6.1 Add a manual integration test plan to verify auth flow and session resume (documented in e2e/ or integration/ as needed).
- [x] 6.2 `go test ./...` passes (554/554) with the copilot agent registered; CI runs `go test ./...` already, no workflow change needed.

## 7. Cleanup & docs

- [x] 7.1 Update README.md / CLAUDE.md (or add COPILOT.md) with usage notes.
- [x] 7.2 Add openspec specs for copilot-session-detection and copilot-mcp.
- [x] 7.3 `CHANGELOG.md` Unreleased entry under **Added** mentions copilot agent + session detection + automatic GH_TOKEN passthrough via the github kit.
