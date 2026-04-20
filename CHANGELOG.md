# Changelog

## Unreleased

## 0.6.6 — 2026-04-20

Shared-mode hygiene and kit revivals. Kit-provided Claude skills (`agent-browser`, `ast-grep`) are no longer bind-mounted over the user's `~/.claude/skills/` — they're loaded from inside the container via Claude's `--add-dir` flag, so the host directory stays untouched. The `rtk` kit works again against current upstream rtk releases. Alongside: a new `~/.agents` host mount, `.yaml`-extension config files, a lowered default port range, and several credential and config-merge fixes.

### Added
- Mount host `~/.agents` directory into the container in shared agent mode, so host-installed skills that symlink into `~/.agents/` resolve inside the container (#24)
- Project config files now accept a `.yaml` extension alias (`.asylum.yaml`, `.asylum.local.yaml`) so editors apply YAML syntax highlighting; `.asylum` and `.asylum.local` remain the defaults (#15)
- Security model documentation page describing what asylum protects against and its deliberate non-goals

### Changed
- Ports kit now allocates starting at port 7001 instead of 10000 — most browsers block access to the 10000+ range. Projects with an existing allocation at or above 10000 are automatically reassigned a new range on their next session.

### Fixed
- Kit-provided Claude skills (`agent-browser`, `ast-grep`) no longer create empty directories in the user's host `~/.claude/skills/` in shared agent-config mode. Skills are now staged under `/opt/asylum-skills` inside the container and loaded via `--add-dir`. Users may safely remove any existing empty `~/.claude/skills/agent-browser/` or `~/.claude/skills/ast-grep/` directories left over from previous versions. (#24, #25)
- `rtk` kit works again with recent rtk versions. Newer `rtk init -g` no longer creates a `~/.claude/hooks/` directory and instead expects the hook to be registered as the command `rtk hook claude` in `settings.json`. The kit now follows that pattern: the obsolete hook-script mounting is gone, and the entrypoint registers `rtk hook claude` directly as the PreToolUse hook command. The shared-mode pollution of the host `~/.claude/settings.json` tracked in #29 is not addressed by this change.
- Java kit now honors the configured `versions` list instead of always installing JDK 17, 21, and 25 (#26)
- Kit credential configuration in overlay config files (`.asylum`, `.asylum.local`) was silently dropped during config merge (#28)
- Credential config changes did not trigger the stale container warning because kit credentials were excluded from the config hash (#28)
- Shadow `node_modules` volume chown used a hardcoded `claude` user name instead of the actual container UID, breaking permissions for hosts with a different username

## 0.6.5 — 2026-04-01

macOS binaries are now code-signed and notarized, eliminating Gatekeeper warnings for users downloading asylum from GitHub Releases. All release binaries now include SHA256 checksums and GitHub build provenance attestation for supply chain verification.

### Added
- macOS code signing with Developer ID certificate and Apple notarization for darwin binaries
- SHA256 checksums file (`checksums.txt`) published with every release
- GitHub build provenance attestation for all release binaries
- Checksum verification in install script with graceful fallback

### Changed
- Release and dev-release workflows split into platform-specific build jobs with a shared reusable workflow for darwin signing

## 0.6.4 — 2026-04-01

Fixes the rtk kit failing to build due to the `rtk` binary not being on PATH during Docker image assembly.

### Fixed
- rtk kit Docker build failure — the install script places `rtk` in `~/.local/bin/` which isn't on PATH in non-interactive Docker `RUN` commands

## 0.6.3 — 2026-04-01

Asylum now detects when a running container's image is stale after config changes and automatically restarts it, fixing the common issue where kit packages added to a project config were silently ignored. Also fixes container startup freezes and mise config trust errors.

### Added
- Stale container detection — asylum checks if the running container's image matches the current config and restarts automatically when no active sessions exist, or prompts when sessions are active
- Config drift warning when volumes, env vars, or ports change on a running container

### Fixed
- Kit packages from project config not triggering project image rebuild (#16)
- Container startup appearing to freeze for 60 seconds when the container crashes immediately — now fails fast with logs
- Tab state lost when switching tabs in `asylum config` TUI
- Untrusted `mise.toml` in project directory crashing the entrypoint under `set -e` — mise configs are now auto-trusted

## 0.6.2 — 2026-04-01

Adds `asylum config` for post-setup kit and credential management, replaces the SSH init command with an always-on SSH kit, and renames the browser kit to `agent-browser` backed by Vercel's agent-browser tool.

### Added
- `asylum config` command — interactive tabbed TUI for managing kits, credentials, and isolation settings after initial setup
- SSH is now an always-on kit with configurable `isolation` (isolated/shared/project) — keys are generated automatically on first container start, replacing the manual `asylum ssh-init` command
- New `rtk` kit (opt-in) — installs [RTK](https://github.com/rtk-ai/rtk) token-reduction proxy that compresses shell command output, reducing LLM token usage by 60-90%
- Sandbox rules file lists disabled kits with a reference to the asylum-reference doc for activation instructions

### Changed
- Browser kit renamed from `browser` to `agent-browser`, now backed by [agent-browser](https://github.com/vercel-labs/agent-browser) instead of Playwright — Claude Code skill generated at build time; the old `browser:` config key still works as an alias
- New kit activation prompt uses TUI multiselect instead of per-kit Y/n text prompts — all new kits shown in one batch with descriptions, default-on kits pre-selected
- ast-grep kit now generates and mounts the upstream Claude Code skill for better rule authoring
- Dockerfile instruction ordering optimized for layer caching — faster rebuilds when only later layers change
- `apt` and `shell` kits hidden from interactive selection UIs

### Fixed
- Container not stopping after session exit when a previous session ended abnormally — replaced file-based session counter with runtime exec session detection
- Docker mount failure when agent config dir is a symlink
- Kit activation via `SyncKitToConfig` mangling config.yaml indentation, comments, and whitespace
- Tilde (`~`) not expanded in volume destination paths, causing Docker mount errors
- Release notification dropping last changelog section

### Removed
- `asylum ssh-init` command (replaced by the SSH kit's automatic key generation)

## 0.6.1 — 2026-03-30

Patch release fixing Docker volume mounting issues on some Docker Desktop versions and a spurious kit sync message for new users.

### Fixed
- Nested bind mount failure on some Docker Desktop versions
- Kit sync messages shown to users without a config file

## 0.6.0 — 2026-03-29

This release introduces the kit system — modular, composable tooling profiles that replace hardcoded language toolchains with a flexible, extensible architecture. Each kit owns its Dockerfile snippets, entrypoint setup, config defaults, credentials, and agent rules. Also adds Claude config isolation, a documentation site, and aligns the container user with the host identity.

### Added
- Kit system: modular tooling profiles (java, python, node, docker, github, shell, etc.) that own Dockerfile snippets, entrypoint setup, cache dirs, config defaults, and agent rules — replaces hardcoded language toolchains
- Config format v2: per-kit options replace top-level `features`, `packages`, `versions` fields; v1 configs migrated automatically
- Hierarchical sub-kits: `java/maven`, `java/gradle`, `node/pnpm`, `node/yarn` with kit dependencies and activation tiers (always-on/default/opt-in)
- New kits: `ast-grep` (AST-based code search), `browser` (Chromium via Playwright), `cx` (semantic code navigation), `ports` (automatic per-project port allocation)
- Kit credential system: auto-discovers and filters host credentials by project needs — Maven discovers server IDs from `pom.xml`; GitHub extracts `gh` auth token from host keyrings
- Kit config sync: new kits detected on startup and offered for activation, preserving comments and user edits in config
- Configurable agent CLI installation via `agents` config map and `--agents` flag; Opencode agent support
- Claude config isolation: `shared` (host config), `isolated`, or `project` (per-project) via onboarding wizard or `agents.claude.config`; session detection (`--continue`) works across all modes
- Onboarding wizard: config isolation and credential prompts grouped into a multi-step TUI wizard; declining persists the choice to prevent repeat prompts
- Documentation site with MkDocs Material, versioned via mike with `dev` and `latest` channels
- Sandbox rules and reference doc injected into containers, giving agents awareness of available tools and environment
- `cleanup --all` for global cleanup (all images, volumes, cached data) with confirmation prompt
- Host IP accessible inside containers via `host.docker.internal`
- `self-update` accepts optional version argument (e.g., `asylum self-update 0.4.0`); `selfupdate` accepted as alias
- E2e test suite with echo agent for full binary lifecycle testing

### Changed
- Container user now matches host user (username, UID, GID, home directory) instead of hardcoded `claude:1000:/home/claude`
- Container names now include project name: `asylum-<hash>-<project>` with automatic data migration
- `cleanup` scopes to current project by default; `cleanup` and `version` promoted from flags to subcommands (flag aliases kept)
- Cleanup and onboarding prompts skipped in non-interactive mode; cleanup preserves active sessions
- `docker exec` only uses `-t` flag when stdin is a TTY

### Fixed
- Broken terminal colors in macOS Terminal.app caused by hardcoded `COLORTERM=truecolor`; now inherited from host
- `.tool-versions` Java version correctly overrides global config but not project-local config
- Config file edits no longer strip blank lines and comments via yaml round-tripping
- `ParseVolume` rejects empty host/container paths and validates mount options
- Race condition in session counter file locking; underflow no longer triggers premature container removal
- Cyclic symlinks and symlinks to regular files in `copyDir` handled correctly
- Signal forwarding goroutine no longer leaks after docker process exits
- Shell metacharacters in onboarding commands properly quoted
- Env var values with newlines rejected; run commands with newlines/empty values rejected during Dockerfile generation
- Self-update HTTP requests have 60s timeout and enforce Content-Length/512MB size limit
- `WriteDefaults` uses O_EXCL to prevent TOCTOU race on first-run config creation

## 0.5.0 — 2026-03-24

### Added
- First-run onboarding: prompts to mount package manager credentials (Maven) on initial setup
- Project onboarding framework: scans for setup tasks, prompts once, executes via `docker exec` with proper error handling
- Node.js dependency auto-install as first onboarding task (disable with `onboarding: { npm: false }`)
- `--skip-onboarding` CLI flag to skip all onboarding tasks for a single invocation
- Onboarding state tracking in `~/.asylum/projects/` — skips completed tasks unless lockfile changes
- `onboarding` config section for per-task control; `features: { onboarding: false }` for global disable

### Changed
- Cache directories (npm, pip, maven, gradle) now use named Docker volumes instead of bind mounts for better IO on macOS

### Fixed
- Tilde (`~`) in volume shorthand (e.g. `~/.m2/settings.xml:ro`) now correctly expands to `/home/claude` inside the container instead of the host home directory
- `.tool-versions` with `java 25` no longer triggers "missing" warning (switched from temurin-prefixed to plain version numbers)
- Set `COLORTERM` and `TERM` env vars for proper color support in container

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
