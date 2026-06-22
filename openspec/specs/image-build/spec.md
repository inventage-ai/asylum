## ADDED Requirements

### Requirement: Base image auto-rebuild
The image package SHALL detect when the embedded Dockerfile or entrypoint.sh has changed and rebuild the base image automatically. `EnsureBase` SHALL be called on every asylum invocation regardless of container state. When a running container exists and `docker inspect` fails, asylum SHALL treat images as up to date rather than erroring out.

#### Scenario: First build
- **WHEN** no `asylum:latest` image exists
- **THEN** the base image is built with hash and version labels

#### Scenario: Hash matches
- **WHEN** the `asylum.hash` label on `asylum:latest` matches the current asset hash
- **THEN** no rebuild occurs

#### Scenario: Hash differs
- **WHEN** the `asylum.hash` label differs from the current asset hash
- **THEN** the base image is rebuilt and dangling images are pruned

#### Scenario: Called with running container
- **WHEN** a container is already running
- **THEN** `EnsureBase` SHALL still be called and return the expected tag for comparison

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

### Requirement: Project Dockerfile format
The generated project Dockerfile SHALL install apt packages as root, and npm/pip/run commands as the claude user.

#### Scenario: All package types
- **WHEN** packages config has apt, npm, pip, and run entries
- **THEN** the generated Dockerfile has apt-get as USER root, and npm/pip/run as USER claude

### Requirement: Pre-build context line
When any actual base or project image build is going to occur during a single asylum invocation, the system SHALL emit exactly one user-facing context line before the per-image `building...` log lines: `Building sandbox image — this takes a few minutes the first time, subsequent runs reuse the cache.` The line SHALL be suppressed entirely when both `EnsureBase` and `EnsureProject` are cache hits in the same invocation.

#### Scenario: First-run build
- **WHEN** neither the base nor the project image exists and asylum is invoked
- **THEN** the context line SHALL be printed exactly once before `building base image...` and SHALL NOT be repeated before `building project image...`

#### Scenario: Base image rebuild only
- **WHEN** the base image hash differs but the project image is up to date (or trivially derived)
- **THEN** the context line SHALL be printed exactly once before `building base image...`

#### Scenario: Project image rebuild only
- **WHEN** the base image is up to date but the project image needs rebuilding
- **THEN** the context line SHALL be printed exactly once before `building project image...`

#### Scenario: Both images cached
- **WHEN** both `EnsureBase` and `EnsureProject` are cache hits
- **THEN** no context line SHALL be printed

#### Scenario: Running container short-circuit
- **WHEN** a container is already running and image-check errors trigger the "using running container" fallthrough
- **THEN** no context line SHALL be printed (no actual build occurs)
