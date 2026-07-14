## MODIFIED Requirements

### Requirement: Agent launch passes --add-dir when skill kits are active
The container layer SHALL communicate to the agent command builder whether any active kit declares `ProvidesSkills: true`. When at least one such kit is active and the agent is Claude, the generated launch command SHALL include `--add-dir /opt/asylum-skills`, positioned **after** any passthrough (`extraArgs`) tokens so the variadic `--add-dir` flag cannot swallow a positional prompt.

#### Scenario: Skill kit active, Claude agent
- **WHEN** the `agent-browser` kit is active and the configured agent is `claude`
- **THEN** the generated agent launch command includes `--add-dir /opt/asylum-skills`

#### Scenario: --add-dir placed after passthrough args
- **WHEN** the `agent-browser` kit is active, the agent is `claude`, and passthrough `extraArgs` are present
- **THEN** `--add-dir /opt/asylum-skills` appears after those passthrough tokens in the generated command

#### Scenario: No skill kit active, Claude agent
- **WHEN** no active kit has `ProvidesSkills: true` and the configured agent is `claude`
- **THEN** the generated agent launch command does not include `--add-dir /opt/asylum-skills`

#### Scenario: Skill kit active, non-Claude agent
- **WHEN** a kit with `ProvidesSkills: true` is active and the configured agent is `gemini`, `codex`, `opencode`, or `echo`
- **THEN** the generated agent launch command is unchanged (no `--add-dir` added)

### Requirement: Interactive claude invocations pick up kit skills
The container entrypoint SHALL install a shell wrapper (function or alias) named `claude` in both the zsh and bash startup files so that interactive invocations of `claude` from a secondary shell inside the container also receive `--add-dir /opt/asylum-skills` when the shared skills directory contains at least one skill. The wrapper SHALL add `--add-dir` only on session invocations — when no arguments are passed, or when the first argument is a flag (begins with `-`). When the first argument is a positional token (a subcommand such as `mcp`, `doctor`, or `update`, or a positional prompt), the wrapper SHALL pass the arguments through to `command claude` unchanged, because `--add-dir` is variadic and would otherwise swallow the positional token.

#### Scenario: Wrapper adds flag for a bare session invocation
- **WHEN** a user runs `claude` with no arguments in a secondary shell and `/opt/asylum-skills/.claude/skills` contains at least one entry
- **THEN** the underlying invocation includes `--add-dir /opt/asylum-skills`

#### Scenario: Wrapper adds flag for a flag-led invocation
- **WHEN** a user runs `claude -c` (or any invocation whose first argument begins with `-`) and `/opt/asylum-skills/.claude/skills` contains at least one entry
- **THEN** the underlying invocation includes `--add-dir /opt/asylum-skills`

#### Scenario: Wrapper leaves subcommands untouched
- **WHEN** a user runs `claude mcp list` (or any invocation whose first argument is a subcommand)
- **THEN** the underlying invocation is `command claude mcp list` with no `--add-dir` inserted, so the subcommand runs correctly

#### Scenario: Wrapper leaves positional prompts untouched
- **WHEN** a user runs `claude "fix the bug"` with a positional prompt as the first argument
- **THEN** the underlying invocation does not insert `--add-dir` and the prompt is not swallowed

#### Scenario: Wrapper does nothing when skills are absent
- **WHEN** a user runs `claude` interactively and `/opt/asylum-skills/.claude/skills` is empty or does not exist
- **THEN** the underlying invocation does not include `--add-dir /opt/asylum-skills`
