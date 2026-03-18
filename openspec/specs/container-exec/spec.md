## ADDED Requirements

### Requirement: Detect running container
The docker package SHALL provide a function to check if a container with a given name is currently running.

#### Scenario: Container is running
- **WHEN** a container named `asylum-<hash>` is running
- **THEN** `IsRunning("asylum-<hash>")` returns `true`

#### Scenario: Container is not running
- **WHEN** no container with that name exists
- **THEN** `IsRunning("asylum-<hash>")` returns `false`

#### Scenario: Container exists but is stopped
- **WHEN** a container with that name exists but is in exited/dead state
- **THEN** `IsRunning("asylum-<hash>")` returns `false`

### Requirement: Exec into running container for shell mode
When a container is already running for the current project and the user runs `asylum shell`, asylum SHALL exec into the running container instead of starting a new one.

#### Scenario: Shell with running container
- **WHEN** the user runs `asylum shell` and a container is running for the project
- **THEN** asylum runs `docker exec -it <container-name> /bin/zsh`

#### Scenario: Admin shell with running container
- **WHEN** the user runs `asylum shell --admin` and a container is running for the project
- **THEN** asylum runs `docker exec -it -u root <container-name> /bin/zsh`

### Requirement: Exec into running container for run mode
When a container is already running and the user runs `asylum run <cmd>`, asylum SHALL exec the command in the running container.

#### Scenario: Run command with running container
- **WHEN** the user runs `asylum run echo hello` and a container is running
- **THEN** asylum runs `docker exec -it <container-name> echo hello`

### Requirement: Agent mode does not exec
When the user runs `asylum` (agent mode) while a container is running, asylum SHALL NOT exec into it.

#### Scenario: Agent mode with running container
- **WHEN** the user runs `asylum` and a container is already running
- **THEN** asylum attempts `docker run` as usual (which fails with a name conflict)

### Requirement: Skip image build when exec-ing
When asylum detects it will exec into a running container, it SHALL skip the image build step.

#### Scenario: No image build on exec
- **WHEN** a container is running and the mode is shell or run
- **THEN** `EnsureBase` and `EnsureProject` are not called
