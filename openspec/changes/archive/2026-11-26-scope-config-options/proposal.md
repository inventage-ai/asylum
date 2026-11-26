# Proposal

## Why

All three config files (`~/.asylum/config.yaml`, `.asylum.local`, `.asylum`) accept every option. `.asylum` is usually committed, so a cloned repository can give its container access outside the project directory: host mounts (`volumes`), the host's `~/.ssh` (`kits.ssh.isolation: shared`), the host's agent config (`agents.*.config: shared`), credentials, or a privileged container (`kits.docker`). Two options are already restricted by hand (`env-from-host`, `agents.*.seed`), and the host-mcp kit (PR #50) adds a third ad-hoc mechanism. This change replaces them with one place that defines where each option is accepted. It also makes sure `.asylum.local` stays personal: without that, a repository could commit its own `.asylum.local` and get around every restriction.

## What Changes

- Every config option gets a **scope** that says which files accept it:
  - **any**: global, `.asylum.local`, `.asylum`
  - **personal**: global, `.asylum.local`
  - **global**: `~/.asylum/config.yaml` only
- The rule for **personal**: a project file must not give the container access to anything outside the project directory, or extra privileges.
- **BREAKING:** the following are now personal and ignored in `.asylum`: `volumes`, `agents.*.config`, `agents.*.companions`, `default-resume`, `kits.*.credentials`, `kits.*.isolation`, `kits.*.allow-agent-terminal-title`, `kits.browser-open.schemes`, and enabling the `docker`, `dropshare`, `iterm`, `cx` and `rtk` kits. `env-from-host` and `agents.*.seed` keep their current restriction under the new mechanism.
- **BREAKING:** `release-channel` and `version-check-interval` are global-only.
- Kits declare the scope for enabling the kit and for kit-specific options themselves.
- A test fails when a config field has no scope.
- When a project file sets an option it may not set, asylum shows a blocking prompt before it starts. The prompt lists file, line, option and value, and names the file where the option belongs. The user chooses between starting without these options and aborting. The choice is remembered until the offending entries change. Without a terminal, asylum warns and starts without them.
- Asylum never moves or rewrites options between files.
- Asylum refuses to start when `.asylum.local` (or `.asylum.local.yaml`) is tracked by git.
- If nothing in git ignores `.asylum.local` yet, asylum adds both filenames to the repository's `info/exclude` on startup.
- This repository's `.asylum` volume moves to `.asylum.local`, and the docs no longer show `volumes` in `.asylum`.

Out of scope:
- Stopping the agent from writing config files: the follow-up to #47.
- Removing the `SYS_ADMIN` capability that cx and rtk need: tracked in the backlog.

## Capabilities

### New Capabilities
- `config-layer-scopes`: which config files accept which options, how rejected options are handled at startup, and how the user's decision is remembered.
- `local-config-git-guard`: `.asylum.local` must not be tracked by git and is added to the repository's local excludes.

### Modified Capabilities
- `config-loading`: loading checks the local layer against git before migration, and removes options that are out of scope for their file before merging.
- `env-from-host`: the project-layer restriction is expressed as the personal scope; a project entry now leads to the startup prompt instead of only a warning.
- `config-isolation`: the same change for `agents.*.seed`.
- `browser-open-kit`: `schemes` no longer accumulate from the project layer, only from the global and local layers.

## Impact

- `internal/config`: scope table, YAML node filter, rejection records, git check of the local layer, removal of the two hand-written guards.
- `internal/kit`: `Scope` type and per-kit scope declarations on `docker`, `dropshare`, `iterm`, `cx`, `rtk`.
- `cmd/asylum`: startup prompt, remembered decision, adding to `info/exclude`.
- Docs (`concepts/mounts.md`, `configuration/index.md`, `concepts/security-model.md`, affected kit pages), CHANGELOG (**Changed**), this repository's `.asylum`.
- Users with these options in a committed `.asylum` see the prompt until they move the options to `.asylum.local` or the global config.
- PR #50 (host-mcp): whichever change merges second makes host-mcp declare enabling the kit as personal and `servers` as global, and drops its own ignore logic.
