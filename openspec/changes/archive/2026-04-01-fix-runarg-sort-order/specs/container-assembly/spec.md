## MODIFIED Requirements

### Requirement: Common volume mounts
The container SHALL include all common mounts: project dir at real path, gitconfig, caches (as named Docker volumes), history, custom volumes, and direnv. SSH mounts are handled by the SSH kit's credential function. All mounts SHALL be represented as RunArgs with source `core` and priority 0, except user-configured volumes which SHALL have source `user config (volumes)` and priority 2. The docker subcommand (`run`) and mode flag (`-d`) SHALL NOT be emitted as RunArgs; they are prepended during flattening.

#### Scenario: All common mounts present
- **WHEN** gitconfig exists and project has .envrc
- **THEN** all conditional mounts are included as RunArgs with source `core`

#### Scenario: Missing optional paths
- **WHEN** gitconfig does not exist
- **THEN** that mount is omitted, all others remain

#### Scenario: Cache directories use named volumes
- **WHEN** the container is started
- **THEN** cache directories are mounted as named Docker volumes via `--mount` RunArgs with source `core`

#### Scenario: User volume conflicts with core mount
- **WHEN** a user-configured volume mounts to the same container path as a core mount
- **THEN** the user volume (priority 2) SHALL override the core mount (priority 0)

#### Scenario: No run or -d in RunArgs
- **WHEN** `RunArgs()` is called
- **THEN** the returned resolved args SHALL NOT contain entries with Flag `run` or `-d`
