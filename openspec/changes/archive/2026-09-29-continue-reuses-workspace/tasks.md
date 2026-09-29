# Tasks

## 1. Workspace reuse

- [x] 1.1 Add a reuse mode to `workspace.Resolve`. It returns the newest directory under `~/asylum-workspace/` whose name matches the generated pattern, and falls back to a fresh workspace. It also reports whether the result was reused. Verify with table-driven tests covering: newest by mtime wins, a non-matching name is ignored even if newer, a missing base dir falls back to fresh, and a safe dir is unchanged.
- [x] 1.2 In `cmd/asylum/main.go`, set reuse when `extraArgs` contains `--continue` or `--resume`, and print "Continuing in the newest workspace:" for a reused workspace. Verify with `go build ./...`, then check both messages by running the binary from a temp `HOME`.
- [x] 1.3 Run `go test ./...` and `go vet ./...`. Both should pass.

## 2. Documentation

- [x] 2.1 Find where the docs and the in-container reference describe the home-directory workspace, and note that `--continue`/`--resume` reuse the newest one. Verify by reading the pages.
- [x] 2.2 Add a **Changed** entry to `CHANGELOG.md` under Unreleased. Verify that it names both flags and the newest-workspace rule.
