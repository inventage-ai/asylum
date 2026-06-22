## MODIFIED Requirements

### Requirement: Reference document
The system SHALL embed a detailed Asylum reference document (`assets/asylum-reference.md`) via `go:embed`. On container start, the reference doc SHALL be written alongside the rules file and mounted read-only at `<project-dir>/.claude/asylum-reference.md`. The rules file SHALL reference it for troubleshooting and config details. The reference doc SHALL include a link to the changelog on GitHub. The SSH section of the reference doc SHALL include guidance on how to use the kit-generated SSH key with a Git hosting provider (the prose previously printed by `ssh-keygen` callers).

#### Scenario: Reference doc accessible but not auto-loaded
- **WHEN** a container is running
- **THEN** the reference doc SHALL be readable at `.claude/asylum-reference.md` but SHALL NOT be in `.claude/rules/` (so it is not auto-loaded by Claude Code)

#### Scenario: Reference doc content
- **WHEN** the reference doc is read
- **THEN** it SHALL describe the container lifecycle, layered config system, available kits, volume mounting, self-update mechanism, and troubleshooting steps

#### Scenario: SSH section explains usage
- **WHEN** the reference doc is read and the SSH section is examined
- **THEN** it SHALL explain where the generated key lives (`~/.asylum/ssh/id_ed25519.pub` for isolated mode, `~/.asylum/projects/<container>/ssh/id_ed25519.pub` for project mode)
- **AND** it SHALL describe how to add the key to a Git hosting provider (the host commands a user would run to display and copy the public key)
- **AND** it SHALL note that the user MAY replace the generated keys with their own
