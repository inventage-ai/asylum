## MODIFIED Requirements

### Requirement: Project image generation
The image package SHALL generate a project-specific Dockerfile from the packages config, kit project snippets, and project kit entrypoint/banner snippets, and build it when any of these are present. `EnsureProject` SHALL be called on every asylum invocation regardless of container state. `EnsureProject` SHALL NOT accept kit-specific parameters (e.g., java version); kit-specific project image contributions SHALL be provided by kits via `ProjectSnippetFunc`.

#### Scenario: No packages configured
- **WHEN** packages config is empty, no kits have project snippets, and no project kits have entrypoint snippets or banner lines
- **THEN** `asylum:latest` is returned as the image tag

#### Scenario: Packages configured
- **WHEN** packages config has apt, npm, pip, or run entries
- **THEN** a project image `asylum:proj-<hash>` is built from a generated Dockerfile

#### Scenario: Project image up to date
- **WHEN** `asylum:proj-<hash>` already exists with matching packages hash
- **THEN** no rebuild occurs

#### Scenario: Kit contributes project snippet
- **WHEN** a kit's `ProjectSnippetFunc` returns a non-empty Dockerfile snippet
- **THEN** the project image SHALL be built and include that snippet

#### Scenario: Project kits with entrypoint snippets only
- **WHEN** packages config is empty but project kits have `EntrypointSnippet`s
- **THEN** a project image SHALL be built containing the project entrypoint script

#### Scenario: Called with running container
- **WHEN** a container is already running
- **THEN** `EnsureProject` SHALL still be called and return the expected tag for comparison
