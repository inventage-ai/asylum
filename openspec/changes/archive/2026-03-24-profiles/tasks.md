## 1. Profile Package Foundation

- [x] 1.1 Create `internal/profile/profile.go` with the Profile struct (Name, Description, DockerSnippet, EntrypointSnippet, CacheDirs, Config, OnboardingTasks, SubProfiles), registry map, and Resolve function (parsing `parent/child` syntax, hierarchical activation, deduplication, deterministic ordering)
- [x] 1.2 Write tests for profile resolution: top-level activates all children, `parent/child` activates parent + child only, deduplication, unknown profile errors, nil-means-all default, empty-means-none

## 2. Built-in Profile Definitions

- [x] 2.1 Create `internal/profile/java.go` — java profile with DockerSnippet (mise install java@17/21/25), EntrypointSnippet (mise activation, ASYLUM_JAVA_VERSION handling), Config defaults (versions.java: 21); maven sub-profile with CacheDirs (.m2); gradle sub-profile with DockerSnippet (mise install gradle), CacheDirs (.gradle)
- [x] 2.2 Create `internal/profile/python.go` — python profile with DockerSnippet (uv tool installs: black, ruff, mypy, pytest, ipython, poetry, pipenv), EntrypointSnippet (none at parent level), Config defaults; uv sub-profile with EntrypointSnippet (venv auto-creation for Python projects), CacheDirs (.cache/pip)
- [x] 2.3 Create `internal/profile/node.go` — node profile with DockerSnippet (fnm install, Node.js global packages: typescript, ts-node, eslint, prettier, nodemon), EntrypointSnippet (fnm env setup); npm sub-profile with CacheDirs (.npm), OnboardingTasks (move existing NPMTask here); pnpm sub-profile with DockerSnippet (pnpm global install); yarn sub-profile with DockerSnippet (yarn global install)

## 3. Dockerfile Decomposition

- [x] 3.1 Split `assets/Dockerfile` into `assets/Dockerfile.core` (OS packages, Docker, GH/GL CLIs, user creation, mise/fnm/uv manager-only installation, agent CLIs) and `assets/Dockerfile.tail` (oh-my-zsh, shell config, git config, tmux, entrypoint COPY, final USER/WORKDIR)
- [x] 3.2 Update `assets/assets.go` with new `go:embed` declarations for Dockerfile.core and Dockerfile.tail
- [x] 3.3 Verify: when all profiles are active, the assembled Dockerfile (core + all profile DockerSnippets + tail) produces equivalent output to the current monolithic Dockerfile

## 4. Entrypoint Decomposition

- [x] 4.1 Split `assets/entrypoint.sh` into `assets/entrypoint.core` (PATH setup, git config, SSH, direnv, Docker daemon) and `assets/entrypoint.tail` (welcome banner framework, exec)
- [x] 4.2 Update `assets/assets.go` with new `go:embed` declarations for entrypoint fragments
- [x] 4.3 Make welcome banner in entrypoint.tail dynamic — only show version lines for tools whose profile is active (assembled at build time from profile EntrypointSnippets)

## 5. Config Integration

- [x] 5.1 Add `Profiles []string` field to Config struct in `internal/config/config.go` with YAML tag `profiles`; use `*[]string` or sentinel to distinguish nil (unspecified = all) from empty (= none)
- [x] 5.2 Update config merge logic: `profiles` follows last-wins semantics (later layer replaces, no list concatenation)
- [x] 5.3 Add `--profiles` CLI flag to `cmd/asylum/main.go` (comma-separated), wire into CLIFlags and config override
- [x] 5.4 Implement profile config default injection: after loading each config layer that specifies profiles, resolve the profiles and merge their Config defaults before applying the next layer

## 6. Image Build Integration

- [x] 6.1 Update `image.EnsureBase` to accept resolved global profiles, assemble Dockerfile from core + profile DockerSnippets + tail, assemble entrypoint from core + profile EntrypointSnippets + tail
- [x] 6.2 Update base image hash computation to include profile snippets and assembled entrypoint content
- [x] 6.3 Update `image.EnsureProject` to accept project-level profiles (those not in global set), prepend their DockerSnippets before existing package installation logic
- [x] 6.4 Update project image hash to include project-level profile snippets

## 7. Container Setup Integration

- [x] 7.1 Replace hardcoded `CacheDirs` in `container.go` with a function that aggregates CacheDirs from active profiles (union of global + project profiles)
- [x] 7.2 Update `appendVolumes` to use dynamic cache dirs
- [x] 7.3 Update cache migration logic in `main.go` to use the dynamic cache dir map
- [x] 7.4 Update `appendEnvVars` to include env vars from profile Config defaults (already handled by config merge, but verify)

## 8. Onboarding Integration

- [x] 8.1 Move `NPMTask` from `internal/onboarding/npm.go` into the node/npm profile's OnboardingTasks field
- [x] 8.2 Update `main.go` onboarding call: collect OnboardingTasks from all active profiles instead of hardcoding `[]onboarding.Task{onboarding.NPMTask{}}`
- [x] 8.3 Verify per-task enable/disable via `onboarding` config still works with profile-sourced tasks

## 9. Wiring and Main

- [x] 9.1 Update `cmd/asylum/main.go`: resolve profiles from merged config, pass global profiles to `EnsureBase`, compute project-only profiles, pass to `EnsureProject`, pass all active profiles to container setup and onboarding
- [x] 9.2 Update `printUsage()` to document `--profiles` flag
- [x] 9.3 Add CHANGELOG entry under Unreleased

## 10. Testing

- [x] 10.1 Unit tests for profile resolution (task 1.2, if not already done inline)
- [x] 10.2 Unit tests for Dockerfile assembly: verify core + snippets + tail concatenation, verify empty profiles produces core + tail only
- [x] 10.3 Unit tests for entrypoint assembly
- [x] 10.4 Unit tests for config merge with profiles: nil defaults to all, empty means none, last-wins semantics, profile config defaults overridden by project config
- [x] 10.5 Unit tests for dynamic CacheDirs aggregation
- [x] 10.6 Integration test: build with all profiles (default) and verify container has java, python, node available — should match current behavior exactly
