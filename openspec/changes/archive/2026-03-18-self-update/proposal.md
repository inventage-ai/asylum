## Why

Users currently update asylum by re-running the install script or manually downloading a binary. There's no way to check for or apply updates from the CLI itself. Additionally, contributors and early adopters have no easy way to run the latest build from `main` without cloning the repo and building from source.

## What Changes

- New `self-update` subcommand that downloads and replaces the running binary with the latest release from GitHub.
- `--dev` flag on `self-update` to pull from a rolling `dev` release built on every push to `main`, instead of the latest tagged release. This is opt-in per invocation — without it, the update always targets the stable channel.
- New `release-channel` config option (`stable` or `dev`). When set to `dev`, `self-update` defaults to the dev channel without needing `--dev` each time. The `--dev` flag still works as an override in either direction.
- New CI workflow that builds and publishes a `dev` release on every push to `main`.
- Install script changed to place the binary in `~/.asylum/bin/` with a symlink from `/usr/local/bin/asylum`. The symlink is the only step that may require `sudo`. Self-update then replaces the binary in `~/.asylum/bin/` without elevated privileges.

## Capabilities

### New Capabilities

- `self-update`: Binary self-update mechanism with stable/dev release channels, GitHub release resolution, and in-place binary replacement.
- `dev-release`: CI workflow that builds and publishes a rolling `dev` pre-release on every push to `main`.

### Modified Capabilities

- `cli-dispatch`: New `self-update` subcommand added to argument parsing and dispatch.
- `config-loading`: New `release-channel` field in the Config struct.

## Impact

- **CLI**: New subcommand `self-update` with `--dev` flag.
- **Config**: New optional `release-channel` field in all config layers.
- **CI**: New workflow for dev releases; existing release workflow unchanged.
- **install.sh**: Changed default install location from `/usr/local/bin` to `~/.asylum/bin/` with a symlink in `/usr/local/bin/`. Existing installs with the binary directly in `/usr/local/bin/` continue to work (self-update replaces wherever the resolved binary path points).
- **Dependencies**: Will use `net/http` and `encoding/json` from stdlib to query the GitHub Releases API. No new external dependencies.
