## Why

The `apt` kit is a configuration container for specifying extra system packages — it has no tools, no Docker snippet of its own, and toggling it on/off in the config screen has no meaningful effect. Showing it alongside real kits (docker, python, ast-grep) confuses users about what it does.

## What Changes

- Add a `Hidden` field to the Kit struct that excludes a kit from all interactive selection surfaces (config TUI, new-kit sync prompt, disabled-kits list in sandbox rules)
- Set `Hidden: true` on the `apt` kit
- Hidden kits remain fully functional — they activate via config like any other kit, they just don't appear in selection UIs

## Capabilities

### New Capabilities

- `hidden-kit-flag`: A Kit-level flag that excludes kits from interactive selection surfaces while keeping them fully functional

### Modified Capabilities

## Impact

- `internal/kit/kit.go` — new `Hidden` field on Kit struct
- `internal/kit/apt.go` — set `Hidden: true`
- `cmd/asylum/config.go` — filter hidden kits from Kits tab
- `cmd/asylum/main.go` — filter hidden kits from sync prompt
- `internal/container/container.go` — filter hidden kits from disabled-kits list in sandbox rules
