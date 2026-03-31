## Why

The config TUI's kit toggle is broken for opt-in kits. When activating kits like `ast-grep`, `agent-browser`, or `apt`, the code uses `ConfigSnippet` which is commented-out for opt-in kits — so it removes the comment block and re-inserts a commented entry, leaving the kit inactive. Beyond this bug, the current approach of manipulating YAML comments to represent kit state is fragile. The `disabled` field already exists on `KitConfig` and is respected by `KitActive`/`DisabledKits`/`kit.Resolve`, but the config TUI doesn't use it.

## What Changes

- First activation of a kit still removes the comment block and inserts a clean YAML entry (e.g., `  ast-grep:`)
- Subsequent toggles flip the `disabled: true` field on the existing entry instead of removing/re-adding comment blocks
- Re-enabling removes the `disabled` field entirely (clean entry = active), disabling adds `disabled: true`
- Comment blocks are preserved for kits that have never been enabled (hand-editing the config still works)
- The config TUI correctly distinguishes "never enabled" (comment) from "enabled" (entry, no disabled) from "disabled" (entry with `disabled: true`)
- Project-level configs can override with `disabled: false` to enable a globally-disabled kit

## Capabilities

### New Capabilities

- `config-disabled-toggle`: Text-based manipulation of the `disabled` field on kit entries in the config file (set/remove `disabled: true` on a named kit's YAML block)

### Modified Capabilities

- `config-command`: Kit activation/deactivation uses `disabled` field for previously-enabled kits instead of always using comment manipulation
- `profile-config-integration`: Clarify that `disabled: false` in a project config overrides `disabled: true` from global config

## Impact

- `cmd/asylum/config.go`: Activation/deactivation logic reworked to use disabled field for known entries
- `internal/config/sync.go`: New function to set/remove `disabled` field on a kit entry
- `internal/config/config.go`: `KitActive` already handles `disabled` — no change needed
- Existing config files continue to work without migration (absent `disabled` = active)
