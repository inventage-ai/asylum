## Why

Asylum frequently passes `--continue` / `--resume` to agents when no resumable session actually exists for the current project directory. This happens because session detection is too coarse:

- **Claude**: Checks if `~/.asylum/agents/claude/projects/` has *any* entries, not whether one matches the current project. Since the config dir is seeded from the host `~/.claude`, it picks up project data from other directories.
- **Gemini**: Checks if `~/.asylum/agents/gemini/tmp/` has any entries — same problem.
- **Codex**: Checks if `history.jsonl` is non-empty — not project-specific at all.

The result: agents crash or show confusing errors when asked to resume a non-existent conversation.

## What Changes

- Make `HasSession` project-aware for all three agents by checking for session data matching the specific project path, not just "any session exists."
- Claude: check for a matching project directory under `projects/` using Claude's path encoding (`/` → `-`)
- Gemini: check for session files under the project-specific temp directory
- Codex: check for project-specific entries in `history.jsonl`

## Capabilities

### New Capabilities

None — this improves an existing mechanism.

### Modified Capabilities

- `session-detection`: All three agent `HasSession` implementations become project-path-aware, only returning `true` when the specific project has resumable data.

## Impact

- Modifies `internal/agent/claude.go`, `gemini.go`, `codex.go`
- Updates tests in `internal/agent/agent_test.go`
- No config or CLI changes
