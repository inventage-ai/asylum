## 1. Config schema

- [x] 1.1 Add `Companions []string` field to `AgentConfig` in `internal/config/config.go` with `yaml:"companions,omitempty"` and `merge:"concat"` (or equivalent — match existing list-merge semantics; verify with `config_test.go` style).
- [x] 1.2 Add accessor `Config.AgentCompanions(name string) []string` returning a de-duplicated, self-reference-stripped list (mirror style of `AgentIsolation`).
- [x] 1.3 Unit tests in `internal/config/config_test.go`: configured list, empty, self-reference stripped, duplicates de-duplicated, last-wins overlay across layers.

## 2. Container assembly — mounts

- [x] 2.1 In `internal/container/container.go` after the primary agent config mount (~line 285–298), iterate `opts.Config.AgentCompanions(opts.Agent.Name())` and for each companion:
  - resolve the companion's `agent.Agent` via `agent.Get(name)`
  - call `agent.ResolveConfigDir(companion, opts.Config.AgentIsolation(companion.Name()), cname)`
  - `os.MkdirAll` host source; `filepath.EvalSymlinks` to canonicalize
  - mount via `vol(hostConfigDir, config.ExpandTilde(companion.ContainerConfigDir(), home), "")`
- [x] 2.2 Ensure companions do not double-mount the primary if the primary appears in its own companion list (covered by self-stripping in `AgentCompanions`).
- [x] 2.3 Confirm sandbox rules (`asylum-sandbox.md`, `asylum-reference.md`) are still dropped only into the primary's `ContainerConfigDir` (`claudeSandboxRulesArgs` is gated on `opts.Agent.Name() == "claude"` and uses `opts.Agent.ContainerConfigDir()` — companions are untouched).

## 3. Container assembly — env vars

- [x] 3.1 In `coreEnvVars` (`container.go:321`), after iterating `opts.Agent.EnvVars()`, iterate companions in declared order and merge their `EnvVars()` with primary-wins semantics on key collisions.
- [x] 3.2 Log a warning via the project `log` package when a collision occurs, naming the key and the companion.

## 4. Validation

- [x] 4.1 Added `agentInstalled` helper in `container.go`; `AgentInstalls []*agent.AgentInstall` is now a `RunOpts` field populated from `main.go`'s resolved install set.
- [x] 4.2 Before mounts are emitted (`coreVolumes`), every companion is validated against the install set; failure returns a clear error naming the missing companion and the configuring agent.
- [x] 4.3 `TestCoreVolumesCompanionNotInstalled` asserts the error path and that the message names the missing companion.

## 5. Tests

- [x] 5.1 `internal/container/container_test.go`: `TestCoreVolumesCompanion{Mount,SharedMount,ProjectMount}` assert the expected `-v` RunArgs for each isolation level.
- [x] 5.2 `TestCoreVolumesCompanionDoesNotAffectAgentsMount` verifies `~/.agents` is not pulled in by a shared companion when the primary is isolated.
- [x] 5.3 `TestCoreEnvVarsCompanion`, `TestCoreEnvVarsCompanionCollisionPrimaryWins` (captures stdout to assert the warning text), and `TestCoreEnvVarsCompanionNoCollisionNoWarning` cover the env var merge and warning behavior. `TestCoreVolumesCompanionDoesNotSuppressAgentsMount` covers the inverse `~/.agents` direction (shared primary + isolated companion).
- [x] 5.4 `TestCoreVolumesCompanionInverseRun` verifies that running codex as primary ignores claude's companion list.

## 6. Documentation & changelog

- [x] 6.1 `CHANGELOG.md` Unreleased entry under **Added**.
- [x] 6.2 Documented in `docs/concepts/agents.md` under a new **Companions** section.
- [x] 6.3 The `agents.md` Companions section explicitly cites the codex Claude Code plugin use case and shows the YAML example.

## 7. Manual verification

- [x] 7.1 In a real project: set `agents.claude.companions: [codex]`, ensure codex is in agent installs, run `asylum claude`, verify `~/.codex` is mounted and writable inside the container.
- [x] 7.2 Run `asylum codex` in the same project: verify claude's config is NOT mounted (one-directional check).
- [x] 7.3 Remove codex from the install set, keep `companions: [codex]`, run `asylum claude`: verify the run fails with the expected error.
- [x] 7.4 Run a project without `companions`: verify behavior identical to before (diff `docker inspect` mounts against a pre-change baseline).
