# Fix `asylum run` so commands find tools on PATH

## Why

`asylum run <cmd>` execs the command bare into the container:

```go
// internal/container/container.go
case ModeCommand:
    args = append(args, opts.ExtraArgs...)   // docker exec -it <c> claude auth login
```

A bare `docker exec` runs neither the entrypoint nor a shell rc, so its `PATH`
is only the Docker default (`/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin`).
Every tool asylum installs lands outside that default:

- `claude` → `~/.local/bin/claude` (native installer)
- `node`, `npm`, `gemini`, `codex` → fnm-managed Node bin (added via `eval "$(fnm env)"`)
- `java`, `python` → mise shims

Those directories reach `PATH` only through `~/.zshrc` / `~/.bashrc` (per-shell)
or the entrypoint's one-time boot export — neither of which a bare `docker exec`
touches. There is no Docker `ENV PATH` that bakes them in.

Result: `asylum run claude auth login` fails with
`exec: "claude": executable file not found in $PATH`. The same holds for
`asylum run npm ...`, `asylum run node ...`, and any other kit/agent tool.
Only commands already on the default PATH (e.g. `asylum run ls -la`) work.

Agent mode and shell mode are unaffected because both route through a shell:
agent mode wraps in `wrapZsh` (`zsh -c "source ~/.zshrc && exec …"`) and shell
mode execs an interactive `/bin/zsh`. `ModeCommand` is the only exec path that
never touches a shell, which is why it is uniquely broken.

## What Changes

- `ModeCommand` execs its command through the same login-shell wrapper used by
  agent mode, so `~/.zshrc` (fnm env, mise, `~/.local/bin`) is sourced before the
  command runs. `asylum run <cmd>` then finds the same tools available in
  `asylum shell`.
- The command's arguments are shell-quoted before being joined into the wrapper
  string, so arguments containing spaces or shell metacharacters are preserved.
- `exec` in the wrapper replaces the shell with the target binary, so the
  command's exit code and signal handling pass through unchanged, and the
  interactive `claude()` skills-injection function is bypassed (correct for raw
  subcommands like `claude auth login`).

## Impact

- Affected spec: `container-exec` (the "Exec into running container for run mode"
  requirement).
- Affected code: `internal/container/container.go` (`ExecArgs`, `ModeCommand`
  branch), reusing the existing `wrapZsh` helper from `internal/agent`.
- Behaviour change: `asylum run <cmd>` now runs inside a login shell. This adds
  the small `~/.zshrc` (oh-my-zsh) startup cost already paid by agent and shell
  modes, and means shell builtins/aliases/functions from the rc are in scope.
- No config, flags, or agent surface change.
