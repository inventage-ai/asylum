# IDE Integration

Claude Code's `/ide` command connects a session to an IDE (VS Code, IntelliJ) so the agent gets your current selection, diagnostics from the language server, and an in-editor diff view. Inside Asylum the agent runs in a container while the IDE runs on your host, and this page describes exactly how far that bridge reaches today.

## How it works

The IDE plugin starts a small WebSocket server on the host and publishes a lock file describing it in `~/.claude/ide/`. Claude Code reads those lock files and dials the port they name — on the container's own loopback by default, where nothing is listening.

Asylum sets `CLAUDE_CODE_IDE_HOST_OVERRIDE=host.docker.internal` in every Claude container, so the CLI dials the host instead. That is the entire mechanism: no forwarding process, no extra mount, no kit, nothing to enable.

## Requirements

All four must hold. None of them produce an error message when they don't — `/ide` simply lists no IDE.

| Requirement | Why |
|-------------|-----|
| Docker Desktop engine | The container reaches the host through `host.docker.internal` |
| `shared` config isolation | The container must read the same lock files the host IDE writes |
| Claude Code as the agent | Other agents have their own IDE mechanisms |
| IDE open on the project | Claude Code matches the lock file's workspace against its working directory |

## Connecting

Run `/ide` inside the session and pick your IDE. The connection is per-session and manual by default.

To connect automatically on every session, enable Claude Code's own `autoConnectIde` setting. It lives in the config directory, which is mounted from the host in `shared` mode, so it persists across containers and rebuilds. Asylum has no separate setting for this.

## Limitations

### Docker Desktop only

On Docker Desktop (macOS, Windows) `host.docker.internal` routes to the host's loopback interface, so the IDE's server is reachable. On a **native Linux Docker engine** the same name resolves to the bridge gateway, which cannot reach a process bound to the host's `127.0.0.1`. There is no workaround in Asylum today; `/ide` will find no IDE.

### `shared` config isolation only

The IDE writes its lock files into the host's real `~/.claude/ide/`. Only [`shared` isolation](isolation.md) mounts that directory into the container.

```yaml
agents:
  claude:
    config: shared   # the default
```

Under `isolated` or `project` the container gets its own config directory, sees no lock files at all, and `/ide` lists nothing. Bridging lock files into those modes would require mirroring them into the container and is not implemented.

### Claude Code only

The override is contributed by the Claude agent alone. Codex has its own IDE integration with a separate discovery mechanism, which Asylum does not currently bridge. Gemini CLI, OpenCode and Pi are unaffected.

### The IDE must be open on the project directory

Claude Code compares the lock file's workspace folders against its working directory and requires a match — the same directory, or a folder containing it. Asylum mounts your project at its real host path rather than under `/workspace`, so paths line up and this normally just works. It fails if the IDE is open on a sibling directory, or on a project checked out at a different path than the one you ran `asylum` from.

### JetBrains lock files appear unpredictably

The JetBrains plugin does not always publish its lock file when you expect. If `/ide` shows an empty list while IntelliJ is open on the project, reopen the project or restart the IDE and try again. This is plugin behaviour, outside Asylum's control.

### The mechanism is undocumented upstream

`CLAUDE_CODE_IDE_HOST_OVERRIDE` is an internal of the Claude Code CLI, not a public interface. A future release could rename or remove it, at which point `/ide` stops finding IDEs until Asylum adapts. The failure mode is the feature not working, nothing worse.

## Troubleshooting an empty `/ide` list

Work down the list:

1. Are you on Docker Desktop? `docker info | grep -i "operating system"` — a native Linux engine cannot work.
2. Is isolation `shared`? Run `asylum config` or check `agents.claude.config`.
3. Does the container see lock files? `ls ~/.claude/ide/` inside the container should show `*.lock` entries.
4. Is the IDE open on the project directory itself, at the same path you ran `asylum` from?
5. On JetBrains, restart the IDE and retry.
6. Is the variable set? `env | grep CLAUDE_CODE_IDE_HOST_OVERRIDE` inside the container. If it's missing, the container predates the feature — restart it.

## Security note

Connecting to an IDE gives the sandboxed agent access to the tools that IDE's plugin exposes, which is a channel out of the container to your host editor. Nothing connects on its own: it takes an explicit `/ide` or an explicit `autoConnectIde` opt-in, and the tools are those of your own IDE on your own machine. See the [Security Model](security-model.md) for the boundary Asylum does and does not provide.
