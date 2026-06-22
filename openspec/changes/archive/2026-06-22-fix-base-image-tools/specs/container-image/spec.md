## ADDED Requirements

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
