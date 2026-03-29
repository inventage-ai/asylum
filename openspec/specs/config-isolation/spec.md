## ADDED Requirements

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
When the isolation level is not configured for Claude, the system SHALL include an isolation selection step in the onboarding wizard instead of showing a standalone prompt. The step SHALL present the same three options (shared, isolated, project) with "isolated" as the default.

#### Scenario: First run with no config
- **WHEN** asylum starts with Claude agent and no `agents.claude.config` value
- **THEN** an isolation step appears in the onboarding wizard with the three isolation options, defaulting to "isolated"

#### Scenario: Config already set
- **WHEN** `agents.claude.config` is already set to a valid value
- **THEN** no isolation step appears in the wizard

#### Scenario: Non-interactive first run
- **WHEN** asylum starts non-interactively with no config value
- **THEN** the default "isolated" mode is used without prompting

### Requirement: Config persistence after prompt
After the user selects an isolation level, the choice SHALL be written to `~/.asylum/config.yaml`.

#### Scenario: Choice saved
- **WHEN** the user selects "shared" in the prompt
- **THEN** `~/.asylum/config.yaml` is updated with `agents: { claude: { config: shared } }`

#### Scenario: Subsequent run uses saved choice
- **WHEN** the config was saved from a previous prompt
- **THEN** asylum uses the saved value without prompting
