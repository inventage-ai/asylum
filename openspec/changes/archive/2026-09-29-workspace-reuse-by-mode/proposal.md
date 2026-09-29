# Proposal

## Why

`asylum --continue` from the home directory reuses the most recently modified workspace, but it doesn't check whether that workspace has anything to continue. A plain launch where the user only chatted leaves an empty workspace that becomes the newest one, so the next `--continue` finds no session. Several related gaps make this worse:
- `asylum self-update` and `asylum update` from `~` create a workspace, even though the spec says the guard only runs on the container-run path.
- A literal `--continue` anywhere in the arguments, for example `asylum run git rebase --continue`, triggers reuse.
- `asylum shell` and `asylum run` from `~` always get a fresh workspace, even when the user has one running.

## What Changes

- Asylum picks a workspace by mode. Among the generated workspaces that qualify, the most recently modified one wins, and asylum creates a fresh one when none qualifies:
  - A plain `asylum` launch always creates a fresh workspace.
  - `asylum --continue` or `--resume` in agent mode reuses the newest workspace where the active agent has a session.
  - `asylum shell` and `asylum run …` attach to the newest workspace whose container is running.
- `self-update` and `update` no longer run the guard.
- The warning says whether the workspace is fresh, continued or attached.
- Two lines that call the container "the trust boundary" get reworded, because the security model says Docker is not a security boundary. These are in the in-container reference and the CHANGELOG entry for skipping the trust dialog.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `project-dir-guard`: Reuse is chosen per mode, resume only picks workspaces with a session, shell and run attach to a running workspace, and `self-update` and `update` skip the guard.

## Impact

- `internal/workspace/workspace.go`: `Resolve` takes a filter function instead of a reuse flag.
- `internal/docker/docker.go`: a function that lists running container names in one call.
- `cmd/asylum/main.go`: builds the filter per mode and skips the guard for `self-update` and `update`.
- Docs: `docs/concepts/sessions.md`, `assets/asylum-reference.md`, `CHANGELOG.md`. The existing Unreleased entries are updated, with no new ones.
