## 1. Wordlist asset

- [x] 1.1 Add a small curated wordlist file under `assets/` (lowercase, one word per line)
- [x] 1.2 Embed it via `go:embed` (extend `assets/assets.go`)

## 2. Workspace package

- [x] 2.1 Create `internal/workspace/` with an unsafe-dir predicate (exact home dir or filesystem root `/`)
- [x] 2.2 Add workspace-name generation: `<YYYY-MM-DD>-<w1>-<w2>-<w3>` from the embedded wordlist (`math/rand`)
- [x] 2.3 Add `Resolve(projectDir, home) (string, bool, error)`: returns the original path when safe; when unsafe, creates `~/asylum-workspace/<name>/` (re-rolling on collision) and returns the new path plus a redirected flag
- [x] 2.4 Add table-driven tests: safe dir unchanged, home subdir safe, home redirected, `/` redirected, name format, collision re-roll

## 3. Wire into the run path

- [x] 3.1 In `cmd/asylum/main.go`, call `workspace.Resolve` immediately after `projectDir := filepath.Abs(".")` on the run path
- [x] 3.2 On redirect, emit a `log.Warn` line naming the absolute workspace path
- [x] 3.3 Confirm the `cleanup` path (`main.go:743`) does NOT call the guard

## 4. Verify

- [x] 4.1 `go test ./...` and `go vet ./...` pass
- [x] 4.2 Manual check: `cd ~ && asylum` redirects and announces; `cd ~/some-project && asylum` is unchanged
- [x] 4.3 Add a CHANGELOG.md entry under Unreleased (Added)
