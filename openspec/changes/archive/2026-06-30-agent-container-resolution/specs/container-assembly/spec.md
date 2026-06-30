## MODIFIED Requirements

### Requirement: Container naming
Container name SHALL be derived from the project directory and the agent set. For the project's configured agent set the name SHALL be `asylum-<sha256(project_dir)[:12]>-<sanitized_basename>` (the primary container, byte-identical to the prior project-only name) and hostname SHALL be `asylum-<sanitized_basename>`. For a requested agent set that the primary container does not support, the name SHALL be derived from `sha256(project_dir + sorted_agents)` (a secondary container). On first run, old-format project directories (`asylum-<hash>` without suffix) SHALL be migrated to the primary format.

#### Scenario: Naming from project path
- **WHEN** the project directory is `/home/user/code/myapp` and the configured agent set is used
- **THEN** the container name is `asylum-<hash(project)[:12]>-myapp` and hostname is `asylum-myapp`

#### Scenario: Secondary naming for an unsupported agent
- **WHEN** the project's primary container does not support the requested agent set
- **THEN** the container name is derived from `sha256(project_dir + sorted_agents)` and differs from the primary name

#### Scenario: Migration of old project directory
- **WHEN** `~/.asylum/projects/asylum-<hash>` exists but `~/.asylum/projects/asylum-<hash>-<project>` does not
- **THEN** the old directory is renamed and port allocations are updated

### Requirement: Port forwarding
Ports from config SHALL be mapped in docker run args for primary containers. Secondary containers (those derived from `project_dir + sorted_agents` because the primary does not support the requested agent) SHALL receive no port-forwarding arguments.

#### Scenario: Simple port
- **WHEN** port is `3000` for a primary container
- **THEN** `-p 3000:3000` is added to args

#### Scenario: Mapped port
- **WHEN** port is `8080:80` for a primary container
- **THEN** `-p 8080:80` is added to args

#### Scenario: Secondary container omits ports
- **WHEN** the container being assembled is a secondary container
- **THEN** no `-p` arguments are added regardless of configured ports
