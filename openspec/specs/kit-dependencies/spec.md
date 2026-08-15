# kit-dependencies Specification

## Purpose

Lets a kit declare that it needs another kit, and validates those declarations during resolution. Kits that install a tool via a runtime someone else provides (the Gemini CLI needing `node`, for example) would otherwise fail deep in an image build rather than up front with a clear message.

## Requirements

### Requirement: Kit dependency declaration
Kits SHALL be able to declare dependencies on other kits via a `Deps` field containing kit names.

#### Scenario: Kit with dependency
- **WHEN** the openspec kit declares `Deps: ["node"]` and the node kit is active
- **THEN** resolution succeeds with no warning

#### Scenario: Kit with missing dependency
- **WHEN** the openspec kit declares `Deps: ["node"]` and the node kit is not active
- **THEN** resolution succeeds but emits a warning that openspec requires the node kit

### Requirement: Dependency validation during resolution
The kit resolution process SHALL check that each resolved kit's dependencies are satisfied by the active kit set and emit warnings for missing dependencies.

#### Scenario: All dependencies satisfied
- **WHEN** all active kits have their dependencies present in the active set
- **THEN** no warnings are emitted

#### Scenario: Multiple missing dependencies
- **WHEN** a kit depends on two kits and neither is active
- **THEN** a warning is emitted for each missing dependency
