## 1. Port allocation changes

- [x] 1.1 Change `BasePort` in `internal/ports/ports.go` from `10000` to `7001`
- [x] 1.2 In `Allocate`, when the matched entry for `projectDir` has `Start >= 10000`, drop it from `reg.Ranges` and fall through to the new-allocation branch (using the current `containerName`)
- [x] 1.3 Update `nextStart` to ignore entries with `Start >= 10000` when computing the next free start
- [x] 1.4 Persist registry writes in both the "stale removed" and "fresh allocation" paths

## 2. Tests

- [x] 2.1 Update existing tests in `internal/ports/ports_test.go` that assume port `10000` as the base
- [x] 2.2 Add a test: project with existing `Start=10000` allocation is reassigned to a range starting at `7001` on next `Allocate`, and the old entry is gone from the registry
- [x] 2.3 Add a test: `nextStart` skips ≥10000 entries — a registry holding one stale `10000` entry plus one fresh `7001` entry allocates the next new project at `7006` (not `10005`)
- [x] 2.4 Add a test: a project with an existing sub-10000 allocation is returned unchanged (no reassignment)
- [x] 2.5 Run `go test ./internal/ports/...`

## 3. Docs & references

- [x] 3.1 Update `assets/asylum-reference.md` to reference the `7001+` range instead of `10000+`
- [x] 3.2 Search the repo for remaining `10000`/`10001`…`10014` mentions in docs/comments and update them (README, mkdocs, CHANGELOG entry, sandbox rules templates if any)
- [x] 3.3 Add a CHANGELOG.md entry under **Unreleased** → **Changed** noting the new base port and one-shot reassignment

## 4. Verification

- [x] 4.1 `go vet ./...` and `go test ./...` pass
- [x] 4.2 Manual check: run asylum in a project that currently has a ≥10000 allocation and confirm the registry entry is replaced with a sub-10000 range on the next session
- [x] 4.3 Run `openspec validate lower-default-port-range --strict`
