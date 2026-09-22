# iTerm2 Kit

Report session state to the iTerm2 status bar from inside the container.

**Activation: Opt-in** — macOS with iTerm2 only.

## What It Does

iTerm2 ships a Claude Code integration that paints session state into the status bar: an orange dot while a tool runs, green when idle, and a line of detail naming the tool. iTerm2 wires it up itself, as ten hook entries in `~/.claude/settings.json` all pointing at `~/.config/iterm2/cc-status`.

Inside a container that integration is not just missing, it is broken. Under `shared` agent config isolation the container reads the host's real `settings.json`, so all ten hooks fire — and `cc-status` is a macOS binary that cannot execute on Linux. Every tool call of every session takes a failing hook.

This kit fixes both halves: the status bar works from inside the sandbox, and the failing hooks stop failing.

```yaml
kits:
  iterm:
```

## Prerequisite: install iTerm2's integration first

iTerm2 offers its Claude Code onboarding when a session's foreground job is named exactly `claude`. Under Asylum that job is `docker`, so **the offer never appears** — even though everything else works once the hooks exist.

If you have never accepted that onboarding, install it by hand from **iTerm2 → Install Claude Code Integration**. Without it there are no hooks in `~/.claude/settings.json`, nothing calls the shim, and this kit does nothing at all.

## How It Works

The kit bind-mounts a small shim **over** `~/.config/iterm2/cc-status` inside the container. The same path then means two different things — the real binary on the host, the shim in the container — so host sessions are untouched and your `settings.json` is never edited.

The shim forwards the hook payload to the **host broker**, which runs the host's own `cc-status` with that payload on stdin. That binary talks to iTerm2 over its API socket, which a container cannot reach: the socket can be bind-mounted, but `connect()` fails because the listener lives in the host kernel.

The shim always exits `0` and gives up after two seconds. A broker that is slow, missing, or failing can never block or fail a tool call.

## Trust Boundary

The container drives `cc-status` and nothing else. It never reaches `it2`, iTerm2's general-purpose CLI, which offers `it2 session run` and `it2 send` — commands that execute in **every terminal you have open**. Exposing that would be a complete sandbox bypass rather than a leak, and no subcommand allowlist would stay safe across iTerm2 releases.

What the container can do is set a colored dot and a line of text. Nothing it sends becomes a command-line argument: the executable path is fixed, the argument list is empty, and the payload reaches the binary only on standard input.

### Which terminal it can address

One container serves many tabs, so each session has to say which terminal its updates belong to. It is not trusted to name one.

At session start, Asylum seals the terminal's identity with a key readable only by your account on the host, stored outside every path it mounts into a container. The container receives only the ciphertext. It cannot read the identity, cannot forge a different one, and cannot alter the one it was given — an envelope that fails to open means no host process runs at all.

The envelope travels on the `docker exec`, not on the container, because the broker outlives every individual session. A value baked in at container creation would describe whichever tab happened to start it, and would be silently wrong for every tab after that.

The detail text itself is container-controlled, because that is the feature. A compromised agent could paint misleading text in your status bar. It is a spoof of a status indicator, not of a real prompt.

## When It Does Nothing

The kit disables itself rather than erroring:

- **Not an iTerm2 session** — started from Terminal.app, a script, or CI. No envelope is sealed and the shim returns immediately.
- **iTerm2's integration not installed on the host** — no binary to shadow, so no mount is made and the route reports the integration unavailable.
- **Not macOS** — same as above.

## Debugging It

The status is cleared whenever a shell prompt starts — iTerm2 does this in `PTYSession.screenPromptDidStartAtLine`. An agent session is a single long-running command, so nothing clears it while the agent works. But it does mean **you cannot test this from an idle shell**: run `cc-status` by hand and the prompt that follows wipes the status before you see it.

To probe by hand, run it from a tab that has a long-running command in the foreground, or watch the target tab from another one.

`cc-status` reads **`TERM_SESSION_ID`**, not `ITERM_SESSION_ID`, and strips everything through the colon to get the bare UUID `it2` wants. It exits `0` silently when that variable is missing or has no colon, so a wrong or absent value looks exactly like success.

## Cost

A hook fires on every tool call, so each one costs a loopback HTTP round trip plus a process spawn on the host. The spawn is what a host session already pays for the same feature. If that is not worth it to you, leave the kit off.
