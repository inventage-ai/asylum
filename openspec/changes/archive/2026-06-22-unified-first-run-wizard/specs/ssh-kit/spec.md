## MODIFIED Requirements

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

### Requirement: Project mode
In `project` mode, the kit SHALL generate a per-project ed25519 key pair at `~/.asylum/projects/<container>/ssh/` if one does not exist, and mount the key pair into `~/.ssh/`. The same silent-on-success / single-line-notice behavior as `isolated` mode SHALL apply, with the success line referencing the project-specific path.

#### Scenario: First run in project mode
- **WHEN** the credential function runs in `project` mode for a container and no key exists
- **THEN** a key pair SHALL be generated in the project-specific SSH directory with stdout/stderr captured
- **AND** on success, one line SHALL be printed: `Generated SSH key at ~/.asylum/projects/<container>/ssh/id_ed25519.pub — see asylum-reference.md for usage.`

#### Scenario: Different projects get different keys
- **WHEN** two projects use `project` mode
- **THEN** each SHALL have its own key pair in its respective project directory
