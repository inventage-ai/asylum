## Why

The entrypoint's `claude()` shell wrapper unconditionally prepends `--add-dir /opt/asylum-skills` to every invocation when kit skills are present. Because `--add-dir <directories...>` is variadic, it swallows any following positional token — so `claude mcp`, `claude doctor`, `claude update`, and even bare positional prompts like `claude "fix the bug"` break inside the sandbox. The wrapper is meant to help interactive sessions discover kit skills, but it silently breaks Claude's subcommands instead.

## What Changes

- The `claude()` entrypoint wrapper only injects `--add-dir /opt/asylum-skills` on **session invocations** — when there are no arguments, or the first argument is a flag (`-…`). Subcommands and positional prompts pass through to `command claude` untouched.
- This fixes all `claude <subcommand>` forms (`mcp`, `doctor`, `update`, `plugin`, `install`, `agents`, `project`, `setup-token`, `ultrareview`, …) and is future-proof: new subcommands work without a maintained list.
- Harden the primary session path: reorder `--add-dir /opt/asylum-skills` to come **after** `extraArgs` in `internal/agent/claude.go` `Command()`, so a positional prompt passed via `asylum -- "…"` can't be swallowed by the same variadic flag.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `kit-skills-delivery`: The "Interactive claude invocations pick up kit skills" requirement is refined so the wrapper only adds `--add-dir` on session invocations (no args, or first arg is a flag), leaving subcommands and positional prompts untouched. The "Agent launch passes --add-dir" requirement is tightened to place `--add-dir` after any passthrough args.

## Impact

- `assets/entrypoint.core` — the `claude()` wrapper function body gains a guard.
- `internal/agent/claude.go` — `Command()` argument ordering (`--add-dir` moved after `extraArgs`).
- `CHANGELOG.md` — a Fixed entry.
- No config, flag, or public API changes.
