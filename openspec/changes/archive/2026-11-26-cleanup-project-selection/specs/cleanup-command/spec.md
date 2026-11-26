# Spec Delta

## MODIFIED Requirements

### Requirement: Cleanup command
The `cleanup` command SHALL remove asylum resources. By default it scopes cleanup to the current project. With `--all`, it removes all asylum images, volumes, and cached data after user confirmation.

#### Scenario: Scoped cleanup (default)
- **WHEN** `asylum cleanup` is run from a project directory
- **THEN** the project's container (running or stopped), volumes (prefixed with `<container-name>-`), project data directory and port allocation are removed
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

## ADDED Requirements

### Requirement: Project image removed with its last project
When a project is cleaned up by `asylum cleanup`, `cleanup --select` or as a deleted project by `cleanup --orphans`, the project image recorded for it SHALL be removed if no remaining known project records it and no container uses it. Project images not recorded by any project SHALL NOT be removed by these commands.

#### Scenario: Image used by no other project
- **WHEN** project A is cleaned up and its recorded image is recorded by no other project
- **THEN** the image is removed

#### Scenario: Image shared with another project
- **WHEN** project A is cleaned up and project B records the same image
- **THEN** the image is kept

#### Scenario: Unrecorded image
- **WHEN** a project image exists that no usage record refers to
- **THEN** none of `cleanup`, `cleanup --select` or `cleanup --orphans` removes it

### Requirement: Interactive project selection
`asylum cleanup --select` SHALL print every known project with path, status, last use and size, then offer all projects that are not `running` in a multi-select. It SHALL require an interactive terminal and SHALL NOT be combinable with `--all` or `--orphans`.

#### Scenario: Overview printed
- **WHEN** `asylum cleanup --select` is run in a terminal
- **THEN** all known projects are listed with path, status, last use and size, including running ones

#### Scenario: Running project not selectable
- **WHEN** a project's container is running
- **THEN** it appears in the overview but not in the selection

#### Scenario: No terminal
- **WHEN** `asylum cleanup --select` is run outside a terminal
- **THEN** cleanup is aborted with an error

#### Scenario: Combined with another mode
- **WHEN** `asylum cleanup --select` is combined with `--all` or `--orphans`
- **THEN** an error is shown and nothing is removed

### Requirement: Preselection of stale projects
In the selection, projects with status `missing` and projects whose last use is more than 30 days ago SHALL be preselected. Projects with unknown last use SHALL NOT be preselected.

#### Scenario: Old project preselected
- **WHEN** an `idle` project was last used 45 days ago
- **THEN** it is preselected

#### Scenario: Recent project not preselected
- **WHEN** an `idle` project was last used 3 days ago
- **THEN** it is not preselected

### Requirement: Selected projects cleaned up after confirmation
After the selection, asylum SHALL list the selected projects with their total size and ask for confirmation. On `y`, each selected project SHALL be cleaned up like `asylum cleanup` in that project. A failure for one project SHALL be reported and SHALL NOT stop the others.

#### Scenario: Confirmed
- **WHEN** the user selects two projects and answers `y`
- **THEN** container, volumes, project data, port allocation and unused project image of both projects are removed

#### Scenario: Declined or cancelled
- **WHEN** the user cancels the selection, selects nothing, or answers anything other than `y`
- **THEN** nothing is removed

#### Scenario: Project without known path
- **WHEN** a selected project is known only by its container name
- **THEN** its container, volumes and project data are removed
