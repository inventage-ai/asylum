## ADDED Requirements

### Requirement: FeatureOff method for default-on features
The config system SHALL provide a `FeatureOff(name)` method that returns true only when a feature is explicitly set to `false` in the features map. This complements `Feature()` which checks for explicitly `true`.

#### Scenario: Feature explicitly disabled
- **WHEN** config has `features: { shadow-node-modules: false }`
- **THEN** `FeatureOff("shadow-node-modules")` returns `true`

#### Scenario: Feature not mentioned
- **WHEN** config has no `shadow-node-modules` entry in features
- **THEN** `FeatureOff("shadow-node-modules")` returns `false`

#### Scenario: Feature explicitly enabled
- **WHEN** config has `features: { shadow-node-modules: true }`
- **THEN** `FeatureOff("shadow-node-modules")` returns `false`
