## ADDED Requirements

### Requirement: Project entrypoint assembly
The image package SHALL provide an `assembleProjectEntrypoint` function that builds a project-level entrypoint script from project kit `EntrypointSnippet`s and `BannerLines`. The script SHALL begin with `#!/bin/bash` and `set -e`.

#### Scenario: Project kits with entrypoint snippets
- **WHEN** project kits have non-empty `EntrypointSnippet`s
- **THEN** the assembled script SHALL contain all snippets concatenated in kit order

#### Scenario: No project kits have entrypoint snippets or banner lines
- **WHEN** no project kits have `EntrypointSnippet`s or `BannerLines`
- **THEN** `assembleProjectEntrypoint` SHALL return nil (no script generated)

### Requirement: Project banner lines exported as variable
When project kits have `BannerLines`, the project entrypoint script SHALL export a `PROJECT_BANNER` variable containing the assembled banner line commands. The base entrypoint SHALL evaluate this variable in the welcome banner block.

#### Scenario: Project kit with banner lines
- **WHEN** a project kit has `BannerLines`
- **THEN** the project entrypoint SHALL contain an `export PROJECT_BANNER=...` line with the assembled banner commands

#### Scenario: No project banner lines
- **WHEN** no project kits have `BannerLines`
- **THEN** the project entrypoint SHALL NOT contain a `PROJECT_BANNER` export

### Requirement: Project entrypoint embedded in project image
When a project entrypoint script is generated, it SHALL be written to the build context and COPY'd into the project image at `/usr/local/bin/project-entrypoint.sh` with execute permissions.

#### Scenario: Project image with entrypoint
- **WHEN** project kits produce a non-nil project entrypoint
- **THEN** the project Dockerfile SHALL include a `COPY` instruction for `project-entrypoint.sh`

#### Scenario: Project image without entrypoint
- **WHEN** no project entrypoint is generated
- **THEN** the project Dockerfile SHALL NOT reference `project-entrypoint.sh`
