# agent-install Specification

## Purpose

Covers the `config` field on an agent's config entry, which selects that agent's config directory isolation level. Leaving it unset is meaningful: it is what triggers the first-run prompt rather than silently assuming a mode.

## Requirements

### Requirement: Agents config field
The AgentConfig struct SHALL include a `Config` field accepting values `shared`, `isolated`, or `project` to control the agent's config directory isolation level.

#### Scenario: AgentConfig with isolation level
- **WHEN** config YAML contains `agents: { claude: { config: shared } }`
- **THEN** the parsed AgentConfig has Config set to `"shared"`

#### Scenario: AgentConfig without isolation level
- **WHEN** config YAML contains `agents: { claude: {} }` or `agents: { claude: }`
- **THEN** the parsed AgentConfig has Config as empty string (triggers prompt on first run)
