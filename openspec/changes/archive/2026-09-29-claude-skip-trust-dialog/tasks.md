# Tasks

## 1. Implementation

- [x] 1.1 Add `CLAUDE_CODE_SANDBOXED: "1"` to `Claude.EnvVars()` in `internal/agent/claude.go`. Verify with a test asserting the var is in the Claude env map.
- [x] 1.2 Add a `RunArgs` test showing a user `env:` entry `CLAUDE_CODE_SANDBOXED: ""` overrides the agent value. Verify that the resolved args contain `CLAUDE_CODE_SANDBOXED=` and not `CLAUDE_CODE_SANDBOXED=1`.
- [x] 1.3 Run `go test ./...` and `go vet ./...`. Both should pass.

## 2. Documentation

- [x] 2.1 In `assets/asylum-reference.md`, note that the trust dialog is skipped and how to restore it. Verify the text names the empty-string opt-out.
- [x] 2.2 In `docs/concepts/security-model.md`, add the trade-off that repo-supplied hooks and MCP servers start without a prompt, plus the opt-out. Verify by reading the page.
- [x] 2.3 Add a **Changed** entry to `CHANGELOG.md` under Unreleased. Verify that it names the dialog, the home-directory case and the opt-out.

## 3. End-to-end check

- [x] 3.1 Start a Claude container in a folder Claude has never trusted and confirm no trust dialog appears. Then set `CLAUDE_CODE_SANDBOXED: ""` in `env:`, restart, and confirm the dialog returns.
