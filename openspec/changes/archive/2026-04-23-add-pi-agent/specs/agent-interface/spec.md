## MODIFIED Requirements

### Requirement: Agent interface
The agent package SHALL define an Agent interface with methods: Name, Binary, NativeConfigDir, ContainerConfigDir, AsylumConfigDir, EnvVars, HasSession, Command — matching PLAN.md section 8. The `Command` method SHALL accept an options parameter that carries context from the container layer (e.g. the shared kit-skills directory path), so agent implementations can tailor their launch command accordingly.

#### Scenario: Interface methods
- **WHEN** an Agent implementation is used
- **THEN** all interface methods return correct values per the agent profile table in PLAN.md section 3.1

#### Scenario: Command receives kit-skills context
- **WHEN** `Command` is called with an options value whose `KitSkillsDir` field is `/opt/asylum-skills`
- **THEN** agent implementations that care about skills (Claude) act on it, and agents that do not (Gemini, Codex, Opencode, Echo, Pi) ignore it

### Requirement: Agent registry
The agent package SHALL provide a Get function that returns an Agent by name ("claude", "gemini", "codex", "opencode", "pi") or an error for unknown names.

#### Scenario: Valid agent lookup
- **WHEN** `Get("gemini")` is called
- **THEN** it returns the Gemini agent implementation with no error

#### Scenario: Valid pi agent lookup
- **WHEN** `Get("pi")` is called
- **THEN** it returns the Pi agent implementation with no error

#### Scenario: Invalid agent lookup
- **WHEN** `Get("unknown")` is called
- **THEN** it returns an error
