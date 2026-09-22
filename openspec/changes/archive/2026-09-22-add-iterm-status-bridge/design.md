## Context

See proposal.md — Why.

What `cc-status` does was established by inspecting the binary and by running it. It is a Swift Mach-O universal binary (x86_64 + arm64) that:

1. reads a Claude Code hook payload on stdin and pulls out `hook_event_name`, `tool_name`, `tool_input` and `session_id`;
2. maps the event to a state and color — `PreToolUse`/`PostToolUse` to working `#ff9500`, `Notification` to `#5f87ff`, `Stop` to idle `#00d75f`, and so on — and formats a per-tool detail string with cases for Bash, Read, Edit, Write, Grep, Glob, Task, WebFetch, AskUserQuestion and ExitPlanMode;
3. reads `TERM_SESSION_ID` from its environment and strips everything through the colon to get the bare UUID `it2` wants;
4. execs `it2 set-status --session … --status … --dot-color … --text-color … --detail …`, optionally with `--background-tasks`.

Three findings shape the design.

**The container cannot reach iTerm2 directly.** `it2 --help` describes itself as controlling iTerm2 "using its API", and `it2 auth` offers a `cookie` subcommand. That API is served on a host Unix socket. Bind-mounting `~/.config/iterm2` makes the socket node visible in the container, but `connect()` on it returns `ECONNREFUSED` — virtiofs passes the inode, not the connection, and the listener lives in the host kernel. Confirmed empirically.

**There is no escape-sequence fallback.** `cc-status` contains no OSC or `1337` strings. Everything travels over the API socket, so no shim writing to the tty can substitute.

**The broker's environment is frozen at spawn.** `EnsureBroker` returns early when a live broker answers, and the `host-broker` spec already states that the broker's lifetime is the container's, not any session's. The broker is spawned by whichever session first needed it, `setsid`s away, and keeps that session's environment forever. Any session-scoped value read from the broker's own environment is correct for the first tab and silently wrong for every later one. iTerm2 assigns a distinct `TERM_SESSION_ID` per tab — confirmed — so this is the common case, not an edge case.

## Goals / Non-Goals

**Goals:**

- The iTerm2 status bar reflects a containerized Claude Code session as faithfully as it reflects a host one.
- Stop the ten hooks that fail on every session today for iTerm2 users under `shared` config isolation.
- Correct with any number of tabs attached to one container, in any open/close order.
- The container cannot address a terminal session other than its own.
- Inert and zero-cost on hosts without iTerm2.

**Non-Goals:**

- Exposing `it2` to the container, under any allowlist.
- A general "run a host command" broker route.
- Agents other than Claude Code. The hook contract is Claude's; nothing here generalizes until a second agent has one.
- Reimplementing the event-to-color mapping or the detail formatting. That logic is iTerm2's and should stay theirs.
- Native Linux engines. There is no iTerm2 to talk to.

## Decisions

**Run the host's `cc-status`, not `it2`.**

This is the whole security argument for the change. `it2` offers `session run "…" --all` and `session send "…"`, which type into every terminal the user has open — proxying it would hand the container host command execution and would be a complete sandbox bypass, not a leak. Any subcommand allowlist over `it2` would be one iTerm2 release away from silently regaining that reach. `cc-status` has no such surface; its entire output is a colored dot and a line of text.

The containment property is that **no part of the handler's argv comes from the container**. The handler execs a fixed absolute path with a fixed argument list. Container-supplied bytes arrive only as stdin and as one sealed token, and `cc-status` alone decides what becomes an `it2` flag. There is no string the container can send that becomes a new argument, because it never reaches the argument layer.

Running the real binary has a second benefit: `cc-status` persists per-session background-task counts on the host, and an iTerm2 update that adds an event or recolors a state works without an asylum release.

**Shadow-mount the shim over the hook path.**

Under `shared` isolation the container and the host read the same `settings.json`, so the hook command cannot be repointed at a container-only shim without breaking host sessions. Mounting asylum's shim over `~/.config/iterm2/cc-status` gives the same path two meanings — real binary on the host, shim in the container — and needs no user edit. iTerm2 maintains that path itself as a symlink into its app bundle, which is also what it writes into `settings.json`.

Claude Code merges hooks from all settings sources rather than overriding them, so contributing a container-side hook would add a second hook without silencing the broken one. Shadowing is the only approach that removes the failure.

*Mechanics.* `CredentialMount.Content` writes staged files `0600`, which is not executable. Rather than widen that struct, the kit's `MountFunc` writes its own shim file `0755` into the per-container staging directory and returns it as `HostPath`. Worth revisiting if a second kit needs an executable generated file.

**Seal the session identity rather than passing it in the clear.**

The container must supply something per-request, since one broker serves many tabs. Passing the terminal's own id would make `--session` an unvalidated string from inside the sandbox. Sealing means the container holds an opaque blob it can neither read nor forge: it can address the session that issued its token and nothing else, and a blob that fails to open means no host process runs at all.

The envelope also solves a problem the plain variable does not. If iTerm2's API authentication rides in the session environment (`it2 auth cookie` suggests a cookie plus key), then the broker's frozen environment carries the *first* tab's credential and later sessions fail to authenticate — the same staleness bug, one layer down. Sealing the credential alongside the session id refreshes it per session and keeps it out of the container. The envelope is a JSON object so its contents can grow without changing the route.

AES-256-GCM with a random nonce, base64url-encoded. The key is 32 random bytes at `~/.asylum/session.key`, `0600`, created on first use. Both the CLI and a long-running broker read it lazily, so a broker spawned days ago opens an envelope sealed now with no handshake.

*Alternative considered: a token registry.* asylum writes `<random token> → session id` as a file in a host-only directory and the broker reads it. Rejected on two counts. asylum `syscall.Exec`s into `docker` and is replaced by it, so nothing can delete the entry when the session ends, leaving stale files that need a sweep. And a container-supplied token used as a filename is untrusted input building a path, needing a validation guard that sealing does not.

*Key placement.* `~/.asylum/projects/<cname>` is the obvious per-container directory but is exactly wrong — on a native Linux engine it is bind-mounted into the container as `/run/asylum` for the broker socket (`container.go:129`). The key sits at the `~/.asylum` top level, outside anything asylum mounts by default. A user who mounts `~/.asylum` wholesale defeats this; documented, not defended against.

**Inject the envelope per exec, not per run.**

`docker run -d` happens once per container; `docker exec` happens once per session. The envelope describes a terminal session, so it belongs on the exec. `ExecArgs` takes no `-e` arguments today, so this adds them.

This also gives the feature a free off switch. With no iTerm2 session in asylum's own environment — Terminal.app, a script, CI — nothing is sealed, the variable is absent, and the shim returns immediately. The integration disables itself in exactly the cases where it could not have worked.

**The shim always exits `0`.**

Claude Code reads hook exit codes, and `2` blocks the action. A status indicator must never be able to block a tool call, so the shim discards curl's status and exits `0` unconditionally, with `--max-time` set low so an unreachable broker costs a bounded delay rather than a hang.

**The handler runs `cc-status` synchronously.**

`browser-open` uses `Start()` and returns immediately, which is right for opening a URL. Here it would let two rapid hooks race and land out of order, painting idle over working. Claude Code fires hooks sequentially, so running to completion preserves the order the events actually happened in. The cost is that hook latency includes a process spawn — which is exactly what a host session already pays for the same feature.

**Opt-in kit.**

The kit is useful only on a macOS host running iTerm2 with its Claude Code integration enabled. `TierOptIn` keeps it out of every other project's config. It is not on-by-default like `browser-open` because it reaches further — a host process spawn per hook rather than per explicit user action.

## Risks / Trade-offs

**A detached caller can drive the iTerm2 API — verified.** A `cc-status` run from a process that had reparented to launchd, with no controlling terminal, updated the named session's status bar with no permission dialog. `it2 auth cookie` exists but the session environment carries no `ITERM2_COOKIE`, so the envelope needs to carry only the session id. `--session` targeting was confirmed separately: `it2` takes the bare UUID, which `cc-status` derives from `TERM_SESSION_ID` itself — another thing gained by running their binary rather than reimplementing it.

**The variable is `TERM_SESSION_ID`, not `ITERM_SESSION_ID` — found the hard way.** The first implementation overrode `ITERM_SESSION_ID`, which `cc-status` never reads, so the broker's inherited `TERM_SESSION_ID` won and every session painted whichever tab spawned the broker. The session that started the container looked correct throughout, which hid the bug. Reading iTerm2's source (GPL, `gnachman/iTerm2`, `cc-status/Sources/cc-status/main.swift`) settled in minutes what several rounds of black-box probing had not. → The handler now overrides both variables, so a future `cc-status` that switches to the other name cannot resurrect the same stale-inheritance failure. The lesson generalizes: when a handler overrides a session-scoped variable inherited by a long-lived process, overriding the wrong name fails silently and looks like correct behavior from the spawning session.

**Whether Claude Code's hook subprocesses inherit the `docker exec` environment is assumed, not verified.** The whole envelope delivery depends on it. → Task 1.2 verifies it directly before the kit is written.

**`it2` may not be on the broker's `PATH`.** `cc-status` execs it via `/usr/bin/env`, so it resolves against whatever the broker inherited. If iTerm2's shell integration is what puts the utilities directory on `PATH`, a broker spawned from a different context will not find it and `cc-status` prints `failed to run it2`. → The handler prepends `/Applications/iTerm.app/Contents/Resources/utilities` — `it2`'s confirmed location — to the child's `PATH` rather than trusting inheritance.

**The hook path is an assumption about iTerm2's behavior.** If a future iTerm2 writes a different command into `settings.json`, the shadow mount covers a path nothing calls and the integration silently stops. → Accepted for v1. The failure is inert rather than harmful, and the fix is a one-line path change.

**The container controls the detail text.** That is the feature — the status bar exists to show what the agent is doing. A compromised agent could paint misleading text, and `cc-status` formats `AskUserQuestion` and `ExitPlanMode` events, so permission-prompt-shaped strings are reachable. → Accepted. It spoofs a status indicator, not an actual prompt, and no amount of validation distinguishes honest detail text from dishonest. The body is size-capped and otherwise passed through unexamined, which keeps the handler honest about what it is.

**A hook fires on every tool call, so this is a hot path.** Each one costs an HTTP round trip to the host plus a process spawn. → The round trip is loopback and the spawn is what the host already pays. The bounded `--max-time` caps the worst case, and the kit is opt-in, so a user who finds the cost unacceptable turns it off.

**The envelope has no expiry.** A container that captured a blob can keep addressing that session for as long as the key lives. → Accepted for v1. The worst case is replaying a handle to a session it was legitimately given. A timestamp inside the envelope adds expiry later without changing the route or the shim.
