# Proposal

## Why

Claude Code writes its scratchpad, background task output and edit-diff repos under `/tmp/claude-<uid>/` inside the container. That directory disappears on every container restart, and the host cannot open the files the agent points the user to. Claude now leans on the scratchpad heavily, so both problems show up daily.

## What Changes

- Mount a per-project host directory `~/.asylum/projects/<container>/tmp` into the container at its real host path when Claude is installed in the session (primary agent or companion).
- Set `CLAUDE_CODE_TMPDIR` to that path, so Claude's temp root moves onto the mount and printed paths open unchanged on the host.
- Set `XDG_RUNTIME_DIR` to a container-local `/run/user/<uid>` in the same sessions, so Claude's Unix sockets (`cc-socks/`) stay off the host mount.
- Create `/run/user/<uid>` with mode 0700, owned by the container user, in the base image.
- Leave `TMPDIR` untouched, so other tools keep their temp files in the container's fast local `/tmp`.
- Make cleanup remove read-only subtrees by granting owner write permission and retrying once.
- No pruning. `asylum cleanup` already deletes `~/.asylum/projects/<container>/`, which now includes the temp directory.

## Capabilities

### New Capabilities
- `claude-temp-dir`: Persistent, host-visible Claude Code temp root per project, plus a container-local runtime dir for Claude's sockets.

### Modified Capabilities
- `cleanup-command`: Cleanup also removes read-only subtrees. Persistent temp files make them likely, and plain removal stops at them. The Go module cache is one example.

## Impact

- `internal/container/container.go`: Claude-gated mount and env vars.
- `assets/Dockerfile.core`: creates `/run/user/<uid>`. This changes the base image hash, so every base image rebuilds once.
- `cmd/asylum/main.go`: cleanup removes read-only subtrees.
- `CHANGELOG.md`: one **Added** entry.
- Relies on Claude Code honouring `CLAUDE_CODE_TMPDIR` for the scratchpad. This was verified by reading the 2.1.284 binary. Older builds reportedly ignored it for the scratchpad.
