## Approach

Replace the heuristic passthrough parser with a strict two-phase parse:

1. **Phase 1**: Consume known asylum flags and subcommands. Reject unknown flags with an error.
2. **Phase 2**: If `--` is encountered, collect everything after it as extra args.

The `run` subcommand acts as its own boundary — everything after `run` is the command to execute, with `--` optional.

## Parser Logic

```
parseArgs(args) → (flags, subcommand, extraArgs, error)

for each arg:
  "--"           → collect rest as extraArgs, stop
  known flag     → consume (with value if needed)
  unknown flag   → error: unknown flag %q
  "shell"        → subcommand = shell, consume --admin if present
  "ssh-init"     → subcommand = ssh-init
  "run"          → subcommand = run, collect rest as extraArgs (skip leading --)
  other          → error: unexpected argument %q
```

## Decision: `run` swallows all remaining args

Everything after `run` is the command — no flag parsing. This means `asylum run --help` runs `--help` inside the container, not asylum's help. Asylum's own flags must come before `run`:

```
asylum -p 8080 run python server.py     ✓ port flag before run
asylum run -p 8080 python server.py     ✗ -p treated as part of command
```

This matches `docker run`, `kubectl exec`, and `ssh` conventions.

## Decision: `shell --admin` parsed explicitly

Instead of relying on passthrough, `shell` checks for `--admin` as a known sub-flag. Unknown flags after `shell` produce an error.

## Return Value Change

`parseArgs` currently returns `(cliFlags, positional, passthrough, error)`. The new signature returns `(cliFlags, subcommand, extraArgs, error)` where subcommand is a string (`""`, `"shell"`, `"ssh-init"`, `"run"`) and extraArgs replaces both positional and passthrough.

This simplifies `resolveMode` — it receives the subcommand directly instead of inferring it from positional args.
