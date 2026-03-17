## Why

Git refuses to operate on mounted project directories inside the container with `fatal: detected dubious ownership`. This happens because git's `safe.directory` check sees a mismatch between the repository owner and the current user — even when UIDs are aligned numerically, the container's user context differs from the host's.

This blocks agents from using git (commits, status, diff) which is essential for their workflow.

## What Changes

- Add `git config --global --add safe.directory '*'` to the entrypoint before any agent or shell runs

The wildcard is appropriate here: containers are ephemeral, and the only mounted repositories are ones the user explicitly chose to mount. There is no untrusted code scenario.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `container-assets`: Entrypoint adds global git safe.directory wildcard trust

## Impact

- Modifies `assets/entrypoint.sh` only
- No config or CLI changes
- Fixes git operations for all agents and shell mode
