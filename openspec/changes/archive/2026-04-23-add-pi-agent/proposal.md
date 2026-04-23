## Why

Pi is a coding agent that runs inside Asylum containers. Users should be able to select pi as their agent, just like claude, gemini, codex, and opencode. Currently pi is not recognized by the agent registry.

## What Changes

- Add `pi` as a new agent implementation in `internal/agent/`, following the existing pattern (claude.go, gemini.go, etc.)
- Register pi in the agent install system with Dockerfile snippet for installation and a banner line
- Update the agent-interface spec to include pi

## Capabilities

### New Capabilities

- `pi-agent`: Pi agent implementation — registry entry, install snippet, command generation, session detection

### Modified Capabilities

- `agent-interface`: Add pi to the agent registry and command generation requirements

## Impact

- `internal/agent/pi.go` — new file
- `openspec/specs/agent-interface/spec.md` — updated with pi scenarios
- No breaking changes; purely additive
