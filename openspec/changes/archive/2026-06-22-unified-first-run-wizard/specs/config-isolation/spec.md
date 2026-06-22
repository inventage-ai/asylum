## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Implicit isolation fallback
When no `agents.<agent>.config` value is resolved (no wizard step ran, no config layer set the value), the runtime SHALL behave as if `shared` were configured for that agent. The fallback applies uniformly to all agents that support isolation, not just Claude.

#### Scenario: No value resolved, fallback applied
- **WHEN** `cfg.AgentIsolation(agentName)` returns `""`
- **THEN** the runtime SHALL select the `shared` codepath for agent config mounting
