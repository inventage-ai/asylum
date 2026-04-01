## MODIFIED Requirements

### Requirement: Port forwarding
Ports from user config and kit-allocated ports SHALL both be represented as RunArgs in the unified pipeline. User-configured ports SHALL have priority 2 (config), kit-allocated ports SHALL have priority 1 (kit). Deduplication on container port SHALL prevent conflicts.

#### Scenario: Simple port from config
- **WHEN** user config has port `3000`
- **THEN** a RunArg with Flag=`-p`, Value=`3000:3000`, Source=`user config (ports)`, Priority=2 SHALL be added

#### Scenario: Mapped port from config
- **WHEN** user config has port `8080:80`
- **THEN** a RunArg with Flag=`-p`, Value=`8080:80`, Source=`user config (ports)`, Priority=2 SHALL be added

#### Scenario: Kit-allocated ports
- **WHEN** the ports kit allocates ports 10000-10004
- **THEN** five RunArgs with Flag=`-p`, Source=`ports kit`, Priority=1 SHALL be added

#### Scenario: Config port overrides kit port on same container port
- **WHEN** user config has port `3000:3000` and the ports kit also allocated port 3000
- **THEN** only the user config mapping SHALL appear in the final args

### Requirement: Common volume mounts
The container SHALL include all common mounts: project dir at real path, gitconfig, caches (as named Docker volumes), history, custom volumes, and direnv. SSH mounts are handled by the SSH kit's credential function. All mounts SHALL be represented as RunArgs with source `core` and priority 0, except user-configured volumes which SHALL have source `user config (volumes)` and priority 2.

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

### Requirement: Agent-specific mounts and env vars
The container SHALL mount the agent's asylum config dir and set agent-specific env vars. These SHALL be represented as RunArgs with source `core`.

#### Scenario: Claude agent
- **WHEN** agent is claude
- **THEN** RunArgs for the config dir mount and env vars SHALL have source `core`
