## Why

iTerm2 ships a Claude Code integration that paints session state into the status bar — a colored dot for working/idle and a line of detail naming the current tool. It is wired up by iTerm2 itself, as ten hook entries in `~/.claude/settings.json` all pointing at `~/.config/iterm2/cc-status`.

Inside an asylum container that integration is not merely absent, it is actively broken. Under the default `shared` agent config isolation the container reads the host's real `~/.claude/settings.json`, so all ten hooks fire — and `cc-status` is a macOS Mach-O binary. Bind-mounting it into the container was confirmed to fail with `exec format error` (exit 126); without the mount the path does not exist at all. Either way every iTerm2 user running asylum takes a failing hook on `PreToolUse`, `PostToolUse`, `Stop`, `SessionStart` and six other events, for every tool call of every session.

The integration cannot be reimplemented in the container. `cc-status` is a thin adapter that reads the hook payload on stdin, reads `TERM_SESSION_ID`, and execs `it2 set-status --session … --status … --dot-color … --detail …`. `it2` reaches iTerm2 through its API on a host Unix socket. That socket is visible through a bind mount but not connectable from the container — `connect()` returns `ECONNREFUSED`, because a Unix socket needs the listener in the same kernel and iTerm2's listener is on the macOS side of the VM boundary. There are no terminal escape sequences anywhere in `cc-status`, so there is nothing to emit on the tty instead.

Asylum already has the machinery to close this: the host broker, which serves kit-contributed routes to a container over an authenticated channel. This change adds one narrow route that runs the host's own `cc-status` on the container's behalf.

## What Changes

- A new opt-in `iterm` kit installs a shim in the container and bind-mounts it over `~/.config/iterm2/cc-status`, so the hook path iTerm2 already wrote into `settings.json` resolves to the shim inside the container and to the real binary on the host. No user edit to `settings.json`, and host sessions are untouched.
- The shim forwards the hook payload on stdin to a new broker route, `/iterm-status`. It always exits `0`, so a broker that is slow, absent, or failing can never block a tool call.
- The route handler runs the host's `cc-status` with the request body as stdin and the session's identity in its environment. The handler passes **no argument** from the container to the host process — everything the container supplies arrives as stdin bytes and one opaque token.
- The broker gains a sealed session envelope. Asylum seals the terminal session's identity with a host-only key at exec time and injects only the ciphertext into the container; the handler opens it. The container cannot read, forge, or replace the session identity it addresses.
- `container-exec` gains per-exec environment injection so the envelope belongs to the terminal session that started it, not to the first session that happened to start the container.
- A pre-existing bug is fixed along the way: accepting a newly-offered opt-in kit at the kit-sync prompt wrote the kit's authored (commented-out) snippet verbatim, so the kit stayed disabled and the prompt appeared to do nothing. Found while enabling this kit; it affects `rtk` and `cx` equally.

Deliberately out of scope: exposing `it2` itself to the container in any form, a general host-command route, and support for agents other than Claude Code. All three are argued against in design.md.

## Capabilities

### New Capabilities
- `iterm-status-kit`: an opt-in kit that lets a containerized Claude Code session drive the iTerm2 status bar of the terminal it is running in, by forwarding hook payloads to the host's own `cc-status` binary through the broker.

### Modified Capabilities
- `host-broker`: gains a sealed, per-session envelope so a route handler can learn host-side facts about the calling session that the container is not trusted to assert.
- `container-exec`: gains per-exec environment injection, so values that are properties of the terminal session rather than of the container are set at `docker exec` time.

## Impact

- `internal/kit/iterm.go` (new) — kit registration, Dockerfile snippet for the shim, mount over the hook path, route registration, rules snippet.
- `internal/broker/` — envelope seal/open and the host-only key file.
- `internal/kit/snippet.go` (new) — snippet comment/uncomment helpers moved out of `internal/firstrun` so the kit-sync path can share them.
- `internal/config/kitsync.go` — normalise an accepted kit's snippet to active form.
- `internal/container/container.go` — `ExecArgs` gains per-exec `-e` arguments (`container.go:781`).
- `cmd/asylum/main.go` — seal the envelope before exec.
- `assets/asylum-reference.md`, `docs/`, `CHANGELOG.md`.
- No new dependencies; sealing uses `crypto/aes` and `crypto/cipher` from the standard library.
- Inert on any host without iTerm2: the kit is opt-in, and with no iTerm2 session on the host no envelope is injected and the shim no-ops.
