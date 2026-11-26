# Spec Delta

## Purpose

Lets asylum tell which projects it has resources for, when each was last used and how much disk space it holds, so that users can decide what to clean up instead of guessing from Docker resource names.

## ADDED Requirements

### Requirement: Project usage record
Every agent, shell or run session start SHALL record in the project data directory the project path, the time of the start and the image tag the container runs on. A failure to write the record SHALL be logged as a warning and SHALL NOT abort the start.

#### Scenario: Fresh start recorded
- **WHEN** a session starts a new container for project `/home/u/app`
- **THEN** the project's usage record holds `/home/u/app`, the start time and the container's image tag

#### Scenario: Attach updates last use
- **WHEN** a second session attaches to the already running container of a project
- **THEN** the recorded start time is updated to the time of that session start

### Requirement: Project enumeration
asylum SHALL treat a project as known when any of these exists for it: a usage record, a port allocation, a labeled volume, a project data directory, or a container. Resources that belong to the same container name SHALL be merged into one project. A project whose path cannot be determined SHALL be shown by its container name.

#### Scenario: Project known only from the port registry
- **WHEN** a project has a port allocation but no usage record, because it was last started with an older asylum version
- **THEN** it is listed with the path from the port allocation

#### Scenario: Unknown path
- **WHEN** a project data directory exists that no usage record, port allocation or labeled volume maps to a path
- **THEN** the project is listed by its container name

### Requirement: Project status
Each known project SHALL have exactly one status: `running` if its container runs, else `missing` if its path is known and does not exist, else `stopped` if a stopped container exists, else `idle`.

#### Scenario: Deleted project with stopped container
- **WHEN** a project's directory no longer exists and its container is stopped
- **THEN** its status is `missing`

#### Scenario: No container
- **WHEN** a project's directory exists and it has no container
- **THEN** its status is `idle`

### Requirement: Last use
The last use of a project SHALL be the current time if its container runs, else the start time from its usage record, else the modification time of its project data directory. Without any of these the last use SHALL be shown as unknown.

#### Scenario: Older project without usage record
- **WHEN** a project has no usage record but has a project data directory
- **THEN** its last use is the modification time of that directory

### Requirement: Project size
The size of a project SHALL be the sum of its volumes, its project data directory and its project image when no other known project records the same image. If Docker does not report sizes, the size SHALL be shown as unknown and listing SHALL continue.

#### Scenario: Shared image not counted
- **WHEN** two projects record the same project image
- **THEN** the image size is counted for neither of them
