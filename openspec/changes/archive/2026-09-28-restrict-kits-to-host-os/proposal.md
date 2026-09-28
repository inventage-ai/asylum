# Proposal

## Why

The `iterm` and `dropshare` kits only work on a macOS host, but Asylum offers them everywhere. On Linux, the kit-sync prompt, the first-run wizard, `asylum config`, the generated config's commented snippets, and the sandbox rules' "Disabled Kits" list all present kits that cannot work there. The agent is also invited to suggest enabling them.

## What Changes

- Kits can declare the one host OS they work on. `iterm` and `dropshare` declare macOS.
- On any other host OS, such kits are left out of every place a kit is offered:
  - the first-run wizard's kit step, and its generated config (no snippet, not even commented)
  - the non-interactive default config
  - the kit-sync prompt and its comment line for declined kits
  - the Kits tab of `asylum config`
  - the "Disabled Kits" list in the sandbox rules
- Such kits are never recorded as known on that host, so a Mac that shares the same `~/.asylum` still gets them offered.
- A kit enabled in config on another OS keeps working exactly as today. Asylum does not block or warn.
- The kits index docs gain the missing `iterm` row.

## Capabilities

### New Capabilities

- `kit-host-os`: Kits that declare a host OS, and how Asylum stops offering them on any other host OS.

### Modified Capabilities

- `kit-state-tracking`: New-kit detection and the post-sync state update skip kits for another host OS.
- `onboarding-wizard`: The kit selection step leaves out kits for another host OS.
- `config-command`: The Kits tab leaves out kits for another host OS.

## Impact

- `internal/kit/kit.go`: a `HostOS` field and a helper that matches it against the host.
- `internal/kit/iterm.go`, `internal/kit/dropshare.go`: set `HostOS: "darwin"`.
- `internal/kit/kit.go` `AssembleConfigSnippets`, `internal/firstrun/config_writer.go`, `internal/config/kitsync.go`, `cmd/asylum/config.go`, `internal/container/container.go`: skip kits for another host OS.
- Existing comment lines for these kits in Linux configs stay as they are.
- `docs/kits/index.md`: add `iterm`. The unreleased CHANGELOG entries for both kits mention the restriction.
