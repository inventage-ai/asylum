# cleanup-command Specification

## Purpose

Removes the containers, images, and state asylum leaves behind, scoped to the current project by default and to everything with `--all`. Disk use grows quietly across projects, and this is the supported way to reclaim it rather than hand-picking Docker resources.

## Requirements

### Requirement: Cleanup command
The `cleanup` command SHALL remove asylum resources. By default it scopes cleanup to the current project. With `--all`, it removes all asylum images, volumes, and cached data after user confirmation.

#### Scenario: Scoped cleanup (default)
- **WHEN** `asylum cleanup` is run from a project directory
- **THEN** the project's container, volumes (prefixed with `<container-name>-`), and project data directory are removed
- **AND** the base image and other projects' resources are preserved

#### Scenario: Scoped cleanup outside project dir
- **WHEN** `asylum cleanup` is run and the working directory cannot be resolved
- **THEN** a warning is shown suggesting `asylum cleanup --all`

#### Scenario: Global cleanup with confirmation
- **WHEN** `asylum cleanup --all` is run in a terminal
- **THEN** all asylum images and volumes are enumerated and displayed
- **AND** the user is prompted to confirm before deletion proceeds

#### Scenario: Global cleanup with cache removal
- **WHEN** `asylum cleanup --all` is run and user confirms, then answers y to cache prompt
- **THEN** images, volumes, and host cache/projects dirs are deleted

#### Scenario: Global cleanup without cache removal
- **WHEN** `asylum cleanup --all` is run and user confirms, then answers N to cache prompt
- **THEN** images and volumes are removed, but host cache dir is preserved

#### Scenario: Global cleanup requires terminal
- **WHEN** `asylum cleanup --all` is run outside a terminal
- **THEN** cleanup is aborted with an error

#### Scenario: Agent config preserved
- **WHEN** `asylum cleanup` or `asylum cleanup --all` is run
- **THEN** `~/.asylum/agents/` is NOT removed

#### Scenario: Flag alias
- **WHEN** `asylum --cleanup` is run
- **THEN** behavior is identical to `asylum cleanup`
