## 1. Fix the entrypoint wrapper

- [x] 1.1 In `assets/entrypoint.core`, guard the `claude()` wrapper so it adds `--add-dir /opt/asylum-skills` only when there are no arguments or the first argument begins with `-`; otherwise call `command claude "$@"` unchanged.
- [x] 1.2 Keep the existing skills-present check (`/opt/asylum-skills/.claude/skills` non-empty) as a precondition for adding the flag.

## 2. Harden the primary session path

- [x] 2.1 In `internal/agent/claude.go` `Command()`, move `--add-dir <KitSkillsDir>` to after the `extraArgs` tokens so the variadic flag cannot swallow a positional passthrough prompt.
- [x] 2.2 Update or add a `claude_test.go` case asserting `--add-dir` follows passthrough args in the generated command.

## 3. Documentation

- [x] 3.1 Add a Fixed entry to `CHANGELOG.md` under Unreleased noting that `claude` subcommands (`mcp`, `doctor`, …) and positional prompts work in the sandbox again.
