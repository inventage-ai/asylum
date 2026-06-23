## Why

Running `asylum` from the user's home directory (or filesystem root) bind-mounts that entire tree into the container with the `:z` flag, which recursively relabels every file on SELinux hosts and drags the user's whole environment into the sandbox — slow, destructive, and defeating the purpose of an isolated workspace. This fails *after* `docker run` is attempted; it should be caught before, with a safe fallback.

## What Changes

- Detect when the resolved project directory is unsafe to sandbox — the exact home directory or the filesystem root `/` — before any Docker invocation.
- When unsafe, auto-create a fresh workspace directory under `~/asylum-workspace/<YYYY-MM-DD>-<three-random-words>/`, swap the project directory to it, and run the container there instead.
- Announce the redirect loudly (a warning line naming the created workspace path) so the user knows where their work lives.
- Generate the three words from a small wordlist embedded via `go:embed`.
- A fresh workspace is created on every home/root launch (no reuse, no cross-launch resume); the created directory is left empty (no `git init`).
- The guard runs only on the container-run path (right after `projectDir := filepath.Abs(".")` in `cmd/asylum/main.go`); the `cleanup` subcommand is unaffected and never creates a workspace.

## Capabilities

### New Capabilities
- `project-dir-guard`: Detection of unsafe project directories and redirection to a freshly created dated workspace before the container is assembled.

### Modified Capabilities
<!-- None: this is additive and lives before container assembly; no existing requirement changes. -->

## Impact

- `cmd/asylum/main.go`: new guard helper invoked after project-dir resolution on the run path.
- New embedded wordlist asset (`assets/`) plus `go:embed` declaration.
- New package or file for workspace-name generation and unsafe-dir detection (e.g. `internal/workspace/`).
- No config schema changes, no new dependencies; `cleanup` path untouched.
