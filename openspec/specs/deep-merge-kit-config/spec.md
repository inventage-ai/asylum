# deep-merge-kit-config Specification

## Purpose

Merges the `kits` and `agents` maps per key rather than wholesale, so a project config can adjust one option of a kit without discarding the global entry for it. Within a kit's config, fields merge through struct-tag-driven reflection, which is what lets some fields accumulate across layers while others are last-wins.

## Requirements

### Requirement: Per-key kit map merge
When merging two config layers, the `Kits` map SHALL be merged per-key: overlay keys add to or override base keys, and base keys not present in the overlay SHALL be preserved.

#### Scenario: Project config adds kit without losing global kits
- **WHEN** global config has `kits: {node: {}, openspec: {}}` and project config has `kits: {shell: {build: ["curl ..."]}}}`
- **THEN** the merged result has kits `node`, `openspec`, and `shell` all active

#### Scenario: Project config overrides a global kit's options
- **WHEN** global config has `kits: {java: {default-version: "17"}}` and project config has `kits: {java: {default-version: "21"}}`
- **THEN** the merged result has `java` with `default-version: "21"`

#### Scenario: Overlay with nil KitConfig preserves base KitConfig
- **WHEN** global config has `kits: {node: {packages: ["tsx"]}}` and project config has `kits: {node:}` (nil value)
- **THEN** the merged result has `node` with `packages: ["tsx"]`

#### Scenario: Base kit absent from overlay is preserved
- **WHEN** global config has `kits: {openspec: {}}` and project config has `kits: {shell: {}}`
- **THEN** the merged result contains both `openspec` and `shell`

### Requirement: Per-key agent map merge
When merging two config layers, the `Agents` map SHALL be merged per-key with the same semantics as kits.

#### Scenario: Project adds agent without losing global agents
- **WHEN** global config has `agents: {claude: {}}` and project config has `agents: {gemini: {}}`
- **THEN** the merged result has both `claude` and `gemini` active

### Requirement: KitConfig field-level merge
When two KitConfig values exist for the same kit key, their fields SHALL be merged using struct-tag-driven reflection. Fields tagged `merge:"concat"` SHALL be accumulated (base + overlay). All other fields SHALL use last-wins semantics (overlay replaces base when the overlay value is non-zero). Adding a new field to `KitConfig` SHALL require no changes to the merge function.

#### Scenario: Scalar fields use last-wins
- **WHEN** base has `java: {default-version: "17"}` and overlay has `java: {default-version: "21"}`
- **THEN** the merged KitConfig has `default-version: "21"`

#### Scenario: Disabled flag overrides
- **WHEN** base has `node: {}` and overlay has `node: {disabled: true}`
- **THEN** the merged KitConfig has `disabled: true`

#### Scenario: Packages list concatenates
- **WHEN** base has `node: {packages: ["tsx"]}` and overlay has `node: {packages: ["vitest"]}`
- **THEN** the merged KitConfig has `packages: ["tsx", "vitest"]`

#### Scenario: Build list concatenates
- **WHEN** base has `shell: {build: ["apt-get install foo"]}` and overlay has `shell: {build: ["curl bar"]}`
- **THEN** the merged KitConfig has `build: ["apt-get install foo", "curl bar"]`

#### Scenario: Versions list replaces
- **WHEN** base has `java: {versions: ["17", "21"]}` and overlay has `java: {versions: ["25"]}`
- **THEN** the merged KitConfig has `versions: ["25"]`

#### Scenario: Non-zero overlay Count replaces base
- **WHEN** base has `ports: {count: 5}` and overlay has `ports: {count: 10}`
- **THEN** the merged KitConfig has `count: 10`

#### Scenario: Zero overlay Count preserves base
- **WHEN** base has `ports: {count: 5}` and overlay does not set count
- **THEN** the merged KitConfig has `count: 5`

#### Scenario: Credentials overlay replaces base
- **WHEN** base has `java: {credentials: auto}` and overlay has `java: {credentials: [nexus]}`
- **THEN** the merged KitConfig has `credentials: [nexus]`

#### Scenario: Absent credentials in overlay preserves base
- **WHEN** base has `java: {credentials: auto}` and overlay has `java: {}` (no credentials key)
- **THEN** the merged KitConfig has `credentials: auto`

#### Scenario: New field without merge tag uses last-wins
- **WHEN** a new field is added to `KitConfig` without a `merge` struct tag
- **AND** base has a value and overlay has a non-zero value
- **THEN** the overlay value SHALL be used without any changes to the merge function
