## Approach

Make each agent's `HasSession(projectPath string)` check for session data matching the specific project path, instead of checking for any session data at all. The `projectPath` parameter is already passed but unused.

## Agent-Specific Detection

### Claude

Claude stores conversations at `~/.claude/projects/<encoded-path>/` where `<encoded-path>` replaces `/` with `-` in the absolute project path (e.g. `/Users/simon/Tools/asylum` → `-Users-simon-Tools-asylum`).

**Detection**: Encode the project path, check for `.jsonl` files in that specific directory.

```
~/.asylum/agents/claude/projects/-Users-simon-Tools-asylum/*.jsonl
```

### Gemini

Gemini uses a project registry with human-readable slugs. Each project directory gets `~/.gemini/tmp/<slug>/` with a `.project_root` file containing the absolute path.

**Detection**: Scan `~/.asylum/agents/gemini/tmp/*/` for a `.project_root` file whose content matches the project path, then check for files in its `chats/` subdirectory.

Legacy format uses SHA-256 hashes instead of slugs — the same scan approach handles both.

### Codex

Codex stores sessions globally at `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl`. The working directory is embedded inside each file and filtered at runtime by `codex resume --last`.

**Detection**: Check if any `rollout-*.jsonl` files exist under `~/.asylum/agents/codex/sessions/`. Since Codex does its own CWD filtering, a coarse check is acceptable — `codex resume --last` handles the "no match" case gracefully.

The current implementation checks `history.jsonl` which doesn't exist in current Codex versions.

## Changes

- `internal/agent/claude.go` — `HasSession`: encode project path, check for `.jsonl` files in matching project directory
- `internal/agent/gemini.go` — `HasSession`: scan `tmp/*/` for matching `.project_root`, check `chats/` for session files
- `internal/agent/codex.go` — `HasSession`: check for any `rollout-*.jsonl` under `sessions/` recursively
- `internal/agent/agent_test.go` — update tests with project-specific fixtures

## What Won't Change

- The `Agent` interface — `HasSession(projectPath string)` signature is already correct
- Container assembly — `containerCommand` already passes `opts.ProjectDir`
- CLI flags — no new flags needed
