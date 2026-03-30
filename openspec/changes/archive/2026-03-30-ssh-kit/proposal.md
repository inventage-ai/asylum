## Why

SSH key management currently requires a manual `asylum ssh-init` command, and the mounting logic is hardcoded in `container.go` rather than being part of the kit system. Converting SSH to an always-on kit brings it in line with how all other tooling is managed, makes it discoverable via kit listing, and removes the need for a separate CLI command.

## What Changes

- New `ssh` kit (`internal/kit/ssh.go`) at `TierAlwaysOn` with configurable `isolation`:
  - **`isolated`** (default): Generates an ed25519 key pair into `~/.asylum/ssh/`, mounts the key pair into `~/.ssh/`
  - **`shared`**: Mounts the host's entire `~/.ssh/` directory directly (no key generation)
  - **`project`**: Per-project keys in `~/.asylum/projects/<container>/ssh/`, generated if missing
- In all modes, the host's `~/.ssh/known_hosts` is mounted read-write if it exists (except `shared`, which already includes it)
- **BREAKING**: Remove the `asylum ssh-init` CLI command (`cmd/asylum/main.go`)
- Remove the `internal/ssh/` package (logic moves into the kit's `CredentialFunc`)
- Remove the hardcoded SSH volume mount from `internal/container/container.go`

## Capabilities

### New Capabilities
- `ssh-kit`: Always-on kit that manages SSH key generation/mounting with configurable isolation levels

### Modified Capabilities
- `ssh-init`: Replaced by the ssh-kit — the standalone command is removed
- `container-assembly`: SSH volume mounting moves from hardcoded logic to kit-driven credential mounting

## Impact

- `internal/ssh/` — removed entirely
- `internal/kit/ssh.go` — new kit file
- `internal/container/container.go` — remove hardcoded SSH mount block, update credential mode gating for always-on kits
- `cmd/asylum/main.go` — remove `ssh-init` command dispatch
- `internal/config/config.go` — add `SSHIsolation()` accessor
- `assets/asylum-reference.md` — update documentation to reflect SSH as a kit
