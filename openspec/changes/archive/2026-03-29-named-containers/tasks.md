## 1. Container Naming

- [x] 1.1 Extract sanitization logic from `safeHostname` into a `sanitizeProject(projectDir string) string` helper in `internal/container/container.go`
- [x] 1.2 Update `ContainerName` to return `asylum-<hash[:12]>-<sanitized-project>`
- [x] 1.3 Add `OldContainerName(projectDir string) string` that returns the old `asylum-<hash[:12]>` format (used by migration)
- [x] 1.4 Update `safeHostname` to use `sanitizeProject`

## 2. Migration

- [x] 2.1 Add `ports.RenameContainer(oldName, newName string) error` in `internal/ports/ports.go` — find entry by old container name and update to new name
- [x] 2.2 Add `MigrateProjectDir(projectDir string) error` in `internal/container/container.go` — rename `~/.asylum/projects/<old>/` to `<new>/` and call `ports.RenameContainer`
- [x] 2.3 Call `MigrateProjectDir(projectDir)` in `cmd/asylum/main.go` early in the main flow (before session tracking, port allocation, etc.)

## 3. Tests

- [x] 3.1 Update `TestContainerName` — new expected format includes project suffix
- [x] 3.2 Add test for `OldContainerName` — returns old hash-only format
- [x] 3.3 Add test for `sanitizeProject` — various edge cases (special chars, long names, empty basename)
- [x] 3.4 Add test for `ports.RenameContainer` — renames matching entry, no-op if not found
- [x] 3.5 Update any other tests that construct or assert container name format
- [x] 3.6 Verify all tests pass

## 4. Spec and Changelog

- [x] 4.1 Update `openspec/specs/container-assembly/spec.md` with new name format
- [x] 4.2 Add CHANGELOG entry under Unreleased
