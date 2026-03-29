## ADDED Requirements

### Requirement: Global port registry
The system SHALL maintain a global port registry at `~/.asylum/ports.json` that maps project directories to allocated port ranges. The file SHALL be locked during read/write operations to prevent concurrent corruption.

#### Scenario: First project allocation
- **WHEN** a project has no existing port allocation and the registry is empty
- **THEN** the system SHALL allocate a range starting at the base port (10000) with the configured count

#### Scenario: Subsequent project allocation
- **WHEN** a new project needs ports and existing allocations end at port 10014
- **THEN** the system SHALL allocate the next range starting at 10015

#### Scenario: Existing project reuse
- **WHEN** a project already has an allocated range in the registry
- **THEN** the system SHALL return the existing range without modification

#### Scenario: Concurrent allocation
- **WHEN** two asylum processes allocate ports simultaneously
- **THEN** the file lock SHALL ensure both get non-overlapping ranges

### Requirement: Port range structure
Each allocation SHALL record the project directory, start port, and count. Host and container ports SHALL be identical (e.g., `-p 10000:10000`).

#### Scenario: Range contents
- **WHEN** a project is allocated 5 ports starting at 10000
- **THEN** the allocation SHALL cover host ports 10000, 10001, 10002, 10003, 10004, each mapped to the same container port

### Requirement: Ports kit definition
A `ports` kit SHALL be registered with `DefaultOn: true`. It SHALL have no `DockerSnippet` or `EntrypointSnippet`. It SHALL provide a dynamic `RulesSnippet` generated at container start time showing the allocated ports.

#### Scenario: Kit active by default
- **WHEN** the user has not explicitly configured or disabled the ports kit
- **THEN** it SHALL be active and ports SHALL be allocated

#### Scenario: Kit disabled
- **WHEN** the user sets `ports: { disabled: true }` in config
- **THEN** no automatic port allocation SHALL occur (user-configured `ports:` still work independently)

### Requirement: Configurable port count
The number of allocated ports SHALL default to 5 and be configurable via `KitConfig`. The config key SHALL be `count` under the `ports` kit.

#### Scenario: Default count
- **WHEN** no count is specified in config
- **THEN** 5 ports SHALL be allocated

#### Scenario: Custom count
- **WHEN** the user configures `kits: { ports: { count: 10 } }`
- **THEN** 10 ports SHALL be allocated for the project

#### Scenario: Count increase for existing project
- **WHEN** a project's configured count increases from 5 to 8 and the 3 ports after its range are unallocated
- **THEN** the range SHALL be extended to 8 ports

#### Scenario: Count increase blocked by neighbor
- **WHEN** a project's configured count increases but the adjacent ports belong to another project
- **THEN** the system SHALL warn and keep the existing count

### Requirement: Port forwarding in container
The allocated ports SHALL be passed to `docker run` as `-p` flags. They SHALL be appended alongside any user-configured ports from the `ports:` config.

#### Scenario: Allocated ports in docker run
- **WHEN** a project has ports 10000-10004 allocated
- **THEN** `RunArgs` SHALL include `-p 10000:10000 -p 10001:10001 -p 10002:10002 -p 10003:10003 -p 10004:10004`

#### Scenario: Coexistence with user ports
- **WHEN** a project has allocated ports AND user-configured `ports: ["3000"]`
- **THEN** both SHALL appear in the docker run command

### Requirement: Port range release
Port allocations SHALL be released when a project's state is cleaned up via `asylum cleanup`.

#### Scenario: Cleanup releases ports
- **WHEN** `asylum cleanup` removes a project's state
- **THEN** the project's port range SHALL be removed from the global registry

### Requirement: Ports in sandbox rules
The sandbox rules file SHALL include a section listing the project's allocated ports, explaining that the agent can bind to these ports and the user can access them at `http://localhost:<port>`.

#### Scenario: Rules content with ports
- **WHEN** a project has ports 10000-10004 allocated
- **THEN** the sandbox rules file SHALL contain a section explaining these ports are forwarded and accessible from the host
