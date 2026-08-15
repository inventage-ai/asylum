# openspec-kit Specification

## Purpose

Defines the default-on `openspec` kit: it installs `@fission-ai/openspec` via npm, bakes the preferred OpenSpec global config into the image so initialization produces a consistent setup, and contributes a rules snippet telling the in-container agent that OpenSpec is available and how to use it.

## Requirements

### Requirement: OpenSpec CLI kit
The system SHALL provide an `openspec` kit that installs the OpenSpec CLI (`@fission-ai/openspec`) via npm. The kit SHALL declare a dependency on the `node` kit and SHALL be default-on.

#### Scenario: OpenSpec kit active with node
- **WHEN** the openspec kit and node kit are both active
- **THEN** the `openspec` CLI is available in the container

#### Scenario: OpenSpec kit active without node
- **WHEN** the openspec kit is active but the node kit is not
- **THEN** a warning is emitted about the missing node dependency

#### Scenario: OpenSpec kit disabled
- **WHEN** the openspec kit is disabled
- **THEN** the OpenSpec CLI is not installed

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
