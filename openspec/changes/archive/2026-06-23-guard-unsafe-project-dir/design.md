## Context

asylum mounts the project directory at its real host path with `:z` (`internal/container/container.go:254`) and sets it as the working directory (`container.go:78`). The project directory is resolved once, early, via `projectDir := filepath.Abs(".")` (`cmd/asylum/main.go:65`), and threaded through config loading, container assembly, and the container name (`ContainerName(projectDir)`).

When that path is the user's home directory or `/`, the `:z` relabel recurses over an enormous tree and the sandbox inherits the user's whole environment — a footgun that currently surfaces only as a slow/failing `docker run`. The `cleanup` subcommand resolves the project directory the same way (`main.go:743`) but must not create directories.

## Goals / Non-Goals

**Goals:**
- Catch the unsafe-directory case before any Docker invocation.
- Transparently redirect to a fresh, isolated workspace so the user can keep going without manually `cd`-ing.
- Keep the change additive and localized — no config schema or dependency changes.

**Non-Goals:**
- Reusing or resuming a previous workspace (every unsafe launch is fresh by decision).
- Initializing a git repo or seeding the workspace.
- Changing the user's shell working directory (a binary cannot; the shell stays in the original dir).
- Guarding subcommands other than the run path.

## Decisions

**Single guard helper, run-path only.** A helper — e.g. `workspace.Resolve(projectDir, home) (string, redirected bool, err error)` in a new `internal/workspace/` package — is called immediately after `filepath.Abs(".")` on the run path in `main.go`. Everything downstream (config, container name, mounts, `-w`) uses the returned path unchanged, so the redirect is invisible to the rest of the pipeline. `cleanup` does not call it, satisfying the run-path-only requirement.

*Alternative considered:* redirect inside container assembly. Rejected — the project dir also feeds the container name and config layering, which happen earlier; redirecting late would desync them.

**Unsafe = exact home or `/`.** Compare the cleaned absolute path against `os.UserHomeDir()` and the filesystem root. Subdirectories of home stay safe. A tight denylist covers the real footguns without surprising users who keep projects in unusual places.

*Alternative considered:* a broader denylist (`/tmp`, `/etc`, `/usr`, …). Deferred — easy to extend the predicate later if needed; starting tight avoids false positives.

**Workspace layout `~/asylum-workspace/<YYYY-MM-DD>-<w1>-<w2>-<w3>/`.** The parent lives under home, but only the small leaf subdirectory is mounted, so `:z` relabels just that subtree — the original problem disappears. Date prefix from `time.Now().Format("2006-01-02")`; three words picked with `math/rand` from a wordlist embedded via `go:embed` (`assets/`). If the generated path already exists, re-roll the words.

*Alternative considered:* short random hash suffix. Rejected by decision — memorable words are friendlier; the wordlist is a static asset with no maintenance cost.

**Loud announcement via the `log` package.** On redirect, emit a `log.Warn` line naming the absolute workspace path so the user knows where their work lives. This reuses existing user-facing output conventions.

## Risks / Trade-offs

- **User confusion about where work landed** → Mitigated by the loud announce line naming the absolute path.
- **No resume across home launches** → Accepted by decision; container name derives from the (fresh) path, so each launch is a clean session. Acceptable for the "just trying an agent" case that triggers this path.
- **Same-day name collision** → Re-roll words when the target path already exists; collision probability is already negligible.
- **Wordlist adds an embedded asset** → Tiny, static, consistent with existing `go:embed` usage in `assets/`.

## Migration Plan

Purely additive; no migration. Existing behavior in safe directories is unchanged. Rollback is removing the guard call — downstream code is untouched.

## Open Questions

None — all decisions resolved during exploration (fresh-every-launch, auto-redirect + announce, home + `/` scope, embedded wordlist, no `git init`, cleanup untouched).
