# claude-temp-dir Specification

## Purpose
Keeps Claude Code's temp root (scratchpad, background task output, edit-diff repos) on the host per project, so the files survive container restarts and the paths the agent prints open unchanged on the host.

## Requirements

### Requirement: Persistent per-project Claude temp root
When Claude is installed in the session, as the primary agent or as a companion, the container SHALL mount the host directory `~/.asylum/projects/<container-name>/tmp` at the same absolute path inside the container and SHALL set `CLAUDE_CODE_TMPDIR` to that path. Asylum SHALL create the host directory if it does not exist. The mount and env var SHALL be RunArgs with source `core`.

#### Scenario: Claude as primary agent
- **WHEN** a container is started with agent `claude`
- **THEN** `~/.asylum/projects/<container-name>/tmp` is mounted read-write at the identical container path
- **AND** `CLAUDE_CODE_TMPDIR` is set to that path

#### Scenario: Claude as companion
- **WHEN** a container is started with a non-Claude primary agent and `claude` listed as a companion
- **THEN** the same mount and `CLAUDE_CODE_TMPDIR` are present

#### Scenario: Claude not installed
- **WHEN** neither the primary agent nor any companion is `claude`
- **THEN** no temp-root mount is added and `CLAUDE_CODE_TMPDIR` is not set

#### Scenario: Scratchpad survives a container restart
- **WHEN** Claude writes a file to its scratchpad and the container is removed and started again
- **THEN** the file still exists at the same path, both inside the new container and on the host

#### Scenario: Host directory missing
- **WHEN** `~/.asylum/projects/<container-name>/tmp` does not exist at container start
- **THEN** it is created before the container runs

### Requirement: Container-local runtime dir for Claude sockets
When Claude is installed in the session, the container SHALL set `XDG_RUNTIME_DIR` to `/run/user/<uid>`, where `<uid>` is the container user's uid. The base image SHALL provide that directory with mode 0700, owned by the container user, on the container's own filesystem rather than on a host mount.

#### Scenario: Runtime dir set with Claude
- **WHEN** a container is started with Claude installed
- **THEN** `XDG_RUNTIME_DIR` is `/run/user/<uid>`
- **AND** the directory exists, has mode 0700 and is owned by the container user

#### Scenario: Runtime dir not set without Claude
- **WHEN** Claude is not installed in the session
- **THEN** asylum does not set `XDG_RUNTIME_DIR`

### Requirement: Other tools keep the container-local temp dir
Asylum SHALL NOT set or change `TMPDIR` as part of the Claude temp root, so that tools other than Claude continue to write temp files to the container's local `/tmp`.

#### Scenario: TMPDIR untouched
- **WHEN** a container is started with Claude installed
- **THEN** `TMPDIR` is not among the env vars asylum sets

### Requirement: No automatic pruning
Asylum SHALL NOT delete contents of the Claude temp root on container start or stop. `asylum cleanup` SHALL remove it as part of the project data directory.

#### Scenario: Restart keeps old sessions
- **WHEN** the container is restarted
- **THEN** scratchpad directories from earlier sessions remain

#### Scenario: Cleanup removes temp root
- **WHEN** `asylum cleanup` is run for the project
- **THEN** `~/.asylum/projects/<container-name>/tmp` no longer exists
