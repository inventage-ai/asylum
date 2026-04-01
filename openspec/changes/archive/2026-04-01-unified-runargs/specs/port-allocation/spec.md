## MODIFIED Requirements

### Requirement: Ports kit definition
A `ports` kit SHALL be registered with tier `TierAlwaysOn`. It SHALL have no `DockerSnippet` or `EntrypointSnippet`. It SHALL have a `ContainerFunc` that calls `ports.Allocate()` and returns `-p` RunArgs for each allocated port. The rules snippet SHALL be dynamically generated via the sandbox rules system (unchanged).

#### Scenario: Kit active by default
- **WHEN** the user has not explicitly configured or disabled the ports kit
- **THEN** it SHALL be active, its ContainerFunc SHALL be called, and ports SHALL be allocated

#### Scenario: Kit disabled
- **WHEN** the user sets `ports: { disabled: true }` in config
- **THEN** the kit SHALL not be in the resolved kit list, its ContainerFunc SHALL not be called, and no automatic port allocation SHALL occur

#### Scenario: ContainerFunc produces port args
- **WHEN** the ports kit's ContainerFunc is called and allocation succeeds for ports 10000-10004
- **THEN** it SHALL return five RunArgs with Flag=`-p`, Values `10000:10000` through `10004:10004`, Source=`ports kit`, Priority=1

#### Scenario: ContainerFunc handles allocation failure
- **WHEN** port allocation fails (e.g., port space exhausted)
- **THEN** the ContainerFunc SHALL return an error, which the pipeline SHALL log as a warning and continue

### Requirement: Port forwarding in container
The allocated ports SHALL be produced by the ports kit's ContainerFunc as RunArgs. They SHALL coexist with user-configured ports from the `ports:` config via the unified deduplication pipeline.

#### Scenario: Allocated ports in docker run
- **WHEN** a project has ports 10000-10004 allocated via the ports kit
- **THEN** the final docker run args SHALL include `-p 10000:10000 -p 10001:10001 -p 10002:10002 -p 10003:10003 -p 10004:10004`

#### Scenario: Coexistence with user ports
- **WHEN** a project has kit-allocated ports AND user-configured `ports: ["3000"]`
- **THEN** both SHALL appear in the docker run command (different container ports, no conflict)

## REMOVED Requirements

### Requirement: Port range release
**Reason**: Port allocations are permanent per project directory. Releasing ports on cleanup added complexity for no practical benefit — port space (10000-65535) is large enough that reuse is unnecessary.
**Migration**: Remove `ports.Release()`, `ports.ReleaseContainer()`, and their callsites in cleanup and prune commands. `ports.RenameContainer()` is retained for project directory migration.
