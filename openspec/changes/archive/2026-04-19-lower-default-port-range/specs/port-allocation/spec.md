## MODIFIED Requirements

### Requirement: Global port registry
The system SHALL maintain a global port registry at `~/.asylum/ports.json` that maps project directories to allocated port ranges. The file SHALL be locked during read/write operations to prevent concurrent corruption. New allocations SHALL start from base port `7001` to avoid the browser-restricted range at and above `10000`.

#### Scenario: First project allocation
- **WHEN** a project has no existing port allocation and the registry is empty
- **THEN** the system SHALL allocate a range starting at the base port (7001) with the configured count

#### Scenario: Subsequent project allocation
- **WHEN** a new project needs ports and existing allocations (below 10000) end at port 7015
- **THEN** the system SHALL allocate the next range starting at 7016

#### Scenario: Existing project reuse
- **WHEN** a project already has an allocated range in the registry with a start port below 10000
- **THEN** the system SHALL return the existing range without modification

#### Scenario: Stale legacy range reassigned
- **WHEN** a project's existing allocation has a start port at or above 10000
- **THEN** the system SHALL remove that stale entry and allocate a new range from the lowered base (7001 upward), using the current container name

#### Scenario: Legacy entries ignored when picking next start
- **WHEN** computing the next start port for a new allocation and the registry still contains one or more entries with start ≥ 10000
- **THEN** the system SHALL ignore those entries and choose the next start based only on entries below 10000

#### Scenario: Concurrent allocation
- **WHEN** two asylum processes allocate ports simultaneously
- **THEN** the file lock SHALL ensure both get non-overlapping ranges

### Requirement: Port range structure
Each allocation SHALL record the project directory, start port, and count. Host and container ports SHALL be identical (e.g., `-p 7001:7001`).

#### Scenario: Range contents
- **WHEN** a project is allocated 5 ports starting at 7001
- **THEN** the allocation SHALL cover host ports 7001, 7002, 7003, 7004, 7005, each mapped to the same container port

### Requirement: Ports kit definition
A `ports` kit SHALL be registered with tier `TierAlwaysOn`. It SHALL have no `DockerSnippet` or `EntrypointSnippet`. It SHALL have a `ContainerFunc` that calls `ports.Allocate()` and returns `-p` RunArgs for each allocated port. The rules snippet SHALL be dynamically generated via the sandbox rules system (unchanged).

#### Scenario: Kit active by default
- **WHEN** the user has not explicitly configured or disabled the ports kit
- **THEN** it SHALL be active, its ContainerFunc SHALL be called, and ports SHALL be allocated

#### Scenario: Kit disabled
- **WHEN** the user sets `ports: { disabled: true }` in config
- **THEN** the kit SHALL not be in the resolved kit list, its ContainerFunc SHALL not be called, and no automatic port allocation SHALL occur

#### Scenario: ContainerFunc produces port args
- **WHEN** the ports kit's ContainerFunc is called and allocation succeeds for ports 7001-7005
- **THEN** it SHALL return five RunArgs with Flag=`-p`, Values `7001:7001` through `7005:7005`, Source=`ports kit`, Priority=1

#### Scenario: ContainerFunc handles allocation failure
- **WHEN** port allocation fails (e.g., port space exhausted)
- **THEN** the ContainerFunc SHALL return an error, which the pipeline SHALL log as a warning and continue

### Requirement: Port forwarding in container
The allocated ports SHALL be produced by the ports kit's ContainerFunc as RunArgs. They SHALL coexist with user-configured ports from the `ports:` config via the unified deduplication pipeline.

#### Scenario: Allocated ports in docker run
- **WHEN** a project has ports 7001-7005 allocated via the ports kit
- **THEN** the final docker run args SHALL include `-p 7001:7001 -p 7002:7002 -p 7003:7003 -p 7004:7004 -p 7005:7005`

#### Scenario: Coexistence with user ports
- **WHEN** a project has kit-allocated ports AND user-configured `ports: ["3000"]`
- **THEN** both SHALL appear in the docker run command (different container ports, no conflict)

### Requirement: Ports in sandbox rules
The sandbox rules file SHALL include a section listing the project's allocated ports, explaining that the agent can bind to these ports and the user can access them at `http://localhost:<port>`.

#### Scenario: Rules content with ports
- **WHEN** a project has ports 7001-7005 allocated
- **THEN** the sandbox rules file SHALL contain a section explaining these ports are forwarded and accessible from the host
