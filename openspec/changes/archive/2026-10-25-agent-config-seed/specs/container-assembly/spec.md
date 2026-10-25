# Spec Delta

## MODIFIED Requirements

### Requirement: Agent config seeding
On first run for an agent in `isolated` or `project` mode, when the target config directory does not exist, the system SHALL create it and seed it from the agent's native host config according to the agent's seed mode: `full` (default) copies the entire native config directory, `settings` copies only the entries of the agent's settings allowlist that exist, and `none` copies nothing.

#### Scenario: First run with existing native config
- **WHEN** `~/.asylum/agents/claude/` does not exist but `~/.claude/` does, and no seed mode is configured
- **THEN** contents of `~/.claude/` are copied to `~/.asylum/agents/claude/`

#### Scenario: First run without native config
- **WHEN** neither `~/.asylum/agents/claude/` nor `~/.claude/` exists
- **THEN** `~/.asylum/agents/claude/` is created empty

#### Scenario: Seed mode settings
- **WHEN** the seed mode is `settings` and `~/.claude/` contains `settings.json`, `CLAUDE.md`, `agents/` and `projects/`
- **THEN** the target directory contains `settings.json`, `CLAUDE.md` and `agents/`, but not `projects/`

#### Scenario: Seed mode none
- **WHEN** the seed mode is `none` and `~/.claude/` exists
- **THEN** the target directory is created empty

#### Scenario: Project mode target
- **WHEN** the agent runs in `project` mode for the first time in a project
- **THEN** the per-project config directory is seeded according to the same seed mode

#### Scenario: Existing target is left alone
- **WHEN** the target config directory already exists
- **THEN** nothing is copied, regardless of the seed mode

#### Scenario: Unknown seed mode
- **WHEN** the seed mode is a value other than `none`, `settings` or `full` and the target config directory does not exist yet
- **THEN** the run aborts with an error naming the valid values
