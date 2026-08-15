# config-isolation Specification

## Purpose

Controls where an agent's config directory comes from: `shared` uses the host's directly, `isolated` an asylum-managed copy shared across projects, `project` a per-project copy. The level is set per agent via `agents.<agent>.config`, prompted for on first run with Claude and persisted, with a defined `shared` fallback for when no layer and no wizard supplied a value.

## Requirements

### Requirement: Agent isolation config field
The AgentConfig struct SHALL include a `Config` field accepting values `shared`, `isolated`, or `project` to control the agent's config directory isolation level.

#### Scenario: AgentConfig with isolation level
- **WHEN** config YAML contains `agents: { claude: { config: shared } }`
- **THEN** the parsed AgentConfig has Config set to `"shared"`

#### Scenario: AgentConfig without isolation level
- **WHEN** config YAML contains `agents: { claude: {} }` or `agents: { claude: }`
- **THEN** the parsed AgentConfig has Config as empty string (triggers prompt on first run)

### Requirement: Config isolation levels
The system SHALL support three agent config isolation levels: `shared` (host config), `isolated` (asylum-managed, shared across projects), and `project` (per-project isolation).

#### Scenario: Shared mode
- **WHEN** `agents.claude.config` is set to `shared`
- **THEN** the host's `~/.claude` is mounted directly into the container

#### Scenario: Isolated mode
- **WHEN** `agents.claude.config` is set to `isolated`
- **THEN** `~/.asylum/agents/claude/` is mounted into the container (current behavior)

#### Scenario: Project mode
- **WHEN** `agents.claude.config` is set to `project`
- **THEN** `~/.asylum/projects/<container>/claude-config/` is mounted into the container

### Requirement: First-run isolation prompt
When the isolation level is not configured for Claude, the system SHALL include an isolation selection step in the onboarding wizard instead of showing a standalone prompt. The step SHALL present the same three options (shared, isolated, project) with **`shared`** as the default and the "(recommended)" annotation. `isolated` and `project` SHALL remain selectable without annotation.

#### Scenario: First run with no config
- **WHEN** asylum starts with Claude agent and no `agents.claude.config` value
- **THEN** an isolation step appears in the onboarding wizard with the three isolation options, defaulting to "shared" and labelling it as recommended

#### Scenario: Config already set
- **WHEN** `agents.claude.config` is already set to a valid value
- **THEN** no isolation step appears in the wizard

#### Scenario: Non-interactive first run
- **WHEN** asylum starts non-interactively with no config value
- **THEN** the default "shared" mode is used without prompting

### Requirement: Config persistence after prompt
After the user selects an isolation level, the choice SHALL be written to `~/.asylum/config.yaml`.

#### Scenario: Choice saved
- **WHEN** the user selects "shared" in the prompt
- **THEN** `~/.asylum/config.yaml` is updated with `agents: { claude: { config: shared } }`

#### Scenario: Subsequent run uses saved choice
- **WHEN** the config was saved from a previous prompt
- **THEN** asylum uses the saved value without prompting

### Requirement: Implicit isolation fallback
When no `agents.<agent>.config` value is resolved (no wizard step ran, no config layer set the value), the runtime SHALL behave as if `shared` were configured for that agent. The fallback applies uniformly to all agents that support isolation, not just Claude.

#### Scenario: No value resolved, fallback applied
- **WHEN** `cfg.AgentIsolation(agentName)` returns `""`
- **THEN** the runtime SHALL select the `shared` codepath for agent config mounting
