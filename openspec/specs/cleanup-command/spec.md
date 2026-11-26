# cleanup-command Specification

## Purpose

Removes the containers, images, and state asylum leaves behind, scoped to the current project by default and to everything with `--all`. Disk use grows quietly across projects, and this is the supported way to reclaim it rather than hand-picking Docker resources.

## Requirements

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

### Requirement: Cleanup removes read-only subtrees
When removing the project data directory, the projects directory, or the cache directory, cleanup SHALL also remove subtrees whose directories lack write permission, such as a Go module cache left in Claude's temp root. If a removal fails with a permission error, cleanup SHALL grant the owner write permission on the directories in the tree and retry once.

#### Scenario: Read-only tree in project data
- **WHEN** `~/.asylum/projects/<container-name>/tmp` contains a directory with mode 0555 holding files with mode 0444
- **AND** `asylum cleanup` is run for the project
- **THEN** the project data directory is removed completely and cleanup reports success

#### Scenario: Retry still fails
- **WHEN** removal still fails after the permissions are granted
- **THEN** cleanup reports the error as it does today

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
