# Tasks

## 1. Workspace package

- [x] 1.1 Change `workspace.Resolve` to take an `accept func(string) bool` and a match outcome in place of `reuse bool`. Replace `Reused` with `Continued` and `Attached`, and make `newest` return the newest generated workspace the filter accepts. Verify with table-driven tests covering:
  - the newest accepted workspace wins over a newer rejected one
  - a nil filter gives a fresh workspace even when workspaces exist
  - an accept-all filter still ignores non-matching names and symlinked entries
  - a missing base directory gives a fresh workspace
  - a safe directory stays unchanged

## 2. Docker and main

- [x] 2.1 Add `docker.RunningNames()`, which lists running container names in a single `docker ps` call. Verify with a unit test of the output parser, following the existing `countExecSessions` pattern.
- [x] 2.2 In `cmd/asylum/main.go`, move the guard after `cliFlags` is built and skip it for `update`. Build the filter from `containerMode` and the resume flags, as the design table says, and print the matching warning for each outcome. Verify with `go build ./...`, then run the binary from a temp `HOME` to check four cases: a plain launch is fresh, `self-update` and `update` create no workspace, `shell` with nothing running is fresh, and `--continue` with no session is fresh.
- [x] 2.3 Run `go test ./...` and `go vet ./...`. Both should pass.

## 3. Documentation

- [x] 3.1 Update the "From the Home Directory" section in `docs/concepts/sessions.md` and the sentence in `assets/asylum-reference.md` to describe the per-mode rule. Verify by reading both.
- [x] 3.2 In `assets/asylum-reference.md` and the CHANGELOG trust-dialog entry, replace "the container is the trust boundary" with the reason based on `--dangerously-skip-permissions`. Update the existing Unreleased `--continue` entry to the per-mode rule instead of adding a new one. Verify with `grep -n "trust boundary"`, which should find nothing.

## 4. End-to-end check

- [x] 4.1 On the host, start Claude from `~`, chat without creating files, and exit. Then run `asylum --continue` from `~` and confirm it resumes that session in the same workspace. While a session is running, run `asylum shell` from `~` and confirm it attaches to that workspace.
