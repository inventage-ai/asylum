# Changelog

## Unreleased

## 0.3.0 — 2026-03-18

### Added
- `self-update` subcommand to update the binary from GitHub releases
- `--dev` flag for `self-update` to track rolling builds from `main`
- `release-channel` config option to permanently select stable or dev channel
- Dev release CI workflow that publishes a `dev` pre-release on every push to `main`
- Install script now places binary in `~/.asylum/bin/` with symlink from `/usr/local/bin/`

### Changed
- Replace SDKMAN with mise for Java and Gradle version management (faster container startup)
- Non-pre-installed Java versions are automatically added to the project Dockerfile

### Fixed
- Auto-restore Claude config from backup when `.claude.json` is missing or lacks auth
- Use `find` instead of `ls` for safer backup file selection in entrypoint
- Suppress "Killed" message when dockerd is terminated on container exit

## 0.2.1 — 2026-03-18

### Fixed
- Kill dockerd immediately on container exit instead of waiting
- Fix fnm PATH in Dockerfile RUN steps and entrypoint

### Changed
- Replace nvm with fnm for Node.js version management

## 0.2.0 — 2026-03-18

### Added
- Integration tests for Docker image, entrypoint, and container behavior
- Strict CLI argument parsing with `run` subcommand and `--` separator
- Git safe.directory wildcard trust in container entrypoint
- Disable git fileMode on Docker Desktop to prevent spurious mode changes

### Fixed
- ParseVolume expanding tilde in container paths
- Numerous robustness fixes: error propagation, input validation, shell injection prevention for package names, deterministic map iteration, symlink handling in copyDir

### Changed
- Extracted `die` helper, inlined redundant variables, removed dead fields
- Consolidated cleanup error handling and image removal

## 0.1.2 — 2026-03-17

### Added
- `--version` flag to CLI

## 0.1.1 — 2026-03-16

### Changed
- Move repo to github.com/inventage-ai/asylum
- Make session detection project-specific for all agents

### Fixed
- Project image running as root when only apt packages configured
- Skip resume on first run when agent config is freshly seeded

## 0.1.0 — 2026-03-15

Initial release.

### Added
- Single Go binary, cross-compiled for linux/darwin on amd64/arm64
- Agent profiles for Claude, Gemini, and Codex
- Layered YAML config system with volume shorthand parsing
- Two-tier Docker image management with hash-based rebuild detection
- Container runtime assembly with volume, env, and port orchestration
- SSH directory setup and key generation
- Colored terminal logging
- GitHub Actions CI/CD and install script
