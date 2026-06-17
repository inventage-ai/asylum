## ADDED Requirements

### Requirement: Companion list configuration
The system SHALL accept an optional `companions` field on each entry under `agents` in YAML config, whose value is a list of agent names. The field SHALL merge using last-wins overlay semantics: an overlay that omits `companions` SHALL inherit the base value; an overlay that sets `companions` (including to an empty list) SHALL replace the base value, allowing higher-precedence config layers to clear inherited companions.

#### Scenario: Configured companions
- **WHEN** the resolved config contains `agents: { claude: { companions: [codex] } }`
- **THEN** the system reports codex as a companion of claude

#### Scenario: No companions configured
- **WHEN** the resolved config contains `agents: { claude: { config: shared } }` with no `companions` field
- **THEN** the system reports an empty companion list for claude and behaves exactly as before this change

#### Scenario: Overlay clears inherited list
- **WHEN** the base layer sets `agents.claude.companions: [codex]` and the overlay sets `agents.claude.companions: []`
- **THEN** the resolved config reports no companions for claude

#### Scenario: Overlay replaces inherited list
- **WHEN** the base layer sets `agents.claude.companions: [codex]` and the overlay sets `agents.claude.companions: [gemini]`
- **THEN** the resolved config reports only gemini as a companion of claude

#### Scenario: Self-reference ignored
- **WHEN** the resolved config contains `agents: { claude: { companions: [claude, codex] } }` and claude is the primary agent
- **THEN** the system mounts codex's config but does not double-mount claude's

#### Scenario: Duplicates de-duplicated
- **WHEN** the resolved config contains `agents: { claude: { companions: [codex, codex] } }`
- **THEN** codex's config is mounted exactly once

### Requirement: Companion config dir mount
For each companion of the primary agent, the system SHALL mount the companion's resolved host config dir at the companion's `ContainerConfigDir` inside the container, writable, using the **companion's own** `agents.<companion>.config` isolation setting to resolve the host source path.

#### Scenario: Companion in shared isolation
- **WHEN** primary is claude with `companions: [codex]` and `agents.codex.config: shared`
- **THEN** the container has the host's `~/.codex` mounted at `~/.codex` writable

#### Scenario: Companion in isolated isolation
- **WHEN** primary is claude with `companions: [codex]` and `agents.codex.config: isolated`
- **THEN** the container has `~/.asylum/agents/codex` mounted at `~/.codex` writable

#### Scenario: Companion in project isolation
- **WHEN** primary is claude with `companions: [codex]` and `agents.codex.config: project`
- **THEN** the container has `~/.asylum/projects/<container>/codex-config` mounted at `~/.codex` writable

### Requirement: Companion env var contribution
For each companion of the primary agent, the system SHALL merge the companion's `EnvVars()` into the container environment. If a key returned by a companion is already set by the primary or by an earlier companion, the existing value SHALL win and a warning SHALL be logged.

#### Scenario: Companion env vars set
- **WHEN** primary is claude with `companions: [codex]`
- **THEN** the container has `CODEX_HOME` set to the in-container codex config dir, in addition to `CLAUDE_CONFIG_DIR`

#### Scenario: Env var key collision
- **WHEN** the primary and a companion both define the same env var key
- **THEN** the primary's value is used and the system logs a warning naming the conflicting key and companion

### Requirement: Primary launch unchanged
The presence of companions SHALL NOT alter the primary agent's `Command()`, `HasSession()`, resume behavior, sandbox rules placement, or any single-agent run characteristic. Companions are never launched by asylum.

#### Scenario: Resume considers only the primary
- **WHEN** primary is claude with `companions: [codex]` and a claude session exists but no codex session exists
- **THEN** `claude --continue` is launched

#### Scenario: No-companion run is identical
- **WHEN** the same project is run with `agents.claude.companions` unset
- **THEN** the resulting `docker run` arguments are identical to the pre-change behavior

### Requirement: Missing companion is an error
If a listed companion is not an installed agent for the current project (no registered `AgentInstall` resolving to that name in the assembled image), the system SHALL fail before starting the container with a clear error message naming the missing agent.

#### Scenario: Companion not installed
- **WHEN** primary is claude with `companions: [codex]` and codex is not present in the resolved agent install set
- **THEN** asylum exits non-zero before `docker run` with an error message that names "codex" as the missing companion

#### Scenario: Companion installed
- **WHEN** primary is claude with `companions: [codex]` and codex is in the resolved agent install set
- **THEN** assembly succeeds and the codex mount and env vars are applied

### Requirement: One-directional semantics
A companion declaration on agent A SHALL apply only when A is the primary for the run. Running another agent SHALL NOT consult A's companion list.

#### Scenario: Inverse run does not inherit
- **WHEN** the config has `agents.claude.companions: [codex]` and the user runs `asylum codex`
- **THEN** claude's config is not mounted, claude's env vars are not set, and no companion processing for the claude entry occurs

### Requirement: `~/.agents` mount unaffected
The existing `~/.agents` shared-mode mount SHALL remain gated solely on the primary agent's isolation level being `shared`. Companions SHALL NOT cause or suppress this mount.

#### Scenario: Companion does not enable shared mount
- **WHEN** primary is claude with `config: isolated` and `companions: [codex]` where codex has `config: shared`
- **THEN** `~/.agents` is not mounted into the container

#### Scenario: Companion does not suppress shared mount
- **WHEN** primary is claude with `config: shared` and `companions: [codex]` where codex has `config: isolated`
- **THEN** `~/.agents` is mounted into the container as before
