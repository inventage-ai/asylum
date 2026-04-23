## Context

Asylum supports multiple coding agents (claude, gemini, codex, opencode) through a uniform Agent interface in `internal/agent/`. Each agent is a struct implementing the interface, registered via `init()`, and paired with an `AgentInstall` for Dockerfile snippets and banner lines. Pi is the Asylum project's own coding agent and should be supported as a first-class option.

## Goals / Non-Goals

**Goals:**
- Add pi as a selectable agent alongside existing agents
- Follow the established pattern (struct + init registration + install snippet)
- Support session detection and command generation for pi

**Non-Goals:**
- Modifying existing agent implementations
- Changing the Agent interface
- Adding new config options or CLI flags

## Decisions

- **Installation method**: pi is installed via npm (`@mariozechner/pi-coding-agent`) through fnm-managed node, following the same pattern as gemini and codex. Requires the `node` kit.
- **Config directory**: `~/.pi` (native), `~/.asylum/agents/pi` (isolated) — consistent with the naming convention used by all other agents.
- **Session detection**: pi stores session data in its config directory. We'll check for existing session files similar to how other agents do it.
- **Command flags**: pi uses `--no-resume` / no flag for fresh/resume sessions. Exact flags determined by pi's CLI interface.

## Risks / Trade-offs

- pi's CLI flags may differ from other agents → mitigated by testing the command generation against pi's actual CLI
