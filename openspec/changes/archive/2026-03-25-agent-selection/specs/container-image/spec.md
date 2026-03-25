## MODIFIED Requirements

### Requirement: All three agent CLIs installed
The Dockerfile SHALL install agent CLIs based on the active agents config. When no agents config is specified (default), all three agents (Claude Code, Gemini CLI, Codex) SHALL be installed.

#### Scenario: Default agents (no config)
- **WHEN** the container starts with no agents config specified
- **THEN** only `claude` is available in PATH

#### Scenario: Only claude selected
- **WHEN** the container starts with `agents: [claude]`
- **THEN** only `claude` is available in PATH; `gemini` and `codex` are not installed

#### Scenario: No agents selected
- **WHEN** the container starts with `agents: []`
- **THEN** no agent CLIs are installed

### Requirement: Welcome banner
The entrypoint SHALL display an Asylum-branded welcome banner with tool versions for active profiles and installed agents only.

#### Scenario: Interactive terminal with all profiles and agents
- **WHEN** the container starts with a TTY and all profiles and agents are active
- **THEN** the banner shows "Asylum Development Environment" with Python, Node.js, Java, Claude, Codex, Gemini, and Opencode versions

#### Scenario: Interactive terminal with subset of profiles and agents
- **WHEN** the container starts with a TTY, only the java profile is active, and only claude is installed
- **THEN** the banner shows Java and Claude versions but not Python, Node.js, Gemini, or Codex
