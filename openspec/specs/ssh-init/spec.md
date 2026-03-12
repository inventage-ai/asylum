## ADDED Requirements

### Requirement: SSH directory initialization
The ssh-init command SHALL create `~/.asylum/ssh/` with mode 0700, copy known_hosts if present, and generate an Ed25519 key pair if none exists.

#### Scenario: First run
- **WHEN** `asylum ssh-init` is run and no key exists
- **THEN** directory is created, known_hosts copied, key generated, public key printed

#### Scenario: Key already exists
- **WHEN** `asylum ssh-init` is run and key exists
- **THEN** user is informed the key exists
