# Tasks

## 1. Route ModeCommand through a login shell
- [x] 1.1 In `internal/container/container.go`, change the `ModeCommand` branch of `ExecArgs` to shell-quote `opts.ExtraArgs` (via `term.ShellQuoteArgs`), join them, and append `zsh -c "source ~/.zshrc && exec <cmd>"` instead of the bare argv.
- [x] 1.2 Confirm `internal/term` is imported in `container.go` (add if needed). — already imported.

## 2. Tests
- [x] 2.1 Update `TestExecArgsAllModes` in `internal/container/container_test.go`: the `ModeCommand` case now expects the `zsh -c "source ~/.zshrc && exec 'ls' '-la'"` form (each arg is single-quoted by `ShellQuoteArgs`).
- [x] 2.2 Add a case with an argument containing a space to assert quoting is applied.
- [x] 2.3 `go test ./...` passes (also `go vet ./...`).

## 3. Verify end-to-end (manual — needs a real Docker host + base image)
- [x] 3.1 `asylum run claude --version` confirmed working on a real host.
- [x] 3.2 `asylum run ls -la` (and fnm/mise tool) confirmed working on a real host.

## 4. Changelog
- [x] 4.1 Add a **Fixed** entry under **Unreleased** in `CHANGELOG.md`.
