# Proposal

## Why

In the `isolated` and `project` config modes, asylum seeds the agent's config directory by copying the entire host config directory (e.g. `~/.claude`), including transcripts, history and todos of every project. This defeats the purpose of `project` mode, and the only workaround is a shell function that recreates asylum's container-name formula and creates the directory empty beforehand (GitHub issue #39).

## What Changes

- New per-agent option `agents.<agent>.seed` with the values `none` (empty directory), `settings` (copy only an agent-specific allowlist of settings, instructions, customizations and credentials) and `full` (copy everything, the current behavior and default).
- Each agent defines its own allowlist next to its native config directory.
- `seed` is honored only in user-maintained configs (`~/.asylum/config.yaml`, `.asylum.local`). In the repo-committed `.asylum` it is ignored with a warning, so a repository cannot override the user's seed choice.
- An unknown `seed` value aborts with an error naming the valid values.

## Capabilities

### New Capabilities

### Modified Capabilities
- `container-assembly`: the "Agent config seeding" requirement now depends on the seed mode and covers both `isolated` and `project` targets.
- `config-isolation`: adds the `seed` field and its restriction to user-maintained config layers.

## Impact

- `internal/config`: `AgentConfig.Seed`, project-layer filtering in `Load`.
- `internal/agent`: new `SeedSettings()` method on the `Agent` interface with an allowlist per agent.
- `internal/container`: `EnsureAgentConfig`/`EnsureAgentConfigAt` take the seed mode.
- `cmd/asylum/main.go`: passes the configured mode.
- Docs: `docs/concepts/isolation.md`; `CHANGELOG.md`.
- Not addressed: credentials stored outside the config directory (e.g. Claude on macOS uses the Keychain, #34).
