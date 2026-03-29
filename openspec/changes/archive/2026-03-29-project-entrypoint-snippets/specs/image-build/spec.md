## MODIFIED Requirements

### Requirement: Project image generation
The image package SHALL generate a project-specific Dockerfile from the packages config and build it when packages are configured OR when project kits have `EntrypointSnippet`s or `BannerLines`.

#### Scenario: No packages configured
- **WHEN** packages config is empty and no project kits have entrypoint snippets or banner lines
- **THEN** `asylum:latest` is returned as the image tag

#### Scenario: Packages configured
- **WHEN** packages config has apt, npm, pip, or run entries
- **THEN** a project image `asylum:proj-<hash>` is built from a generated Dockerfile

#### Scenario: Project image up to date
- **WHEN** `asylum:proj-<hash>` already exists with matching packages hash
- **THEN** no rebuild occurs

#### Scenario: Project kits with entrypoint snippets only
- **WHEN** packages config is empty but project kits have `EntrypointSnippet`s
- **THEN** a project image SHALL be built containing the project entrypoint script
