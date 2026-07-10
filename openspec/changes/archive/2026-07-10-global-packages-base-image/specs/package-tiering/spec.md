## ADDED Requirements

### Requirement: Package provenance tiering
The system SHALL determine, for each configured package and `shell.build` run-command, whether it installs into the base image or the project image based on the config layer that declared it. Entries declared in the global config (`~/.asylum/config.yaml`) SHALL install into the base image. Entries declared in a project config layer (`.asylum` or `.asylum.local`) SHALL install into the project image. This tiering SHALL apply uniformly to all package types: `apt`, node `npm`, python `pip`, `cx-lang`, and `shell.build` run-commands.

#### Scenario: Global package goes to base image
- **WHEN** the global config declares `kits.node.packages: ["@mermaid-js/mermaid-cli"]` and no project config declares node packages
- **THEN** `@mermaid-js/mermaid-cli` is installed in the base image and is NOT part of the project image

#### Scenario: Project package goes to project image
- **WHEN** a project config (`.asylum` or `.asylum.local`) declares `kits.python.packages: ["ansible"]` and the global config does not
- **THEN** `ansible` is installed in the project image and is NOT part of the base image

#### Scenario: Same kit configured in both layers
- **WHEN** the global config declares `kits.node.packages: ["turbo"]` and a project config declares `kits.node.packages: ["vite"]`
- **THEN** `turbo` is installed in the base image and `vite` is installed in the project image

#### Scenario: All tiered package types
- **WHEN** the global config declares `apt`, node `npm`, python `pip`, `cx-lang` packages, and `shell.build` commands
- **THEN** all of them are installed in the base image and none appear in the project image

#### Scenario: No global packages
- **WHEN** the global config declares no packages or build commands
- **THEN** the base image contains no user-configured package installs and project-declared packages install into the project image as before

### Requirement: Global packages gated on effective base kits
A global-config package or `shell.build` entry SHALL be installed in the base image only if its provider kit (node for `npm`, python for `pip`, cx for `cx-lang`, apt for `apt`, shell for `shell.build`) is present in the effective base kit set. Provider kits excluded by a `--kits` flag or a `disabled: true` toggle SHALL cause their global package entries to be dropped from the base image rather than emitted with no installer present.

#### Scenario: Provider kit excluded by --kits flag
- **WHEN** the global config declares `kits.node.packages: ["@mermaid-js/mermaid-cli"]` and asylum is invoked with `--kits python` (excluding the node kit from the base image)
- **THEN** no npm install for `@mermaid-js/mermaid-cli` is emitted into the base image and the base image builds successfully

#### Scenario: Provider kit disabled in project config
- **WHEN** the global config declares `kits.node.packages: ["turbo"]` and a project config sets `kits.node.disabled: true`
- **THEN** `turbo` is not installed in the base image (its provider kit is excluded from the effective base kit set)

### Requirement: Duplicate package across tiers
When the same package appears in both the global and a project config layer, the system SHALL install it in the base image (from the global declaration) and MAY additionally emit it in the project image. Re-installing an already-present package is a no-op and SHALL NOT cause an error.

#### Scenario: Package in both global and project config
- **WHEN** the global config declares `kits.node.packages: ["turbo"]` and a project config also declares `kits.node.packages: ["turbo"]`
- **THEN** `turbo` is installed in the base image, and the build succeeds regardless of whether the project image re-runs its install
