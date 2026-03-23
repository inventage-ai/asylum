## ADDED Requirements

### Requirement: PreRunCmds in exec args
When `ExecOpts.PreRunCmds` is non-empty, the agent command SHALL be wrapped in `bash -c` with PATH setup, the pre-run commands joined by `&&`, and the agent command executed via `exec`.

#### Scenario: Pre-run commands present
- **WHEN** `PreRunCmds` contains install commands
- **THEN** the exec args become `bash -c "PATH_SETUP; cmd1 && cmd2 ; exec agent_binary args"`

#### Scenario: No pre-run commands
- **WHEN** `PreRunCmds` is empty
- **THEN** the exec args are the agent command directly (no bash wrapper)

### Requirement: Shadow volume ownership
After starting a new container, the system SHALL fix ownership of shadow `node_modules` volumes by running `chown` as root so the container user can write to them.

#### Scenario: New container with shadow volumes
- **WHEN** a container is started for a project with Node.js package.json files
- **THEN** each shadow volume directory is chowned to `claude:claude`

#### Scenario: Existing running container
- **WHEN** the container is already running
- **THEN** no chown is performed (it was done at first start)
