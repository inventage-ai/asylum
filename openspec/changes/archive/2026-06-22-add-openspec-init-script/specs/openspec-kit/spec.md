## ADDED Requirements

### Requirement: Preferred OpenSpec global config seeded
The `openspec` kit SHALL seed the preferred OpenSpec global config into the image at build time so that initialization uses the `custom` profile with the workflow set `propose, explore, apply, verify, archive`. The seeded config SHALL be written to the location OpenSpec reads for global configuration (`~/.config/openspec/config.json`).

#### Scenario: Global config present in image
- **WHEN** the openspec kit's Dockerfile snippet has been built
- **THEN** `~/.config/openspec/config.json` SHALL contain the `custom` profile and the workflow list `propose, explore, apply, verify, archive`

#### Scenario: Config drives workflow selection on init
- **WHEN** `openspec init` runs in a container with the seeded global config
- **THEN** it SHALL generate the `verify` workflow and omit the `sync` workflow without interactive selection

### Requirement: OpenSpec kit rules guidance
The `openspec` kit SHALL provide a `RulesSnippet` that informs the in-container agent that OpenSpec is installed and that, when the user wants spec-driven change management in a project where `openspec/` does not yet exist, the agent SHALL run `asylum-openspec-init`.

#### Scenario: Rules describe the init script
- **WHEN** the sandbox rules are assembled for a container with the openspec kit active
- **THEN** the rules SHALL describe running `asylum-openspec-init` to set up OpenSpec in an uninitialized project

### Requirement: OpenSpec kit default-on tier
The `openspec` kit SHALL be default-on, consistent with the documented activation behavior.

#### Scenario: Kit added on first detection
- **WHEN** a project's kit configuration is first generated
- **THEN** the openspec kit SHALL be active by default rather than requiring explicit opt-in
