# kit-defaults Specification

## Purpose

Defines how a kit gets into the resolved set without being named in config. Each kit declares an activation tier — always-on, default-on, or opt-in — and always-on kits are included even when the user names an explicit kit set. `disabled: true` is the escape hatch that overrides both default-on and always-on behavior.

## Requirements

### Requirement: Always-on kits
Kits with tier `TierAlwaysOn` SHALL be included in the resolved set when the user specifies explicit kits but does not mention the always-on kit, unless explicitly disabled.

#### Scenario: Explicit kits plus always-on
- **WHEN** config has `kits: {java: {}}` and the shell kit has tier `TierAlwaysOn`
- **THEN** both java and shell are active

#### Scenario: Always-on kit explicitly disabled
- **WHEN** config has `kits: {java: {}, shell: {disabled: true}}`
- **THEN** java is active but shell is not

#### Scenario: No kits key (nil)
- **WHEN** no config layer specifies kits
- **THEN** all kits are active (including always-on kits) — unchanged behavior

#### Scenario: Empty kits map
- **WHEN** config has `kits: {}`
- **THEN** no kits are active (always-on kits are NOT added to explicit empty)

### Requirement: Kit activation tier
Each kit SHALL declare an activation tier: `TierAlwaysOn` (active even without config), `TierDefault` (active when present in config, added by default), or `TierOptIn` (only active if user explicitly enables). This replaces the `DefaultOn bool` field.

#### Scenario: Always-on tier
- **WHEN** a kit has tier `TierAlwaysOn`
- **THEN** it is active even if not listed in the config's `kits` map

#### Scenario: Default tier
- **WHEN** a kit has tier `TierDefault`
- **THEN** it is active only if its key is present and uncommented in the config's `kits` map

#### Scenario: Opt-in tier
- **WHEN** a kit has tier `TierOptIn`
- **THEN** it is active only if the user explicitly adds it to their config

### Requirement: Kit disabling
Presence of a kit's key in the `kits` map means the kit is configured. A kit SHALL be disableable by setting `disabled: true` in its KitConfig. This overrides default-on behavior and can disable globally-configured kits at project level. A higher-precedence layer MAY set `disabled: false` to re-enable a kit that a lower layer disabled.

#### Scenario: Disable global kit at project level
- **WHEN** global config has `kits: {java: {}, github: {}}` and project config has `kits: {github: {disabled: true}}`
- **THEN** java is active but github is not

#### Scenario: Kit re-enabled at project level overrides global disabled
- **WHEN** global config has `kits: {ast-grep: {disabled: true}}` and project config has `kits: {ast-grep: {disabled: false}}`
- **THEN** the ast-grep kit is active in the merged config

#### Scenario: Disabled kit not resolved
- **WHEN** a kit has `disabled: true` in its KitConfig
- **THEN** it is excluded from the resolved kit list and its DockerSnippet is not included
