## MODIFIED Requirements

### Requirement: Profiles field in config
The Config struct SHALL include a `Kits` field that is a map of kit name to KitConfig. Kit presence in the map means the kit is configured. KitConfig SHALL include a `Disabled` field that when true excludes the kit from resolution. A project-level config MAY set `disabled: false` to override a globally-disabled kit.

#### Scenario: Kit disabled in config
- **WHEN** config has `kits: {shell: {disabled: true}}`
- **THEN** `KitActive("shell")` returns false

#### Scenario: Kit disabled at project level overrides global
- **WHEN** global config has `kits: {github: {}}` and project config has `kits: {github: {disabled: true}}`
- **THEN** the github kit is not active in the merged config

#### Scenario: Kit re-enabled at project level overrides global disabled
- **WHEN** global config has `kits: {ast-grep: {disabled: true}}` and project config has `kits: {ast-grep: {disabled: false}}`
- **THEN** the ast-grep kit is active in the merged config
