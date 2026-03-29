## ADDED Requirements

### Requirement: Cleanup command
The `cleanup` command SHALL remove asylum images, named volumes (shadow and cache), and optionally remove cached data from the host filesystem, while preserving agent config. The `--cleanup` flag SHALL be accepted as an alias.

#### Scenario: Cleanup with cache removal
- **WHEN** `asylum cleanup` is run and user answers y
- **THEN** images and all asylum-prefixed volumes (shadow and cache) are removed, and host cache/projects dirs are deleted

#### Scenario: Cleanup without cache removal
- **WHEN** `asylum cleanup` is run and user answers N
- **THEN** images and all asylum-prefixed volumes are removed, but host cache dir is preserved

#### Scenario: Agent config preserved
- **WHEN** `asylum cleanup` is run
- **THEN** `~/.asylum/agents/` is NOT removed

#### Scenario: Flag alias
- **WHEN** `asylum --cleanup` is run
- **THEN** behavior is identical to `asylum cleanup`
