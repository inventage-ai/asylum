## ADDED Requirements

### Requirement: Container home directory matches host
The container user SHALL be created with a home directory path matching the host user's home directory, passed as a build argument.

#### Scenario: macOS host
- **WHEN** the host home directory is `/Users/simon`
- **THEN** the container user's home directory is `/Users/simon`

#### Scenario: Linux host
- **WHEN** the host home directory is `/home/simon`
- **THEN** the container user's home directory is `/home/simon`

### Requirement: Absolute symlinks resolve correctly
Absolute-path symlinks created on the host SHALL resolve correctly inside the container when the containing directory is mounted.

#### Scenario: Claude config symlink
- **WHEN** `~/.claude/` contains a symlink pointing to `$HOME/.claude/targets/foo`
- **THEN** that symlink resolves inside the container because `$HOME` is the same path

### Requirement: No hardcoded home directory paths
All references to the container user's home directory SHALL use `$HOME` (in shell scripts) or runtime-resolved paths (in Go code), not hardcoded `/home/claude`.

#### Scenario: Entrypoint script
- **WHEN** the entrypoint references the user's SSH directory
- **THEN** it uses `$HOME/.ssh` not `/home/claude/.ssh`

#### Scenario: Kit cache directories
- **WHEN** a kit specifies cache directories
- **THEN** they use `~/` prefix resolved at runtime, not absolute paths

### Requirement: Kit snippets use build arg for username
Kit DockerSnippets that switch between root and the container user SHALL use `${USERNAME}` instead of hardcoded `claude`.

#### Scenario: Maven kit switches to root
- **WHEN** the java/maven kit installs packages as root
- **THEN** it switches back with `USER ${USERNAME}` not `USER claude`
