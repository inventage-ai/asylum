## Why

The current CLI parser uses heuristics to decide where asylum flags end and agent/command args begin: the first unrecognized flag triggers passthrough of all remaining args. This causes problems:

- **Typos pass silently.** A misspelled asylum flag (e.g. `--rebuld`) gets forwarded to the agent, which also doesn't recognize it, producing a confusing error far from the source.
- **Flag collision risk.** Adding a new asylum flag could silently steal args that users were passing through to their agent.
- **Running commands is implicit.** `asylum openspec init` works because `openspec` isn't a keyword, triggering command mode — but this isn't obvious or discoverable.

## What Changes

Introduce strict argument parsing with explicit boundaries:

1. **Unknown flags become errors** instead of triggering passthrough.
2. **`--` separator** is the only way to pass extra args to the agent.
3. **`run` subcommand** for executing arbitrary commands in the container (replaces the implicit positional-triggers-command behavior).

### New CLI grammar

```
asylum [flags] [subcommand] [-- agent-args]

Subcommands:
  shell [--admin]       Interactive shell
  ssh-init              Initialize SSH directory
  run <cmd> [args...]   Run command in container

No subcommand → launch agent
```

### Examples

```
asylum                            Launch default agent
asylum -a gemini                  Launch gemini
asylum -a gemini -- --verbose     Launch gemini with extra arg
asylum shell                      Interactive shell
asylum shell --admin              Admin shell
asylum ssh-init                   SSH init
asylum run python test.py -v      Run command in container
asylum run -- python test.py      Same (-- optional after run)
asylum --bogus                    Error: unknown flag "--bogus"
```

## Capabilities

### New Capabilities

- `run` subcommand: explicit way to execute arbitrary commands in the container

### Modified Capabilities

- `cli-parsing`: strict flag handling, unknown flags rejected, `--` required for agent passthrough
- `help-text`: updated usage to reflect new grammar

## Impact

- **Breaking change**: users passing agent flags without `--` must add the separator. Passthrough without `--` no longer works.
- Modifies `cmd/asylum/main.go` only (parseArgs, resolveMode, printUsage).
- No config, image, or container changes.
