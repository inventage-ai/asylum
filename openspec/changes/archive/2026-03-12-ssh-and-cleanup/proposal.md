## Why

SSH init and cleanup are standalone features per PLAN.md sections 5.6 and 5.7. They were implemented alongside the CLI entry point since they're small and self-contained.

## What Changes

- `internal/ssh/ssh.go`: SSH directory setup and Ed25519 key generation per PLAN.md section 5.6
- Cleanup in `cmd/asylum/main.go`: Image removal and cached data cleanup per PLAN.md section 5.7
- Both already implemented in the cli-entrypoint change

## Capabilities

### New Capabilities
- `ssh-init`: SSH directory setup with key generation and known_hosts copy
- `cleanup-command`: Image removal and interactive cache cleanup

### Modified Capabilities

None.

## Impact

- Already implemented in `internal/ssh/ssh.go` and `cmd/asylum/main.go`
- This change records the spec for completeness
