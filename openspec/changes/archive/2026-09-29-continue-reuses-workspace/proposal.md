# Proposal

## Why

Running asylum in the home directory redirects into a fresh workspace under `~/asylum-workspace/`. `asylum --continue` does the same, so it lands in a new, empty workspace where the agent has no session to continue, and the flag fails. Resuming work started from the home directory is impossible without `cd`-ing into the old workspace by hand.

## What Changes

- When the project directory is unsafe and the arguments contain `--continue` or `--resume`, asylum reuses the newest existing workspace instead of creating one.
- "Newest" means the workspace directory with the most recent modification time. Only directories whose names match the generated `<YYYY-MM-DD>-<word>-<word>-<word>` pattern count.
- If no workspace exists yet, asylum creates a fresh one as today.
- The warning names the reused workspace so the user knows where the session runs.
- Launches without a resume flag keep creating a fresh workspace.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `project-dir-guard`: Resume flags reuse the newest workspace instead of always creating a fresh one.

## Impact

- `internal/workspace/workspace.go`: `Resolve` gains a reuse mode.
- `cmd/asylum/main.go`: passes the reuse decision from the parsed args, and adjusts the warning text.
- `CHANGELOG.md`: one **Changed** entry.
