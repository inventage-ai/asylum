# Changelog

## Unreleased

### Added
- Project onboarding framework: scans for setup tasks, prompts once, executes via `docker exec` with proper error handling
- Node.js dependency auto-install as first onboarding task (disable with `onboarding: { npm: false }`)
- `--skip-onboarding` CLI flag to skip all onboarding tasks for a single invocation
- Onboarding state tracking in `~/.asylum/projects/` — skips completed tasks unless lockfile changes
- `onboarding` config section for per-task control; `features: { onboarding: false }` for global disable

## 0.4.0 — 2026-03-19

### Added
- Shadow `node_modules` with named Docker volumes to isolate container from host-built native binaries (disable with `features: { shadow-node-modules: false }`)
- `--cleanup` now also removes asylum-managed Docker volumes
- Multiple concurrent sessions per project — all modes (agent, shell, run) exec into a running container
- Container automatically cleaned up when the last session exits (file-based session counter)
- Integration tests for detached container lifecycle and multi-session behavior

### Changed
- Container starts detached with idle process; all sessions use `docker exec` instead of `docker run`

### Fixed
- fnm not found in interactive shell (missing from PATH in bashrc/zshrc)

## 0.3.3 — 2026-03-18

### Added
- Set terminal tab title on container start (default `🤖 projectname`, configurable via `tab-title` with `{project}`, `{agent}`, `{mode}` placeholders)
- Dev channel self-update shows commit count and recent commit messages
- `self-update --safe` emergency updater that bypasses all checks
- `features` config section for boolean feature flags
- `session-name` feature flag: names new Claude sessions after the project directory (opt-in)
- `allow-agent-terminal-title` feature flag: lets the agent control the terminal tab title (opt-in, default: asylum controls it)

## 0.3.2 — 2026-03-18

### Added
- MIT license file
- `env` config field (`map[string]string`) for setting arbitrary environment variables via config layers
- `-e KEY=VALUE` CLI flag for setting environment variables (repeatable, highest precedence)

### Fixed
- Self-update always skipping download on dev channel

## 0.3.1 — 2026-03-18

### Added
- `asylum shell` and `asylum run` exec into a running container instead of failing with a name conflict
- Read Java version from `.tool-versions` (mise/asdf) into config automatically
- Mount git worktree directories so git works inside containers for worktree checkouts
- Show commit hash in `--version` and `self-update` output

### Changed
- Non-pre-installed Java versions activated in entrypoint instead of showing a warning
- Only remove containers started during the session on dockerd exit (preserve pre-existing)

### Fixed
- Fix mise not found in project Dockerfile (`$HOME/.local/bin/mise` full path)
- Fix Java version format in project Dockerfile (pass as-is, don't prepend `temurin-`)
- Fix self-update showing branch name instead of commit hash
- Validate port values before passing to Docker
- Validate Java version before interpolation into Dockerfile
- Error when `run` subcommand has no command
- Merge `known_hosts` instead of overwriting on `ssh-init`
- Handle tab-separated `.tool-versions` for Java detection
- Skip files deleted between WalkDir and Info in copyDir
- Verify download size against Content-Length before replacing binary
- Close response body on non-200 status in fetchTagCommit
- Check version for dev channel to skip redundant downloads
- Numerous robustness fixes: deterministic env var ordering, slice mutation safety, partial cleanup reporting

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
