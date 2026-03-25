## ADDED Requirements

### Requirement: Agent install registry
The agent package SHALL provide a registry of agent install definitions, each containing a DockerSnippet, optional ProfileDeps, and a BannerLine.

#### Scenario: Registry contains all built-in agents
- **WHEN** the registry is queried for all agent names
- **THEN** it returns claude, codex, gemini, and opencode

#### Scenario: Unknown agent name
- **WHEN** an agent install is looked up by a name not in the registry
- **THEN** an error is returned

### Requirement: Agent install resolution
The system SHALL resolve a list of agent names into agent install definitions. When nil (unspecified), it SHALL default to `["claude"]`. An explicit empty list means no agents.

#### Scenario: Nil defaults to claude only
- **WHEN** the agents config field is nil (not specified)
- **THEN** only the Claude install definition is returned

#### Scenario: Empty means no agents
- **WHEN** the agents config field is an explicit empty list
- **THEN** no agent installs are returned

#### Scenario: All agents selected explicitly
- **WHEN** the agents config field is `["claude", "codex", "gemini", "opencode"]`
- **THEN** all four agent install definitions are returned

#### Scenario: Specific agent selection
- **WHEN** the agents config field is `["gemini"]`
- **THEN** only the Gemini install definition is returned

### Requirement: Profile dependency validation
When resolving agent installs, the system SHALL check that each agent's ProfileDeps are satisfied by the active profile set. If a dependency is missing, a warning SHALL be emitted.

#### Scenario: Gemini with node profile active
- **WHEN** gemini is selected and the node profile is active
- **THEN** resolution succeeds with no warning

#### Scenario: Gemini without node profile
- **WHEN** gemini is selected but the node profile is not active
- **THEN** resolution succeeds but emits a warning that gemini requires the node profile

### Requirement: Agent Dockerfile snippets
Each agent install SHALL provide a DockerSnippet containing the RUN instructions to install that agent's CLI.

#### Scenario: Claude install snippet
- **WHEN** the claude agent install is active
- **THEN** its DockerSnippet installs Claude Code via the official install script

#### Scenario: Gemini install snippet
- **WHEN** the gemini agent install is active
- **THEN** its DockerSnippet installs @google/gemini-cli via npm

#### Scenario: Codex install snippet
- **WHEN** the codex agent install is active
- **THEN** its DockerSnippet installs @openai/codex via npm

#### Scenario: Opencode install snippet
- **WHEN** the opencode agent install is active
- **THEN** its DockerSnippet installs opencode via go install

### Requirement: Agent banner lines
Each agent install SHALL provide a BannerLine that displays the agent's version in the welcome banner.

#### Scenario: Only claude installed
- **WHEN** only claude is in the active agents
- **THEN** the welcome banner shows Claude version but not Gemini or Codex versions

#### Scenario: All agents installed
- **WHEN** all agents are explicitly installed
- **THEN** the welcome banner shows Claude, Codex, Gemini, and Opencode versions

### Requirement: Agents config field
The Config struct SHALL include an `Agents` field that accepts a list of agent names, with nil-means-all and empty-means-none semantics, last-wins merge across config layers.

#### Scenario: Agents in YAML config
- **WHEN** a config file contains `agents: [claude, gemini]`
- **THEN** the parsed Config has Agents set to `["claude", "gemini"]`

#### Scenario: CLI flag overrides config
- **WHEN** config has `agents: [claude]` and CLI passes `--agents gemini`
- **THEN** the effective agent list is `["gemini"]`
