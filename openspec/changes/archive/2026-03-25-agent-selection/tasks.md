## 1. Agent Install Registry

- [x] 1.1 Create `internal/agent/install.go` with the `AgentInstall` struct (Name, DockerSnippet, ProfileDeps, BannerLine), registry map, `RegisterInstall`, `AllInstalls`, and `ResolveInstalls(names *[]string, activeProfiles []string)` (nil=claude-only, empty=none, profile dep warnings)
- [x] 1.2 Add install definitions to existing agent files: `internal/agent/claude.go` (curl install script), `internal/agent/gemini.go` (npm install, ProfileDeps: ["node"]), `internal/agent/codex.go` (npm install, ProfileDeps: ["node"]), and create `internal/agent/opencode.go` (go install, no profile deps) with Agent interface implementation + install definition
- [x] 1.3 Write tests for `ResolveInstalls`: nil-defaults-to-claude, empty-means-none, explicit all, specific selection, unknown agent error, profile dep warning when dep missing, no warning when dep satisfied

## 2. Config Integration

- [x] 2.1 Add `Agents *[]string` field to Config struct with YAML tag `agents`, same `*[]string` pattern as Profiles
- [x] 2.2 Update config merge logic: `agents` follows last-wins semantics (same as profiles)
- [x] 2.3 Add `--agents` CLI flag to `cmd/asylum/main.go` (comma-separated), wire into CLIFlags and config override
- [x] 2.4 Write tests for config merge with agents: nil stays nil, overlay replaces, empty replaces non-nil, CLI flag overrides

## 3. Dockerfile Core Cleanup

- [x] 3.1 Remove Claude Code, Gemini CLI, and Codex install lines from `assets/Dockerfile.core`
- [x] 3.2 Verify: core Dockerfile no longer contains any agent CLI installations

## 4. Image Build Integration

- [x] 4.1 Add `AssembleAgentSnippets` helper to `internal/agent/install.go` (concatenates DockerSnippets from active installs)
- [x] 4.2 Add `AssembleAgentBannerLines` helper (concatenates BannerLines from active installs)
- [x] 4.3 Update `assembleDockerfile` in `internal/image/image.go` to insert agent snippets after profile snippets and before tail
- [x] 4.4 Update `assembleEntrypoint` to insert agent banner lines into the tail placeholder (alongside profile banner lines)
- [x] 4.5 Update `baseHash` to include agent install snippets
- [x] 4.6 Update `EnsureBase` signature to accept agent installs

## 5. Entrypoint Banner Update

- [x] 5.1 Update `assets/entrypoint.tail`: move hardcoded agent version lines (Claude, Gemini, Codex) out of the static tail and into the `PROFILE_BANNER_PLACEHOLDER` section (handled by assembled banner lines)
- [x] 5.2 Verify: with all agents active, banner output matches current behavior; with subset, only selected agents shown

## 6. Main Wiring

- [x] 6.1 Update `cmd/asylum/main.go`: resolve agents from merged config via `agent.ResolveInstalls`, pass active installs to image builder
- [x] 6.2 Update `printUsage()` to document `--agents` flag
- [x] 6.3 Add CHANGELOG entry under Unreleased

## 7. Testing

- [x] 7.1 Unit tests for Dockerfile assembly with agent snippets: all agents, subset, none
- [x] 7.2 Unit tests for entrypoint banner with agent lines: all agents, subset, none
- [x] 7.3 Verify all existing tests still pass after Dockerfile.core cleanup
