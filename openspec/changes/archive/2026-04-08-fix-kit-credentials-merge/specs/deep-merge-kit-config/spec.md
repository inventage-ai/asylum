## MODIFIED Requirements

### Requirement: KitConfig field-level merge
When two KitConfig values exist for the same kit key, their fields SHALL be merged with field-appropriate semantics.

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
