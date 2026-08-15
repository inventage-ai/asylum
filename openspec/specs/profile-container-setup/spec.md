# profile-container-setup Specification

## Purpose

Covers where an agent's config directory is mounted from, which follows that agent's configured isolation level: the host directory for `shared`, an asylum-managed one for `isolated`, or a per-project one for `project`.

## Requirements

### Requirement: Agent config volume mount
The agent config volume mount SHALL respect the isolation level configured for each agent.

#### Scenario: Shared isolation
- **WHEN** the Claude agent has `config: shared`
- **THEN** the host `~/.claude` directory is mounted directly at `~/.claude` inside the container

#### Scenario: Isolated (default)
- **WHEN** the Claude agent has `config: isolated` or no config value
- **THEN** `~/.asylum/agents/claude/` is mounted at `~/.claude` inside the container

#### Scenario: Project isolation
- **WHEN** the Claude agent has `config: project`
- **THEN** a per-project directory (`~/.asylum/projects/<container>/claude-config/`) is mounted at `~/.claude` inside the container
