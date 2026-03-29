# Changelog

## Unreleased

### Changed
- `cleanup` now scopes to the current project by default (removes container, volumes, and project data for the current directory only)
- `cleanup` and `version` are now proper subcommands (`asylum cleanup`, `asylum version`); `--cleanup` and `--version` flags kept as aliases
- Kit activation tiers: `TierAlwaysOn` (shell, node, title), `TierDefault` (docker, java, etc.), `TierOptIn` (apt) replace the boolean `DefaultOn`

### Added
- `cleanup --all` for global cleanup (all images, volumes, cached data) with a confirmation prompt showing exactly what will be deleted
- Documentation site built with MkDocs Material, deployed to GitHub Pages via `.github/workflows/docs.yml`
- Structured docs pages for commands, configuration, kits, concepts, and development
- New `ast-grep` kit: installs ast-grep (`sg`) via npm for AST-based code search, linting, and rewriting
- New `browser` kit: installs Chromium via Playwright for browser automation, with persistent cache volume
- New `cx` kit: installs cx for semantic code navigation (file overviews, symbol search, definitions, references) with configurable language grammars via `packages`
- New `ports` kit (default-on): automatically allocates and forwards a range of high ports per project, with global tracking to prevent collisions
- Kit config sync: new kits are detected on startup and inserted into existing `config.yaml` via `yaml.Node` tree manipulation (preserving comments and user edits)
- Kit state tracking: `~/.asylum/state.json` tracks known kits; new kits trigger activation prompts in interactive mode
- Sandbox rules file injected into containers via `.claude/rules/asylum-sandbox.md`, giving Claude awareness of available tools, kits, sandbox constraints, and Asylum version
- Detailed Asylum reference doc mounted at `.claude/asylum-reference.md` for on-demand troubleshooting and config guidance
- Host IP accessible inside containers via `host.docker.internal` (`--add-host`)
- Kit dependencies: kits can declare `Deps` on other kits (auto-activated at resolve time)
- Default-on kits: `shell`, `node`, and `title` are active unless explicitly disabled with `disabled: true`
- New `github` kit: GitHub CLI (gh) extracted from core Dockerfile
- New `openspec` kit: OpenSpec CLI extracted from node kit, depends on node
- New `shell` kit: oh-my-zsh, tmux config, direnv hooks, terminal size handling extracted from Dockerfile tail
- Kit disabling: `disabled: true` in KitConfig excludes a kit; project config can disable globally-configured kits
- Maven moved to `java/maven` sub-kit (no longer in core apt-get)
- Python build deps (`python3-dev`, `libssl-dev`, etc.) moved to `python` kit
- `self-update` accepts an optional version argument to install a specific release (e.g., `asylum self-update 0.4.0`)
- `selfupdate` accepted as alias for `self-update`

### Fixed
- Global config migration now produces full documented default config with all kits
- Docker kit no longer duplicates GPG key setup already done by core Dockerfile
- Crash with "unknown kit apt" when config contains apt packages or tab-title settings
- E2e test suite with echo agent for full binary lifecycle testing
- `internal/term` package consolidating shared `ShellQuote` and `IsTerminal` helpers

### Changed
- Opencode installed via curl instead of `go install`
- Agent env vars emitted before hardcoded container env vars (Docker last-wins ensures correct precedence)
- Cleanup prompt skipped in non-interactive mode with a warning
- Onboarding prompt skipped in non-interactive mode
- Cleanup now preserves project directories with active sessions
- `docker exec` only uses `-t` flag when stdin is a TTY (fixes non-interactive environments)

### Fixed
- Broken terminal colors in macOS Terminal.app caused by hardcoded `COLORTERM=truecolor`; now inherited from host
- `.tool-versions` Java version now correctly overrides global config but not project-local config
- `ParseVolume` rejects empty host/container paths in all volume spec formats
- `ParseVolume` validates mount options in 3+ part volume specs
- Race condition in session counter file locking (read through locked fd)
- Session counter underflow when increment fails no longer triggers premature container removal
- Cyclic symlinks in `copyDir` no longer cause infinite recursion
- Symlinks to regular files in `copyDir` now resolve to target contents instead of recreating potentially dangling links
- Signal forwarding goroutine no longer leaks after docker process exits
- Shell metacharacters in onboarding command arguments are now properly quoted
- Env var values containing newlines are rejected instead of producing broken Docker flags
- `WriteDefaults` uses O_EXCL to prevent TOCTOU race on first-run config creation
- Self-update HTTP requests have a 60-second timeout (previously no timeout)
- Self-update enforces Content-Length and 512MB size limit on downloads
- Run commands with newlines or empty values are rejected during Dockerfile generation
- Agent error message now lists all registered agents dynamically

## 0.5.0 — 2026-03-24

### Added
- Profile system: language toolchains (java, python, node) are now self-contained profiles that own Dockerfile snippets, entrypoint setup, cache dirs, config defaults, and onboarding tasks
- `profiles` config option (YAML and `--profiles` CLI flag) to select which profiles are active; `nil` = all (backwards compatible), `[]` = none
- Hierarchical sub-profiles: `java/maven`, `java/gradle`, `python/uv`, `node/npm`, `node/pnpm`, `node/yarn`
- Dockerfile and entrypoint split into core + profile snippets + tail, assembled at build time
- Dynamic welcome banner showing versions only for active profiles and installed agents
- Configurable agent CLI installation via `agents` config and `--agents` CLI flag (default: claude only)
- Opencode agent support (`agents: [opencode]`)
- Docker-in-Docker as optional `docker` kit (active by default, remove to disable for smaller images and non-privileged containers)
- First-run onboarding: prompts to mount package manager credentials (Maven) on initial setup
- Project onboarding framework: scans for setup tasks, prompts once, executes via `docker exec` with proper error handling
- Node.js dependency auto-install as first onboarding task (disable with `onboarding: { npm: false }`)
- `--skip-onboarding` CLI flag to skip all onboarding tasks for a single invocation
- Onboarding state tracking in `~/.asylum/projects/` — skips completed tasks unless lockfile changes
- `onboarding` config section for per-task control; `features: { onboarding: false }` for global disable

### Changed
- **BREAKING**: Config format v2 — profiles renamed to kits, per-kit options replace top-level fields (`features`, `packages`, `versions`, `onboarding`). Existing configs are migrated automatically.
- Agents config field is now a map (`agents: {claude:}`) instead of a list
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
