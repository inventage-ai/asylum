## Why

The GitHub kit installs `gh` but provides no way to authenticate it. Users must run `gh auth login` inside each new container. Other kits (Maven) already support credential mounting — GitHub should too, using the existing credential system.

## What Changes

- Add a `CredentialFunc` to the GitHub kit that mounts the host's `~/.config/gh/` directory read-only into the container when `credentials: auto` is set
- Extend `CredentialMount` with a `HostPath` field for pass-through mounting (existing `Content` field generates files; `HostPath` mounts a host file directly)
- The GitHub kit uses `auto` mode only (no `explicit` mode needed — there's only one config file)

## Capabilities

### New Capabilities
- `github-kit-credentials`: GitHub kit credential provider that mounts gh auth config from host

### Modified Capabilities
- `kit-credentials`: Add `HostPath` field to `CredentialMount` for direct host file mounting (no content generation)

## Impact

- `internal/kit/kit.go`: `CredentialMount` struct gains `HostPath` field
- `internal/kit/github.go`: Add `CredentialFunc` and `CredentialLabel`
- `internal/container/container.go`: Handle `HostPath` mounts (bind directly instead of writing content)
