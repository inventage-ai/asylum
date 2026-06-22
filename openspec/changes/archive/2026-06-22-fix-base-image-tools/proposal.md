# Proposal: Fix base image CLI tools (fd, file)

## Why

Two common CLI tools are not usable in the container under their canonical names:

- **`fd`** is not on PATH. The base image installs Debian's `fd-find` package, which ships the binary as `fdfind` (renamed to avoid an old packaging clash). Agents and users invoke `fd` and get "command not found".
- **`file`** is not installed at all — it was never in the apt list — so type-detection invocations fail.

(The originally-reported "`rg` shadowed by grep" was investigated and dropped from scope: in-container, `rg` resolves cleanly to `/usr/bin/rg`. The `grep` override is a shell function injected by Claude Code's bundled-grep feature, not by Asylum — Asylum's Dockerfile/entrypoint never touch `grep` or `rg`.)

## What Changes

- Add `file` to the apt install list in `assets/Dockerfile.core`.
- Make `fd` available under its canonical name by symlinking it to the `fdfind` binary (same pattern Debian users apply for `bat`/`batcat`).

## Capabilities

### New Capabilities
- (none)

### Modified Capabilities
- `container-image`: Add a requirement that the base image provides core CLI tools under their canonical command names (`fd`, `file`, alongside the existing `rg`).

## Impact

- Affected code: `assets/Dockerfile.core` (apt list + symlink).
- Rebuild: changing the base Dockerfile invalidates the base image hash, triggering a base-image rebuild that cascades to project images. No config or runtime-flag changes.
- No new dependencies; `fd-find` is already installed.
