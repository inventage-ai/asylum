## Why

With shadow volumes, `node_modules` inside the container always starts empty. Agents that work on Node.js projects fail immediately because dependencies aren't installed. Users must manually run `npm ci` or similar before the agent can work.

## What Changes

- In agent mode, before the agent starts, asylum detects Node.js projects with lockfiles and prompts the user to install dependencies.
- The prompt happens on the host (Go), the install runs inside the container as part of the `docker exec` session.
- A feature flag `auto-install-node-modules` (default on) allows disabling this behavior.
- Shadow volumes created by Docker as root are chowned to the container user after container start.

## Capabilities

### New Capabilities

- `auto-install-node-modules`: Detect lockfiles, prompt user, install Node.js dependencies inside the container before the agent starts.

### Modified Capabilities

- `container-assembly`: `ExecArgs` gains `PreRunCmds` support to prepend shell commands before the agent binary. Shadow volume ownership is fixed after container start.

## Impact

- **CLI**: New prompt in agent mode when Node.js lockfiles are detected.
- **Container exec**: Agent command wrapped in `bash -c` when pre-run commands exist.
- **Docker**: New `docker.Exec` helper for running commands as a specific user.
- **Config**: Uses existing `features` map for opt-out.
