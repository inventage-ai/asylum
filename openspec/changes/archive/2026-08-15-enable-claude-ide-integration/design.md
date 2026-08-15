## Context

See proposal.md - Why.

Three things must line up for `/ide` to work inside a container. Investigation against the Claude Code binary (`2.1.227`) and a live container found that two of them already hold, which is what makes this change one line.

**Lock file visibility.** The CLI scans `$CLAUDE_CONFIG_DIR/ide/*.lock`. Under the default `shared` agent config isolation, `ResolveConfigDir` bind-mounts the host's real `~/.claude` into the container, so lock files written by a host IDE are already visible. Under `isolated` and `project` isolation the container gets an asylum-managed directory and sees nothing.

**Workspace matching.** The CLI compares each lock file's `workspaceFolders` against its own cwd and requires equality or containment. Asylum mounts the project at its real host path, so the comparison succeeds unchanged.

**Reachability.** The CLI resolves the host to dial through a function that returns `127.0.0.1` unless `CLAUDE_CODE_IDE_HOST_OVERRIDE` is set. Inside a container that address is the container's own loopback, where nothing listens. This is the only gap.

Verified empirically from inside a container against a host IntelliJ instance: a WebSocket connection to `host.docker.internal:<port>` carrying the `mcp` subprotocol and the lock file's `X-Claude-Code-Ide-Authorization` header completed an MCP `initialize` handshake and returned `serverInfo.name = "Claude Code JetBrains Plugin"`. The IDE server applies no Host-header or DNS-rebinding check that would reject the non-loopback `Host` value.

There is also a liveness check to be aware of: the CLI drops lock files whose `pid` is not running, which would reject every host IDE pid. It is gated on the terminal being a recognized IDE terminal (`TERM_PROGRAM` set to a VS Code or JetBrains value), which asylum does not forward into the container, so the check does not run. See Risks.

## Goals / Non-Goals

**Goals:**

- `/ide` connects on the configuration the majority of users run: Docker Desktop, `shared` agent config isolation.
- Zero runtime cost. No process, no mount, no broker route, no image change.
- Cheap to remove if the feature turns out to be unused.

**Non-Goals:**

- Native Linux engines. `host.docker.internal` resolves to the bridge gateway there and cannot reach a host process bound to `127.0.0.1`.
- `isolated` and `project` isolation modes. These need lock files bridged into the container, which is a materially larger change.
- Auto-connecting on session start. The CLI's `autoConnectIde` setting already covers this and persists in the mounted config directory, so it is the user's choice and needs nothing from asylum.
- Agents other than Claude. Codex has its own IDE integration with its own discovery mechanism.

## Decisions

**Set the variable in `Claude.EnvVars()` rather than building a kit.**

`EnvVars()` is the existing mechanism for agent-scoped container environment and it is already wired through `coreEnvVars`. A kit would add a config entry, a rules snippet, and an off switch, which is the right shape for a feature with a meaningful cost or risk. This one is a single environment variable that does nothing until the user explicitly runs `/ide` or opts into `autoConnectIde`, so the gate would guard an action the user must take deliberately anyway. Revisit if the exposure calculus changes.

**Set it unconditionally rather than detecting the engine.**

The correct condition is "the engine routes `host.docker.internal` to host loopback", which asylum knows indirectly: the broker falls back to TCP over `host.docker.internal` exactly when it is not on a native Linux engine. Threading that signal into `EnvVars()` would mean changing the `Agent` interface, since `EnvVars()` takes no arguments today. Not worth it: on a native Linux engine the override points at an unreachable gateway, and the unmodified default points at an unreachable container loopback. Both fail identically, so the variable cannot regress anything.

**Alternative considered: a lock-file watcher plus `socat` forwarding.** A process in the container watching `~/.claude/ide` and starting `socat TCP-LISTEN:<port>,bind=127.0.0.1 ...` per lock file would be agent-agnostic and would work on native Linux via a Unix socket relayed through the existing broker. Rejected for now as a large amount of machinery for an unvalidated feature. It remains the natural follow-on if `/ide` proves useful and the Linux or `isolated` cases need covering, and nothing in this change blocks it: the override variable would simply be dropped in favor of forwarding on `127.0.0.1`.

**Alternative considered: a host-side watcher in the broker.** Same forwarding behavior, but written in Go on the host where it is testable, and able to rewrite mirrored lock files (port, pid) rather than pass them through. Better than the in-container variant if forwarding is ever needed. Also deferred.

## Risks / Trade-offs

**The override is an undocumented internal of the Claude CLI and could be removed or renamed in any release.** → The failure mode is that `/ide` stops working, which is today's behavior, so the blast radius is the feature itself. The change is one map entry and is trivially removable.

**The pid liveness check is dormant, not absent.** If asylum ever forwards `TERM_PROGRAM` into the container, or Anthropic ungates the check, every host IDE lock file is filtered out and `/ide` silently shows an empty list. → Accepted. Detectable at that point and fixable by mirroring lock files with a rewritten pid, which is part of the deferred watcher work.

**A silent empty `/ide` list on unsupported configurations.** A user on a native Linux engine or in `isolated` isolation gets no error explaining why. → Mitigated by documenting both preconditions in `asylum-reference.md`, which is the reference the in-container agent is instructed to read.

**Connecting hands the sandboxed agent the host IDE's toolset.** The verified handshake advertised `tools.listChanged` from the host plugin, so this is a real channel out of the sandbox. → Accepted deliberately: nothing connects until the user runs `/ide` or enables `autoConnectIde`, both explicit acts, and the tools are those of the user's own IDE on their own machine.
