## ADDED Requirements

### Requirement: All three agent CLIs installed
The Dockerfile SHALL install agent CLIs based on the active agents config. When no agents config is specified (default), only Claude Code SHALL be installed.

#### Scenario: Default agents (no config)
- **WHEN** the container starts with no agents config specified
- **THEN** only `claude` is available in PATH

#### Scenario: Only claude selected
- **WHEN** the container starts with `agents: [claude]`
- **THEN** only `claude` is available in PATH; `gemini` and `codex` are not installed

#### Scenario: No agents selected
- **WHEN** the container starts with `agents: []`
- **THEN** no agent CLIs are installed

### Requirement: ASYLUM_ environment variables
The entrypoint SHALL use ASYLUM_ prefixed environment variables instead of AGENTBOX_.

#### Scenario: Docker flag check
- **WHEN** `ASYLUM_DOCKER=1` is set
- **THEN** the entrypoint starts the Docker daemon

#### Scenario: Java version
- **WHEN** `ASYLUM_JAVA_VERSION` is set
- **THEN** the entrypoint selects that Java version via mise

### Requirement: Welcome banner
The entrypoint SHALL display an Asylum-branded welcome banner with tool versions for active profiles and installed agents only. If a project entrypoint has set the `PROJECT_BANNER` variable, those banner lines SHALL also be displayed.

#### Scenario: Interactive terminal with all profiles and agents
- **WHEN** the container starts with a TTY and all profiles and agents are active
- **THEN** the banner shows "Asylum Development Environment" with Python, Node.js, Java, Claude, Codex, Gemini, and Opencode versions

#### Scenario: Interactive terminal with subset of profiles and agents
- **WHEN** the container starts with a TTY, only the java profile is active, and only claude is installed
- **THEN** the banner shows Java and Claude versions but not Python, Node.js, Gemini, or Codex

#### Scenario: Project kit banner lines
- **WHEN** the container starts with a TTY and project kits have banner lines
- **THEN** the banner includes the project kit banner lines from the `PROJECT_BANNER` variable

### Requirement: Base entrypoint sources project entrypoint
The base entrypoint SHALL source `/usr/local/bin/project-entrypoint.sh` if it exists, before the welcome banner block. Failures in the project entrypoint SHALL NOT abort the base entrypoint.

#### Scenario: Project entrypoint exists
- **WHEN** the container starts and `/usr/local/bin/project-entrypoint.sh` exists
- **THEN** the base entrypoint SHALL source it before printing the welcome banner

#### Scenario: Project entrypoint does not exist
- **WHEN** the container starts and `/usr/local/bin/project-entrypoint.sh` does not exist
- **THEN** the base entrypoint SHALL continue without error

#### Scenario: Project entrypoint fails
- **WHEN** the project entrypoint script exits with a non-zero status
- **THEN** the base entrypoint SHALL continue without aborting

### Requirement: Core CLI tools under canonical names
The base image SHALL provide common CLI tools under their canonical command names, so agents and users can invoke them as documented upstream rather than under Debian-renamed binaries.

#### Scenario: ripgrep available as rg
- **WHEN** the container starts
- **THEN** `rg` is on PATH and `rg --version` succeeds

#### Scenario: fd available as fd
- **WHEN** the container starts
- **THEN** `fd` is on PATH (resolving to the `fdfind` binary) and `fd --version` succeeds

#### Scenario: file available
- **WHEN** the container starts
- **THEN** `file` is on PATH and `file --version` succeeds
