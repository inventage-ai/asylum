# Spec Delta

## MODIFIED Requirements

### Requirement: Project image generation
The image package SHALL generate a project-specific Dockerfile from the project-layer packages config, kit project snippets, and project kit entrypoint/banner snippets, and build it when any of these are present. Packages declared in the global config SHALL NOT be included in the project image (they are installed in the base image). `EnsureProject` SHALL be called on every asylum invocation regardless of container state. `EnsureProject` SHALL NOT accept kit-specific parameters (e.g., java version); kit-specific project image contributions SHALL be provided by kits via `ProjectSnippetFunc`. The project image SHALL record the hash of the base image it was built on, and SHALL be considered up to date only if both its packages hash and its recorded base hash match the current values.

#### Scenario: No packages configured
- **WHEN** project-layer packages config is empty, no kits have project snippets, and no project kits have entrypoint snippets or banner lines
- **THEN** `asylum:latest` is returned as the image tag

#### Scenario: Project-layer packages configured
- **WHEN** project-layer packages config has apt, npm, pip, or run entries
- **THEN** a project image `asylum:proj-<hash>` is built from a generated Dockerfile

#### Scenario: Only global packages configured
- **WHEN** all configured packages come from the global config and there are no project-layer packages, project snippets, or project entrypoint snippets
- **THEN** `asylum:latest` is returned as the image tag and no project image is built

#### Scenario: Project image up to date
- **WHEN** `asylum:proj-<hash>` already exists with matching packages hash and was built on the current base image
- **THEN** no rebuild occurs

#### Scenario: Base image rebuilt by another project
- **WHEN** `asylum:proj-<hash>` exists with matching packages hash, but the base image was rebuilt since, for example during another project's run
- **THEN** the project image is rebuilt on the current base image

#### Scenario: Kit contributes project snippet
- **WHEN** a kit's `ProjectSnippetFunc` returns a non-empty Dockerfile snippet
- **THEN** the project image SHALL be built and include that snippet

#### Scenario: Project kits with entrypoint snippets only
- **WHEN** packages config is empty but project kits have `EntrypointSnippet`s
- **THEN** a project image SHALL be built containing the project entrypoint script

#### Scenario: Called with running container
- **WHEN** a container is already running
- **THEN** `EnsureProject` SHALL still be called and return the expected tag for comparison
