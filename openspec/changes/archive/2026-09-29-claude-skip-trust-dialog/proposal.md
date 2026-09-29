# Proposal

## Why

Claude Code asks whether you trust the folder every time it opens a new project in an asylum container. In the home directory it asks on every start, because Claude never persists trust for home. Asylum already treats the container as the trust boundary and runs Claude with `--dangerously-skip-permissions`, so the dialog costs a keypress per session and protects little.

## What Changes

- The Claude agent sets `CLAUDE_CODE_SANDBOXED=1`, which makes Claude Code treat the workspace as trusted and skip the dialog.
- Companion sessions get the variable too, through the existing companion env merging.
- Users opt out with `env: {CLAUDE_CODE_SANDBOXED: ""}`. User config already overrides agent env vars. No new config option.
- The in-container reference, the security model page and the changelog describe the trade-off. Repo-supplied hooks and MCP servers now start without a prompt.

## Capabilities

### New Capabilities
- `claude-workspace-trust`: Claude containers skip the workspace trust dialog, and users can opt out.

### Modified Capabilities

None.

## Impact

- `internal/agent/claude.go`: one env var.
- `assets/asylum-reference.md`, `docs/concepts/security-model.md`: document the behaviour and the opt-out.
- `CHANGELOG.md`: one **Changed** entry.
- Relies on an undocumented Claude Code variable, verified in 2.1.284. If it disappears, the dialog returns and nothing else breaks.
