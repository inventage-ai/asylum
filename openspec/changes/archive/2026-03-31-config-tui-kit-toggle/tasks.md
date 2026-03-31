## 1. Disabled field manipulation

- [x] 1.1 Add `SetKitDisabled(path, kitName)` function in `internal/config/sync.go` — inserts `disabled: true` as first property under the kit entry using `findKey`/`insertAfter` helpers
- [x] 1.2 Add `RemoveKitDisabled(path, kitName)` function in `internal/config/sync.go` — finds and removes the `disabled: true` line from a kit entry
- [x] 1.3 Add `KitExistsInFile(path, kitName) bool` function in `internal/config/sync.go` — checks if a kit key exists as an active YAML entry (not comment) in the file, used to distinguish "never enabled" from "disabled"
- [x] 1.4 Write tests for `SetKitDisabled`, `RemoveKitDisabled`, and `KitExistsInFile`

## 2. Config TUI logic

- [x] 2.1 Fix `activeKits` population in `cmd/asylum/config.go` to use `cfg.KitActive(name)` instead of map membership, so disabled kits show as unchecked
- [x] 2.2 Rework kit activation in `config.go`: if kit exists in global config file (via `KitExistsInFile`) call `RemoveKitDisabled`, otherwise remove comment block and insert `  <name>:` entry
- [x] 2.3 Rework kit deactivation in `config.go`: call `SetKitDisabled` instead of `RemoveKitEntry` + `SyncKitCommentToConfig`
- [x] 2.4 Remove the `kit.ConfigSnippet` dependency from the config command (activation no longer uses it)

## 3. Merge semantics

- [x] 3.1 Verify `disabled: false` in project config correctly overrides `disabled: true` from global config (existing `mergeKitConfig` should handle this — add test if missing)

## 4. Cleanup

- [x] 4.1 Remove dead code: `RemoveKitEntry` (no longer called; `SyncKitCommentToConfig` retained — still used by `kitsync.go`)
- [x] 4.2 Update `config-command` spec in `openspec/specs/` to reflect new behavior
