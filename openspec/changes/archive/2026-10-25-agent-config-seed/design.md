# Design

## Context

`EnsureAgentConfig` (isolated) and `EnsureAgentConfigAt` (project) in `internal/container` create the agent's config directory on first run and copy the native host directory with `copyDir`. `cmd/asylum/main.go` picks the function by isolation level. `AgentConfig` fields merge last-wins via reflection, so a new string field needs no merge code.

## Goals / Non-Goals

**Goals:** let users choose `none`, `settings` or `full` per agent; keep `full` as default; keep the repo layer from choosing.

**Non-Goals:** credentials outside the config directory (macOS Keychain, `~/.claude.json`, OpenCode's `~/.local/share/opencode/auth.json`); re-seeding an existing directory; fixing OpenCode's config location.

## Decisions

- **Allowlist on the agent (`SeedSettings() []string`).** Which files are settings is agent knowledge, like `NativeConfigDir()`. Entries are paths relative to the native dir and may be nested (Pi uses `agent/...`). A central map keyed by agent name was rejected because it separates the list from the agent definition.
- **One function for both targets.** `EnsureAgentConfig` becomes a thin wrapper around `EnsureAgentConfigAt`; the two were duplicates and would otherwise both need the seed logic.
- **Validate the mode when seeding.** An unknown value errors only when a directory is actually created, so a typo in a `shared` setup does not break runs. Validating in `Load` was rejected: `config` has no notion of which agents run.
- **Filter `seed` in `Load` for the project layer**, the only place that knows the source file, mirroring the env-from-host handling (#42).
- **Credentials are part of `settings`**, so a new project does not force a new login where credentials live in files (Linux).
- **Claude `plugins/` excluded.** Its JSON stores absolute host paths that break when the container home differs.

## Risks / Trade-offs

- [Allowlists for Codex, Gemini, Copilot and Pi are based on documentation, not tested installs] → Missing entries are skipped silently; a wrong list only copies less. Lists can be extended without spec changes.
- [Files named in `settings.json` outside the allowlist (e.g. a statusline script directly in `~/.claude`) are not copied] → Claude's list includes `hooks/`; the docs name the limitation and the workaround (`hooks/` or `full`).
- [Symlinked entries (dotfile managers) must be copied] → `copyDir` walks the resolved path; this also fixes `full` when the whole native dir is a symlink, which previously copied nothing.
- [Merge conflict with #42, which edits the same place in `Load`] → Trivial to resolve; whichever merges second combines the two blocks.
