## MODIFIED Requirements

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
