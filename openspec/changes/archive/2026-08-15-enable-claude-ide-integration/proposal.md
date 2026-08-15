## Why

Claude Code's `/ide` command does not work inside asylum containers. The host IDE (VS Code, IntelliJ) publishes a lock file describing a WebSocket MCP server bound to the host's `127.0.0.1`, and the containerized CLI dials that address inside its own network namespace, where nothing listens. Users lose selection context, diagnostics, and the diff view when working in a sandbox.

The CLI already reads a `CLAUDE_CODE_IDE_HOST_OVERRIDE` environment variable that replaces the host it dials. On Docker Desktop, `host.docker.internal` reaches host loopback services, so exporting that variable is sufficient to make `/ide` work with no forwarding process, no bind mount, and no kit.

## What Changes

- `Claude.EnvVars()` returns `CLAUDE_CODE_IDE_HOST_OVERRIDE=host.docker.internal` alongside the existing `CLAUDE_CONFIG_DIR`, so every Claude container dials the host instead of its own loopback when connecting to an IDE.
- `assets/asylum-reference.md` documents that `/ide` works, and states the two conditions under which it does: a Docker Desktop engine and `shared` agent config isolation.
- `CHANGELOG.md` gains an **Added** entry.

Deliberately out of scope for this change: bridging lock files into `isolated` and `project` isolation modes, and reaching a host IDE from a native Linux Docker engine. Both are noted in design.md as follow-on work if the feature proves useful.

## Capabilities

### New Capabilities
- `claude-ide-integration`: Claude Code containers resolve the host IDE's WebSocket endpoint via `host.docker.internal` so `/ide` can connect to an IDE running on the host.

### Modified Capabilities

(none - `agent-interface` already specifies that agents contribute environment variables via `EnvVars`; this change adds a variable without changing the interface contract)

## Impact

- `internal/agent/claude.go` - one additional map entry in `EnvVars()`.
- `internal/agent/claude_test.go` - assertion on the returned map, if one exists for `EnvVars`.
- `assets/asylum-reference.md` - user-facing documentation, embedded into the image and mounted at `~/.claude/asylum-reference.md`.
- `CHANGELOG.md`.
- No change to the broker, container assembly, kits, Dockerfile, or entrypoint. No new dependencies. No behavior change for other agents.
