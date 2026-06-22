## ADDED Requirements

### Requirement: Active agent identity env var
The container assembly SHALL set an `ASYLUM_AGENT` environment variable carrying the name of the active agent for the run. This SHALL be set for every agent, independent of any agent-specific or kit-contributed environment variables.

#### Scenario: Agent name exposed to container
- **WHEN** a container is assembled for an agent
- **THEN** the container run args SHALL include `-e ASYLUM_AGENT=<name>` where `<name>` is the active agent's name

#### Scenario: Available to in-container scripts
- **WHEN** an in-container script reads `ASYLUM_AGENT`
- **THEN** it SHALL receive the name of the agent driving the current session
