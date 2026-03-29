## MODIFIED Requirements

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
