# Spec Delta

## ADDED Requirements

### Requirement: Agent seed config field
The AgentConfig struct SHALL include a `seed` field accepting `none`, `settings` or `full` that controls how the agent's config directory is seeded in `isolated` and `project` mode. When unset, `full` SHALL apply. The field SHALL have no effect in `shared` mode.

#### Scenario: Seed configured per agent
- **WHEN** the config contains `agents: {claude: {config: project, seed: settings}}`
- **THEN** a new per-project Claude config directory receives only Claude's settings allowlist

#### Scenario: Seed in shared mode
- **WHEN** the config contains `agents: {claude: {config: shared, seed: none}}`
- **THEN** the host `~/.claude` is mounted unchanged

### Requirement: Seed honored only in user-maintained configs
`seed` SHALL be honored in `~/.asylum/config.yaml` and the local layer (`.asylum.local` / `.asylum.local.yaml`). In the project layer (`.asylum` / `.asylum.yaml`) it SHALL be ignored and a warning SHALL be logged.

#### Scenario: Seed in the project layer
- **WHEN** the global config sets `seed: settings` for Claude and `.asylum` sets `seed: full`
- **THEN** the effective seed mode is `settings` and a warning names the `.asylum` file

### Requirement: Agent settings allowlist
Each agent with a native config directory SHALL define the allowlist used by `seed: settings`. It SHALL contain user settings, global instructions, user-defined customizations (e.g. agents, commands, skills) and credential files stored in the config directory, and SHALL NOT contain transcripts, history, sessions, todos, caches or logs.

#### Scenario: Claude allowlist excludes session data
- **WHEN** Claude is seeded with `seed: settings`
- **THEN** `projects/`, `history.jsonl`, `todos/` and `shell-snapshots/` are not copied
