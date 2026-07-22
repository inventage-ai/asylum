# Design

## Context

Three exec modes share `container.ExecArgs`:

| Mode          | Exec form                                    | Sources rc? | PATH ok? |
|---------------|----------------------------------------------|-------------|----------|
| `ModeAgent`   | `zsh -c "source ~/.zshrc && exec claude …"`  | yes         | ✓        |
| `ModeShell`   | `/bin/zsh` (interactive → sources `.zshrc`)  | yes         | ✓        |
| `ModeCommand` | `claude auth login` (bare exec)              | **no**      | **✗**    |

`ModeAgent` gets its shell via `wrapZsh` in `internal/agent/agent.go`:

```go
func wrapZsh(cmd string) []string {
    return []string{"zsh", "-c", "source ~/.zshrc && exec " + cmd}
}
```

`ModeCommand` bypasses this and execs the raw argv. The fix is to route
`ModeCommand` through the same wrapper.

## Goals / Non-Goals

- **Goal:** `asylum run <cmd>` resolves tools exactly as `asylum shell` does.
- **Goal:** Preserve exit code, signals, stdin/stdout/stderr, and argument
  boundaries (quoting).
- **Non-goal:** Baking a static `ENV PATH` into the image. It would fix only
  `~/.local/bin` tools; fnm/mise tools need `eval`-time setup that a static ENV
  cannot provide.
- **Non-goal:** Changing the interactive `claude()` wrapper function or
  skills-dir injection.

## Decision: wrap `ModeCommand` in a login shell

Reuse the login-shell wrapping pattern in the `ModeCommand` branch of
`ExecArgs`, shell-quoting the user's arguments before joining:

```go
case ModeCommand:
    cmd := strings.Join(term.ShellQuoteArgs(opts.ExtraArgs), " ")
    args = append(args, "zsh", "-c", "source ~/.zshrc && exec "+cmd)
```

`wrapZsh` currently lives in `internal/agent` and is unexported. Rather than
export it across packages, `container` builds the same `zsh -c "source … && exec …"`
argv inline — `container` already imports `term` for quoting, so this adds no new
dependency. (If a shared helper is preferred, it can move to a small internal
package, but a two-line inline form keeps the change minimal per the project's
"less code is better" rule.)

### Why `exec`

`exec` replaces the shell process with the target binary, so:
- the command's exit status is what asylum sees (no shell wrapper masking it);
- signals (SIGINT/SIGTERM/SIGHUP already forwarded by the CLI) reach the command
  directly;
- the interactive `claude()` shell function is bypassed — `exec` only runs
  external binaries — which is the desired behaviour for raw subcommands like
  `claude auth login` (no unwanted `--add-dir /opt/asylum-skills` injection).

### Quoting

`term.ShellQuoteArgs` is already used by the agent `Command` implementations to
protect passthrough arguments. Applying it here keeps argument boundaries intact
(e.g. `asylum run node -e "console.log('a b')"`).

## Risks / Trade-offs

- **`.zshrc` startup cost:** sourcing oh-my-zsh adds latency per `asylum run`.
  Accepted — agent and shell modes already pay it, and consistency is worth more
  than shaving a fraction of a second off `run`.
- **rc side effects in scope:** aliases/functions/`setopt` from `~/.zshrc` now
  apply to `run` commands. This matches `asylum shell` semantics and is the
  intended "run it like I would in a shell" behaviour. `exec` still bypasses
  shell *functions* for the top-level command, so a `claude()` function does not
  shadow the `claude` binary.

## Alternatives Considered

1. **Static `ENV PATH` in the Dockerfile** — fixes `~/.local/bin` (claude) only;
   fnm/mise tools remain broken because their environment is set by `eval` at
   shell init, not a fixed path. Rejected: partial fix.
2. **`docker exec -e PATH=$(cat /tmp/asylum-path)`** — the entrypoint writes the
   resolved PATH to `/tmp/asylum-path`. Injecting it fixes PATH but not other
   shell-env setup (fnm env exports, mise shims/vars). Rejected: incomplete and
   more moving parts than sourcing the rc.
