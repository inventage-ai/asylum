## Why

The config hash used for stale container detection did not include kit credential configuration. This meant that changing credentials in `.asylum` or `.asylum.local` would not trigger asylum's "config changed — restart with --rebuild" warning, leaving the container silently running with outdated credential mounts.

## What Changes

- `ConfigHash` now incorporates kit credential settings (auto mode or explicit server ID list) into its hash computation
- Credential changes now produce a different hash, triggering the existing stale container warning

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `stale-container-detection`: The config hash now covers credentials in addition to volumes, env vars, and ports.

## Impact

- `internal/config/config.go` — `ConfigHash` function extended to iterate kit credentials
- No config format changes, no new fields
