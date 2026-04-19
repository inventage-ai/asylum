# kit-skills-delivery Specification

## Purpose
Defines the shared mechanism by which kits expose Claude Code skills to the container: a container-owned staging root at `/opt/asylum-skills` populated at image-build time, plus the agent-launch integration that makes Claude discover those skills via `--add-dir`.

## Requirements

### Requirement: ProvidesSkills kit field
The `kit.Kit` struct SHALL expose a `ProvidesSkills bool` field. Kits that stage one or more Claude Code skills for runtime use SHALL set this field to `true`.

#### Scenario: Skill-providing kit declares ProvidesSkills
- **WHEN** a kit such as `agent-browser` or `ast-grep` is registered
- **THEN** its `ProvidesSkills` field is `true`

#### Scenario: Non-skill kit leaves ProvidesSkills false
- **WHEN** a kit that does not stage a Claude skill is registered (e.g. `docker`, `python`, `cx`, `rtk`)
- **THEN** its `ProvidesSkills` field remains at the zero value `false`

### Requirement: Shared skill staging root
The system SHALL define `/opt/asylum-skills` as the single container-local root under which all skill-providing kits stage their skills. Each kit SHALL stage its skill into `/opt/asylum-skills/.claude/skills/<skill-name>/` at image-build time.

#### Scenario: Kit stages skill under shared root
- **WHEN** a kit with `ProvidesSkills: true` builds its Docker layer
- **THEN** the resulting image contains the skill at `/opt/asylum-skills/.claude/skills/<skill-name>/SKILL.md` (and related files)

#### Scenario: Multiple skill kits share one root
- **WHEN** two or more kits with `ProvidesSkills: true` are active in the same image
- **THEN** each kit's skill is present under `/opt/asylum-skills/.claude/skills/` as a sibling directory

### Requirement: Skill-providing kits do not bind-mount into $HOME/.claude
The system SHALL NOT generate entrypoint logic that creates directories under `$HOME/.claude/skills/` or that uses `mount --bind` to deliver skills. Skill delivery SHALL be exclusively via the shared staging root and `--add-dir`.

#### Scenario: No mkdir against ~/.claude/skills in entrypoint
- **WHEN** the assembled entrypoint script is inspected for a configuration that includes only skill-providing kits (no `cx` or `rtk`)
- **THEN** the script contains no `mkdir -p "$HOME/.claude/skills/..."` and no `mount --bind` targeting `$HOME/.claude/skills/`

#### Scenario: Host ~/.claude/skills untouched in shared mode
- **WHEN** a container runs in shared agent-config mode with `agent-browser` and `ast-grep` active and the host's `~/.claude/skills/` does not contain `agent-browser/` or `ast-grep/` before the run
- **THEN** after the container exits, `~/.claude/skills/agent-browser/` and `~/.claude/skills/ast-grep/` still do not exist on the host

### Requirement: Agent launch passes --add-dir when skill kits are active
The container layer SHALL communicate to the agent command builder whether any active kit declares `ProvidesSkills: true`. When at least one such kit is active and the agent is Claude, the generated launch command SHALL include `--add-dir /opt/asylum-skills`.

#### Scenario: Skill kit active, Claude agent
- **WHEN** the `agent-browser` kit is active and the configured agent is `claude`
- **THEN** the generated agent launch command includes `--add-dir /opt/asylum-skills`

#### Scenario: No skill kit active, Claude agent
- **WHEN** no active kit has `ProvidesSkills: true` and the configured agent is `claude`
- **THEN** the generated agent launch command does not include `--add-dir /opt/asylum-skills`

#### Scenario: Skill kit active, non-Claude agent
- **WHEN** a kit with `ProvidesSkills: true` is active and the configured agent is `gemini`, `codex`, `opencode`, or `echo`
- **THEN** the generated agent launch command is unchanged (no `--add-dir` added)

### Requirement: Interactive claude invocations pick up kit skills
The container entrypoint SHALL install a shell wrapper (function or alias) named `claude` in both the zsh and bash startup files so that interactive invocations of `claude` from a secondary shell inside the container also receive `--add-dir /opt/asylum-skills` when the shared skills directory contains at least one skill.

#### Scenario: Wrapper adds flag when skills are present
- **WHEN** a user runs `claude` interactively in a secondary shell and `/opt/asylum-skills/.claude/skills` contains at least one entry
- **THEN** the underlying invocation includes `--add-dir /opt/asylum-skills`

#### Scenario: Wrapper does nothing when skills are absent
- **WHEN** a user runs `claude` interactively and `/opt/asylum-skills/.claude/skills` is empty or does not exist
- **THEN** the underlying invocation does not include `--add-dir /opt/asylum-skills`
