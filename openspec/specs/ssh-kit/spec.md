# ssh-kit Specification

## Purpose

Gives the container an SSH identity for Git operations, at a level of exposure the user picks. `isolated` (the default) generates a dedicated ed25519 key so the host's own keys never enter the sandbox, `project` does the same per project, and `shared` mounts the host's `~/.ssh` directly. The host's `known_hosts` is mounted in the generated-key modes so accepted host keys persist.

## Requirements

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
In `isolated` mode, the kit SHALL generate an ed25519 key pair at `~/.asylum/ssh/` if one does not exist, and mount the key pair into `~/.ssh/`. Key generation SHALL be silent on success: `ssh-keygen` stdout and stderr SHALL be captured and discarded when the process exits zero. On a non-zero exit, the captured output SHALL be included in the returned error. After a successful generation, the kit SHALL emit a single user-facing line via the project `log` package indicating the public key path and pointing at `asylum-reference.md` for usage details. The kit SHALL NOT print the public key contents, randomart, ssh-keygen banner, or instructions about adding the key to a Git host — those details live in `asylum-reference.md`.

#### Scenario: First run in isolated mode
- **WHEN** the credential function runs in `isolated` mode and no key exists at `~/.asylum/ssh/id_ed25519`
- **THEN** the directory SHALL be created with mode 0700
- **AND** a new ed25519 key pair SHALL be generated with `ssh-keygen`'s stdout and stderr captured
- **AND** on success, exactly one line SHALL be printed to the user: `Generated SSH key at ~/.asylum/ssh/id_ed25519.pub — see asylum-reference.md for usage.`
- **AND** the public key contents SHALL NOT be printed

#### Scenario: ssh-keygen fails
- **WHEN** `ssh-keygen` exits non-zero during isolated-mode key generation
- **THEN** the captured stdout/stderr SHALL be included in the returned error
- **AND** no success-line SHALL be emitted

#### Scenario: Key already exists in isolated mode
- **WHEN** the credential function runs in `isolated` mode and `~/.asylum/ssh/id_ed25519` exists
- **THEN** no key generation SHALL occur, no success-line SHALL be emitted, and the existing key pair SHALL be mounted

### Requirement: Shared mode
In `shared` mode, the kit SHALL mount the host's entire `~/.ssh/` directory into the container in read-write mode without generating any keys.

#### Scenario: Shared mode mounting
- **WHEN** the credential function runs in `shared` mode
- **THEN** `~/.ssh/` SHALL be mounted as a single read-write directory bind mount

### Requirement: Project mode
In `project` mode, the kit SHALL generate a per-project ed25519 key pair at `~/.asylum/projects/<container>/ssh/` if one does not exist, and mount the key pair into `~/.ssh/`. The same silent-on-success / single-line-notice behavior as `isolated` mode SHALL apply, with the success line referencing the project-specific path.

#### Scenario: First run in project mode
- **WHEN** the credential function runs in `project` mode for a container and no key exists
- **THEN** a key pair SHALL be generated in the project-specific SSH directory with stdout/stderr captured
- **AND** on success, one line SHALL be printed: `Generated SSH key at ~/.asylum/projects/<container>/ssh/id_ed25519.pub — see asylum-reference.md for usage.`

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
