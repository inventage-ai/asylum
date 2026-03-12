## ADDED Requirements

### Requirement: Cleanup command
The --cleanup flag SHALL remove asylum images and optionally remove cached data, while preserving agent config.

#### Scenario: Cleanup with cache removal
- **WHEN** `asylum --cleanup` is run and user answers y
- **THEN** images are removed and cache/projects dirs are deleted

#### Scenario: Cleanup without cache removal
- **WHEN** `asylum --cleanup` is run and user answers N
- **THEN** images are removed but cache is preserved

#### Scenario: Agent config preserved
- **WHEN** `asylum --cleanup` is run
- **THEN** `~/.asylum/agents/` is NOT removed
