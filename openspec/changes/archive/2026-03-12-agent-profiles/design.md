## Context

PLAN.md sections 3.1–3.4 define agent profiles with specific binaries, config dirs, YOLO flags, resume mechanisms, and session detection. The Agent interface (section 8) is prescribed.

## Goals / Non-Goals

**Goals:**
- Agent interface with all methods from PLAN.md section 8
- Three implementations: Claude, Gemini, Codex
- Command generation for all modes (default, new, with args)
- Session detection via filesystem checks
- Registry function to get agent by name

**Non-Goals:**
- No agent config seeding here — that belongs in the container/runtime layer
- No actual Docker interaction

## Decisions

- **Interface with concrete types**: The Agent interface is justified here — three implementations with different behavior. Each agent is a struct implementing the interface.
- **Session detection**: `HasSession` takes a base config dir (the `~/.asylum/agents/<agent>/` path) and the project path. It checks for agent-specific session markers. For testability, the config dir is passed in rather than hardcoded.
- **Command wrapping**: All agent commands are wrapped in `zsh -c "source ~/.zshrc && exec <cmd>"` to ensure NVM/SDKMAN are available.
- **Registry**: A simple `Get(name string) (Agent, error)` function with a map. No need for registration patterns.

## Risks / Trade-offs

- Session detection depends on filesystem layout. Tests use temp directories with fake session data to verify the logic without real agent installations.
