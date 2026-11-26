# Spec Delta

## ADDED Requirements

### Requirement: Stale node_modules volumes are pruned on start
When asylum starts a fresh container for a project with the shadow feature enabled, it SHALL remove that project's `node_modules` volumes that do not correspond to any currently detected `node_modules` directory. Failure to remove a volume SHALL be reported as a warning and SHALL NOT stop the session.

#### Scenario: package.json removed
- **WHEN** `packages/old/package.json` was deleted since the last start and a fresh container starts
- **THEN** the volume for `packages/old/node_modules` is removed
- **AND** volumes for still-existing `node_modules` directories are kept

#### Scenario: Container already running
- **WHEN** asylum attaches to an already running container
- **THEN** no volumes are pruned

#### Scenario: Shadow feature disabled
- **WHEN** config has `features: { shadow-node-modules: false }`
- **THEN** no volumes are pruned

#### Scenario: Removal fails
- **WHEN** a stale volume cannot be removed because another container uses it
- **THEN** a warning is logged and the session starts normally
