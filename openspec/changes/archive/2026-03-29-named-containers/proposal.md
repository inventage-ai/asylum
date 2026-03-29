## Why

Container names are currently opaque hashes (`asylum-7a3f2b1c9e04`). When listing Docker containers or volumes, there's no way to tell which project they belong to without reverse-engineering the hash. Adding the project name makes `docker ps`, `docker volume ls`, and `~/.asylum/projects/` human-readable at a glance.

## What Changes

- **Container name format**: `asylum-<hash[:12]>` → `asylum-<hash[:12]>-<project>` where `<project>` is the sanitized directory basename (same logic as `safeHostname` but without the `asylum-` prefix)
- **Migration**: When asylum starts in a project that has an old-format `~/.asylum/projects/<old-name>/` directory, rename it to the new format. Also update `ports.json` entries.
- **Volumes**: Existing Docker volumes can't be renamed. Old volumes will be orphaned and cleaned up on next `asylum cleanup`. New volumes use the new container name prefix.

## Capabilities

### Modified Capabilities
- `container-assembly`: Container name now includes project basename suffix

## Impact

- **internal/container/container.go**: `ContainerName` returns new format; add `sanitizeProject` helper (extracted from `safeHostname`)
- **internal/container/container.go**: New `MigrateProjectDir` function to rename old-format dirs and update ports
- **cmd/asylum/main.go**: Call migration before container start
- **internal/ports/ports.go**: Update container name in existing allocations
- **Tests**: Update expected container names
