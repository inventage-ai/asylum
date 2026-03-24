## ADDED Requirements

### Requirement: All three agent CLIs installed
The Dockerfile SHALL install Claude Code, Gemini CLI, and Codex in the same image.

#### Scenario: All agents available
- **WHEN** the container starts
- **THEN** `claude`, `gemini`, and `codex` are all available in PATH

### Requirement: ASYLUM_ environment variables
The entrypoint SHALL use ASYLUM_ prefixed environment variables instead of AGENTBOX_.

#### Scenario: Docker flag check
- **WHEN** `ASYLUM_DOCKER=1` is set
- **THEN** the entrypoint starts the Docker daemon

#### Scenario: Java version
- **WHEN** `ASYLUM_JAVA_VERSION` is set
- **THEN** the entrypoint selects that Java version via mise

### Requirement: Welcome banner
The entrypoint SHALL display an Asylum-branded welcome banner with tool versions for active profiles only.

#### Scenario: Interactive terminal with all profiles
- **WHEN** the container starts with a TTY and all profiles are active
- **THEN** the banner shows "Asylum Development Environment" with Python, Node.js, Java, and agent CLI versions

#### Scenario: Interactive terminal with subset of profiles
- **WHEN** the container starts with a TTY and only the java profile is active
- **THEN** the banner shows "Asylum Development Environment" with Java and agent CLI versions but not Python or Node.js
