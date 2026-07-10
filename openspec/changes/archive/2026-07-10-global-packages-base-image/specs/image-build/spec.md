## ADDED Requirements

### Requirement: Base image package installation
`EnsureBase` SHALL accept a global-packages map (apt, npm, pip, cx-lang, run) and install those packages in the base image. The install commands SHALL be emitted as a Dockerfile block placed after all kit snippets and before the agent block, so that every providing kit (node/fnm, python/uv, cx) is already installed when the packages install. The global package block SHALL participate in the base image hash so that adding, removing, or changing a global package triggers a base image rebuild. The block SHALL install apt packages as `USER root` and npm/pip/cx-lang/run commands as the container user, using the same generation as the project image.

#### Scenario: Global packages present
- **WHEN** `EnsureBase` is called with a non-empty global-packages map
- **THEN** the base Dockerfile includes the package install commands after the kit snippets and before the agent block

#### Scenario: Global package block affects base hash
- **WHEN** the global-packages map changes between two invocations
- **THEN** the `asylum.hash` label no longer matches and the base image is rebuilt (cascading to project images)

#### Scenario: No global packages
- **WHEN** `EnsureBase` is called with an empty global-packages map
- **THEN** no user-configured package install block is added to the base Dockerfile

#### Scenario: Package USER context in base image
- **WHEN** the global-packages map has apt, npm, pip, and run entries
- **THEN** the base Dockerfile installs apt as `USER root` and npm/pip/run as the container user

## MODIFIED Requirements

### Requirement: Project image generation
The image package SHALL generate a project-specific Dockerfile from the project-layer packages config, kit project snippets, and project kit entrypoint/banner snippets, and build it when any of these are present. Packages declared in the global config SHALL NOT be included in the project image (they are installed in the base image). `EnsureProject` SHALL be called on every asylum invocation regardless of container state. `EnsureProject` SHALL NOT accept kit-specific parameters (e.g., java version); kit-specific project image contributions SHALL be provided by kits via `ProjectSnippetFunc`.

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
- **WHEN** `asylum:proj-<hash>` already exists with matching packages hash
- **THEN** no rebuild occurs

#### Scenario: Kit contributes project snippet
- **WHEN** a kit's `ProjectSnippetFunc` returns a non-empty Dockerfile snippet
- **THEN** the project image SHALL be built and include that snippet

#### Scenario: Project kits with entrypoint snippets only
- **WHEN** project-layer packages config is empty but project kits have `EntrypointSnippet`s
- **THEN** a project image SHALL be built containing the project entrypoint script

#### Scenario: Called with running container
- **WHEN** a container is already running
- **THEN** `EnsureProject` SHALL still be called and return the expected tag for comparison
