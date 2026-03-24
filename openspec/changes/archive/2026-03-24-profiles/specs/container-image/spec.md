## MODIFIED Requirements

### Requirement: Welcome banner
The entrypoint SHALL display an Asylum-branded welcome banner with tool versions for active profiles only.

#### Scenario: Interactive terminal with all profiles
- **WHEN** the container starts with a TTY and all profiles are active
- **THEN** the banner shows "Asylum Development Environment" with Python, Node.js, Java, and agent CLI versions

#### Scenario: Interactive terminal with subset of profiles
- **WHEN** the container starts with a TTY and only the java profile is active
- **THEN** the banner shows "Asylum Development Environment" with Java and agent CLI versions but not Python or Node.js
