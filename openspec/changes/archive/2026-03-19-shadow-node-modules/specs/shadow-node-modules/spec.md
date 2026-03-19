## ADDED Requirements

### Requirement: Detect node_modules directories
The system SHALL walk the project directory to find every `package.json` and return the `node_modules` path next to it, whether or not `node_modules` exists yet. It SHALL skip `node_modules` directories themselves and irrelevant heavy directories during the walk.

#### Scenario: Root package.json
- **WHEN** the project root has a `package.json`
- **THEN** `<project>/node_modules` is returned

#### Scenario: Subdirectory package.json
- **WHEN** `package.json` exists only in `frontend/`
- **THEN** `<project>/frontend/node_modules` is returned

#### Scenario: Monorepo with multiple package.json
- **WHEN** the project has `package.json` at the root and under `packages/app/`
- **THEN** both `node_modules` paths are returned

#### Scenario: node_modules does not exist yet
- **WHEN** `package.json` exists but `node_modules` has not been created
- **THEN** the `node_modules` path is still returned (shadow volume created proactively)

#### Scenario: package.json inside node_modules ignored
- **WHEN** `node_modules/some-pkg/package.json` exists
- **THEN** it is not walked and no additional `node_modules` path is returned for it

#### Scenario: No package.json anywhere
- **WHEN** no `package.json` exists anywhere in the project
- **THEN** no paths are returned

#### Scenario: Heavy directories skipped
- **WHEN** `package.json` exists inside `.venv`, `.git`, `vendor`, `target`, or `dist`
- **THEN** those directories are not walked and no `node_modules` path is returned for them

### Requirement: Shadow node_modules with named volumes
Each detected `node_modules` directory SHALL be shadowed with a named Docker volume using `--mount type=volume,src=<name>,dst=<path>`.

#### Scenario: Volume naming
- **WHEN** a `node_modules` at relative path `node_modules` is detected for container `asylum-a1b2c3d4e5f6`
- **THEN** the volume is named `asylum-a1b2c3d4e5f6-npm-<hash>` where `<hash>` is the first 11 hex chars of SHA-256 of the relative path

#### Scenario: Volume persists across restarts
- **WHEN** dependencies are installed inside the container and the container is restarted
- **THEN** the named volume retains the installed dependencies

### Requirement: Feature can be disabled
The shadow feature SHALL be disabled when `features: { shadow-node-modules: false }` is set in config.

#### Scenario: Feature disabled
- **WHEN** config has `features: { shadow-node-modules: false }`
- **THEN** no `--mount` flags for `node_modules` are added

#### Scenario: Feature enabled by default
- **WHEN** config does not mention `shadow-node-modules`
- **THEN** the shadow behavior is active

### Requirement: Cleanup removes shadow volumes
The `--cleanup` command SHALL remove all Docker volumes with the `asylum-` prefix alongside image removal.

#### Scenario: Volumes removed on cleanup
- **WHEN** `asylum --cleanup` is run and asylum-prefixed volumes exist
- **THEN** the volumes are removed

#### Scenario: No volumes to remove
- **WHEN** `asylum --cleanup` is run and no asylum-prefixed volumes exist
- **THEN** cleanup proceeds without error
