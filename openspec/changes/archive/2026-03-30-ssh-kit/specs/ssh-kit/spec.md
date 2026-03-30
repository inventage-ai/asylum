## ADDED Requirements

### Requirement: SSH kit registration
The system SHALL register an `ssh` kit with `TierAlwaysOn` tier, providing SSH key management and credential mounting.

#### Scenario: Kit is always active
- **WHEN** kits are resolved for any configuration
- **THEN** the `ssh` kit SHALL be included regardless of config entries

### Requirement: SSH isolation configuration
The SSH kit SHALL support an `isolation` config key with values `isolated` (default), `shared`, and `project`.

#### Scenario: No isolation configured
- **WHEN** the SSH kit's isolation is not set in config
- **THEN** the system SHALL default to `isolated` mode

#### Scenario: Explicit isolation value
- **WHEN** the SSH kit's isolation is set to `shared`, `isolated`, or `project`
- **THEN** the system SHALL use that mode for key storage and mounting

### Requirement: Isolated mode
In `isolated` mode, the kit SHALL generate an ed25519 key pair at `~/.asylum/ssh/` if one does not exist, and mount the key pair into `~/.ssh/`.

#### Scenario: First run in isolated mode
- **WHEN** the credential function runs in `isolated` mode and no key exists at `~/.asylum/ssh/id_ed25519`
- **THEN** the directory SHALL be created with mode 0700, a new ed25519 key pair SHALL be generated, and the public key SHALL be printed to stdout

#### Scenario: Key already exists in isolated mode
- **WHEN** the credential function runs in `isolated` mode and `~/.asylum/ssh/id_ed25519` exists
- **THEN** no key generation SHALL occur and the existing key pair SHALL be mounted

### Requirement: Shared mode
In `shared` mode, the kit SHALL mount the host's entire `~/.ssh/` directory into the container in read-write mode without generating any keys.

#### Scenario: Shared mode mounting
- **WHEN** the credential function runs in `shared` mode
- **THEN** `~/.ssh/` SHALL be mounted as a single read-write directory bind mount

### Requirement: Project mode
In `project` mode, the kit SHALL generate a per-project ed25519 key pair at `~/.asylum/projects/<container>/ssh/` if one does not exist, and mount the key pair into `~/.ssh/`.

#### Scenario: First run in project mode
- **WHEN** the credential function runs in `project` mode for a container and no key exists
- **THEN** a key pair SHALL be generated in the project-specific SSH directory and mounted

#### Scenario: Different projects get different keys
- **WHEN** two projects use `project` mode
- **THEN** each SHALL have its own key pair in its respective project directory

### Requirement: Known hosts mounting
In `isolated` and `project` modes, the host's `~/.ssh/known_hosts` SHALL be mounted at `~/.ssh/known_hosts` in read-write mode if the file exists.

#### Scenario: Host known_hosts exists
- **WHEN** the credential function runs in `isolated` or `project` mode and `~/.ssh/known_hosts` exists on the host
- **THEN** it SHALL be mounted to `~/.ssh/known_hosts` in read-write mode

#### Scenario: Host known_hosts does not exist
- **WHEN** `~/.ssh/known_hosts` does not exist on the host
- **THEN** no known_hosts mount SHALL be returned

### Requirement: Credential mode bypass for always-on kits
The container assembly credential loop SHALL treat unconfigured credential mode as `auto` for kits with `TierAlwaysOn` tier, instead of skipping them.

#### Scenario: Always-on kit with no credential config
- **WHEN** a `TierAlwaysOn` kit has a credential function and no credential mode is configured
- **THEN** the credential function SHALL be called with `CredentialAuto` mode

### Requirement: Container name in credential opts
The `CredentialOpts` struct SHALL include a `ContainerName` field so credential functions can resolve per-project paths.

#### Scenario: Container name available to credential function
- **WHEN** a kit's credential function is called
- **THEN** the `ContainerName` field SHALL be populated with the current container name
