## Context

Asylum is distributed as a single Go binary, cross-compiled for four targets. Users install via `install.sh` which downloads the latest GitHub release. There is no mechanism for the binary to update itself, and no way to track the `main` branch without building from source.

The `version` variable is set at build time via `-ldflags -X main.version=<version>`. Tagged releases produce semver versions (e.g., `1.2.0`); the default is `"dev"`.

## Goals / Non-Goals

**Goals:**
- Users can update asylum to the latest stable release with `asylum self-update`.
- Users can opt into a rolling dev channel built from `main` on every push.
- The channel preference can be persisted in config (`release-channel: dev`) or overridden per-invocation (`--dev`).
- The update is atomic — a failed download doesn't corrupt the running binary.

**Non-Goals:**
- Auto-update on launch or background update checks.
- Update notifications or version pinning.
- Signature verification of downloaded binaries (can be added later).
- Updating asylum inside the container (it runs on the host).

## Decisions

### 1. New `internal/selfupdate` package

A new package `internal/selfupdate` handles GitHub API interaction, binary download, and replacement. This keeps the update logic out of `main.go` and testable in isolation.

**Why not put it in `cmd/asylum/`?** The main package is already the largest file; a separate package with a clean API (`selfupdate.Run(opts)`) keeps main.go focused on dispatch.

### 2. GitHub Releases API for version resolution

- **Stable channel**: `GET /repos/{owner}/{repo}/releases/latest` → returns the latest non-prerelease tag.
- **Dev channel**: `GET /repos/{owner}/{repo}/releases/tags/dev` → returns the rolling dev pre-release.

Both endpoints return JSON with an `assets` array. We match the binary name by `asylum-{os}-{arch}`.

**Why not download from a fixed URL?** The API gives us the version tag, which we use to skip no-op updates and display what was installed.

### 3. Install location: `~/.asylum/bin/` with symlink

The install script places the binary in `~/.asylum/bin/asylum` and creates a symlink at `/usr/local/bin/asylum` pointing to it. This gives us:
- A user-writable location for the binary (no sudo for self-update).
- A well-known PATH entry via the symlink (no dotfile editing needed).
- The symlink creation is the only step that may require `sudo` — a one-time cost at install.

Self-update resolves the real path of the running binary (following symlinks via `os.Executable()` + `filepath.EvalSymlinks()`) and replaces the target. This means it works regardless of whether the user installed to `~/.asylum/bin/` (new) or `/usr/local/bin/` directly (legacy).

### 4. Atomic binary replacement

1. Download to a temp file in the same directory as the resolved binary (ensures same filesystem).
2. `chmod +x` the temp file.
3. `os.Rename(tmp, resolvedBinary)` — atomic on POSIX.

**Why same directory?** `os.Rename` doesn't work across filesystem boundaries. Writing the temp file next to the target guarantees they're on the same mount.

### 5. Dev release as a GitHub pre-release

A new CI workflow (`.github/workflows/dev-release.yml`) runs on every push to `main`:
1. Deletes the existing `dev` tag and release (if any).
2. Rebuilds all four binaries with `VERSION=dev`.
3. Creates a new pre-release tagged `dev` with the binaries attached.

**Why a single rolling tag?** It avoids accumulating dev releases. Users always get the latest `main` build. The `dev` tag is force-pushed each time.

**Why pre-release?** So `releases/latest` still returns the stable release, and the install script is unaffected.

### 6. Channel resolution logic

```
effective_channel = --dev flag ? "dev" : config.release-channel ? config.release-channel : "stable"
```

The `--dev` CLI flag always wins. Without it, the config value is used. Without either, stable is the default. There is no `--stable` flag — omitting `--dev` is sufficient.

### 7. Skip update when already current

Compare the resolved remote version tag against the running `version` variable. If they match, print a message and exit. For dev channel, always update (dev builds don't have meaningful version comparison).

### 8. `self-update` is a subcommand, not a flag

Like `shell`, `run`, and `ssh-init`, `self-update` is a positional subcommand. It accepts `--dev` as its only flag. This is consistent with the existing CLI grammar and avoids adding more top-level flags.

## Risks / Trade-offs

- **Permission errors**: For new installs (`~/.asylum/bin/`), self-update needs no elevated privileges. For legacy installs where the binary lives directly in `/usr/local/bin`, the rename may fail with EPERM — we print a clear message suggesting `sudo asylum self-update`. We do not auto-escalate.
- **Dev channel instability**: Dev builds are unversioned and untested beyond CI. The opt-in friction (`--dev` or explicit config) is intentional.
- **GitHub API rate limiting**: Unauthenticated requests are limited to 60/hour. Sufficient for self-update, but could be hit in CI environments. Not worth solving now.
- **No rollback**: If the new binary is broken, there's no built-in rollback. Users can reinstall a specific version via `install.sh <version>`. Acceptable given the simplicity of the tool.
