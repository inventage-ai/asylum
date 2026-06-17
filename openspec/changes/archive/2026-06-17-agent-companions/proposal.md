## Why

Claude Code plugins like the codex Claude Code plugin invoke a second agent's CLI from inside the container. That binary needs its own config directory mounted and its environment variables set (e.g. `CODEX_HOME` for codex), but today only the launched agent's config is mounted — even if other agents are installed in the image. There is no way to express "while claude is running, also make codex's config available."

## What Changes

- Add an optional `companions` list to per-agent config (`agents.<name>.companions: [<name>, ...]`).
- When the listed (primary) agent is launched, each companion's config dir SHALL be mounted into the container at its `ContainerConfigDir`, writable, using the **companion's own** isolation setting (shared/isolated/project) to resolve the host source path.
- Each companion's `EnvVars()` SHALL be merged into the container environment alongside the primary's.
- Companions SHALL be one-directional: `agents.claude.companions: [codex]` does not imply `agents.codex.companions: [claude]`.
- The launched agent (`Command()`, `HasSession()`, resume semantics) is unchanged — companions never launch.
- If a listed companion is not an installed agent for this project (no `AgentInstall` for it), startup SHALL fail with a clear error. No auto-install.
- Cycles are harmless (the companion set is collected once); a self-reference SHALL be ignored.

## Capabilities

### New Capabilities
- `agent-companions`: companion config-dir mounts and env-var contribution for non-launched agents, configured per primary agent in YAML.

### Modified Capabilities
<!-- none — companions add new behavior on top of agent-interface, config-isolation, and container-assembly without changing their existing requirements -->

## Impact

- **Code**:
  - `internal/config/config.go`: add `Companions []string` field on `AgentConfig`; add accessor.
  - `internal/container/container.go` (around line 285): after the primary agent's config mount, iterate companions and mount each one's resolved config dir at the companion's `ContainerConfigDir`.
  - `internal/container/container.go` (coreEnvVars, around line 327): merge each companion's `EnvVars()` into the container env. On key collisions, the primary wins.
  - Validation: at run assembly, error if any companion is not present in the installed-agents set (i.e. has no registered `AgentInstall`).
- **Config schema**: additive (`agents.<name>.companions`). No migration needed.
- **Docs**: `CHANGELOG.md` Unreleased entry; mention in agent-related reference if applicable.
- **Behavior preserved**: `~/.agents` shared-mode mount is unchanged. Single-agent runs without `companions` behave exactly as today.
