## Context

`assets/entrypoint.core` installs a `claude()` shell function that prepends `--add-dir /opt/asylum-skills` whenever the shared skills directory is non-empty. `--add-dir <directories...>` is a **variadic** flag: it consumes every following token until it hits a `-`-prefixed argument or end of input. Placing it before `"$@"` means any positional token — a subcommand (`mcp`, `doctor`, `update`, …) or a bare prompt — is consumed as a directory path, leaving no subcommand and breaking the invocation.

The same latent trap exists in `internal/agent/claude.go` `Command()`, which appends `extraArgs` after `--add-dir`; it is masked today only because asylum passthrough args are almost always flags.

## Goals / Non-Goals

**Goals:**
- `claude <subcommand>` works inside the sandbox for all current and future subcommands.
- Bare positional prompts (`claude "…"`) are no longer swallowed.
- Interactive/session invocations still auto-discover kit skills.

**Non-Goals:**
- Auto-loading kit skills on positional-prompt one-shots (`claude "fix bug"`). These do not work today (the variadic swallows them), and passing them through cleanly is strictly better; users who want skills use bare `claude` or `claude -p "…"`.
- Replacing `--add-dir` with a settings/env-based skills mechanism (would risk polluting the host `~/.claude` in shared mode — the hygiene property preserved by #24/#25).

## Decisions

**Guard the wrapper on invocation shape, not a subcommand list.** Inject `--add-dir` only when `$# -eq 0` or the first argument begins with `-`:

```sh
claude() {
    if [ -d /opt/asylum-skills/.claude/skills ] && [ -n "$(ls -A /opt/asylum-skills/.claude/skills 2>/dev/null)" ] \
       && { [ $# -eq 0 ] || [ "${1#-}" != "$1" ]; }; then
        command claude --add-dir /opt/asylum-skills "$@"
    else
        command claude "$@"
    fi
}
```

- _Alternative — blocklist known subcommands:_ rejected. The list (`mcp`, `doctor`, `update`, `plugin`, `install`, `agents`, `project`, `setup-token`, `ultrareview`, …) rots as Claude adds subcommands; a missed entry silently reintroduces the bug.
- When the flag is added, the next token is guaranteed to be nothing or a `-flag`, so the variadic can never swallow user input — front placement stays safe.

**Reorder `--add-dir` after `extraArgs` in `Command()`.** Making the skills dir the last token keeps the variadic from swallowing a positional passthrough prompt, closing the same trap on the primary path.

## Risks / Trade-offs

- **Positional-prompt one-shots lose auto skill-loading** → Acceptable: they are broken today, so this is a net fix; the common secondary-shell forms (`claude`, `claude -c`, `claude -p "…"`) keep skills.
- **`${1#-}` shell portability** → Works in both bash and zsh (POSIX parameter expansion); the wrapper already targets both `.zshrc` and `.bashrc`.
