## Why

Asylum supports three AI agents (Claude, Gemini, Codex), each with unique CLI flags, config directories, session detection, and resume semantics. The agent package encapsulates these differences behind a common interface.

## What Changes

- Create `internal/agent` package with Agent interface per PLAN.md section 8
- Implement Claude, Gemini, and Codex agent profiles with command generation per PLAN.md section 3.2
- Session detection logic checking host directories per PLAN.md section 3.2
- Agent registry for lookup by name
- Unit tests for command generation and session detection

## Capabilities

### New Capabilities
- `agent-interface`: Agent interface, registry, and three concrete implementations with command generation and session detection

### Modified Capabilities

None.

## Impact

- Adds `internal/agent/agent.go`, `claude.go`, `gemini.go`, `codex.go`, and test files
- Used by container and CLI packages
