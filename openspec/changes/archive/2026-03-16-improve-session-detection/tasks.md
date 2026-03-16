## Tasks

- [x] **Claude: project-specific session detection** — Encode `projectPath` by replacing `/` with `-`, check for `.jsonl` files in `~/.asylum/agents/claude/projects/<encoded>/`. Return false if directory doesn't exist or has no `.jsonl` files.

- [x] **Gemini: project-specific session detection** — Scan `~/.asylum/agents/gemini/tmp/*/` directories for a `.project_root` file whose trimmed content matches `projectPath`. If found, check for files in that directory's `chats/` subdirectory. Return false if no match.

- [x] **Codex: fix session file location** — Replace `history.jsonl` check with recursive scan for any `rollout-*.jsonl` files under `~/.asylum/agents/codex/sessions/`. This file path matches current Codex versions.

- [x] **Update tests** — Add table-driven tests with `testdata/` fixtures for each agent's `HasSession`: matching project (true), non-matching project (false), empty/missing directories (false).
