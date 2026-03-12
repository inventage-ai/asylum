## ADDED Requirements

### Requirement: Docker availability check
The docker package SHALL verify the Docker daemon is running before any operation.

#### Scenario: Docker available
- **WHEN** `DockerAvailable()` is called and Docker daemon is running
- **THEN** it returns nil

#### Scenario: Docker not available
- **WHEN** `DockerAvailable()` is called and Docker daemon is not running
- **THEN** it returns an error

### Requirement: Image build
The docker package SHALL build images from a Dockerfile with support for tags, labels, build args, and no-cache mode.

#### Scenario: Build with labels and args
- **WHEN** `Build` is called with a context dir, Dockerfile, tag, labels, and build args
- **THEN** it runs `docker build` with appropriate `--label`, `--build-arg`, `-t`, and `-f` flags

### Requirement: Label inspection
The docker package SHALL retrieve individual label values from Docker images.

#### Scenario: Existing label
- **WHEN** `InspectLabel` is called for an existing image with a known label
- **THEN** it returns the label value

#### Scenario: Missing image
- **WHEN** `InspectLabel` is called for a non-existent image
- **THEN** it returns an error

### Requirement: Image cleanup
The docker package SHALL support removing specific images and pruning images by label filter.

#### Scenario: Remove images
- **WHEN** `RemoveImages` is called with image names
- **THEN** it runs `docker rmi` for those images

#### Scenario: Prune by label
- **WHEN** `PruneImages` is called with a filter label
- **THEN** it runs `docker image prune` with the label filter
