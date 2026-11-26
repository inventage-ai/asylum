# Spec Delta

## ADDED Requirements

### Requirement: Orphan cleanup
`asylum cleanup --orphans` SHALL list all orphaned asylum volumes, ask for confirmation, and remove them. It SHALL require an interactive terminal and SHALL NOT be combinable with `--all`.

#### Scenario: Orphans listed and removed after confirmation
- **WHEN** `asylum cleanup --orphans` is run in a terminal and orphaned volumes exist
- **THEN** the orphaned volumes are listed grouped by project path with the reason they are orphaned
- **AND** after the user answers `y` they are removed

#### Scenario: User declines
- **WHEN** the user answers anything other than `y`
- **THEN** nothing is removed

#### Scenario: Nothing to clean
- **WHEN** no orphaned volumes exist
- **THEN** a message says so and no confirmation is asked

#### Scenario: No terminal
- **WHEN** `asylum cleanup --orphans` is run outside a terminal
- **THEN** cleanup is aborted with an error

#### Scenario: Combined with --all
- **WHEN** `asylum cleanup --all --orphans` is run
- **THEN** an error is shown and nothing is removed

### Requirement: Orphan detection
A volume SHALL be orphaned when its project directory does not exist, or when it is a `node_modules` volume of an existing project whose name does not match any `node_modules` directory currently detected in that project. The project directory SHALL be taken from the `asylum.project` label, or for unlabeled volumes from the port allocation registry by container name.

#### Scenario: Project directory deleted
- **WHEN** a volume is labeled `asylum.project=/home/u/old` and `/home/u/old` does not exist
- **THEN** the volume is orphaned

#### Scenario: Stale node_modules volume
- **WHEN** project `/home/u/app` exists and has a volume `<container>-npm-<hash>` for which no `package.json` exists at the corresponding path
- **THEN** that volume is orphaned and the project's other volumes are not

#### Scenario: Unlabeled volume resolved via port registry
- **WHEN** an unlabeled volume `asylum-abc123def456-old-cache-npm` exists and the port registry maps container `asylum-abc123def456-old` to a directory that does not exist
- **THEN** the volume is orphaned

#### Scenario: Unresolvable volume
- **WHEN** an unlabeled volume's project cannot be determined
- **THEN** it is not removed and is listed separately as having an unknown project

#### Scenario: Volume of running container
- **WHEN** an orphaned volume is used by a running container
- **THEN** it is not removed and is reported as in use

### Requirement: Orphaned projects are removed completely
When a project directory no longer exists, orphan cleanup SHALL also remove that project's stopped containers, its project data directory, and its port allocation. Stopped containers of existing projects that reference orphaned volumes SHALL be removed so the volumes can be deleted.

#### Scenario: Deleted project leftovers
- **WHEN** project `/home/u/old` no longer exists and has a stopped container, a data directory under `~/.asylum/projects/`, and a port allocation
- **AND** the user confirms orphan cleanup
- **THEN** the container, the data directory, the port allocation and all volumes of that project are removed

#### Scenario: Stopped container holds stale node_modules volume
- **WHEN** an existing project's stopped container references an orphaned `node_modules` volume
- **THEN** the stopped container is removed and the volume is deleted
