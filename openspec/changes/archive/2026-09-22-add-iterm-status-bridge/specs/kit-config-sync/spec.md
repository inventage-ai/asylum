## MODIFIED Requirements

### Requirement: Insert active kit into existing config
When a new kit is activated — by user consent at the kit-sync prompt, or by first-run defaults — the system SHALL insert it as an **active** entry into the `kits` mapping of `~/.asylum/config.yaml` using `yaml.Node` tree manipulation, preserving existing comments and key ordering.

This SHALL hold regardless of the kit's tier. Kits author their config snippet in the form matching their default tier, so an opt-in kit's authored snippet is commented out; the system SHALL normalise the snippet to active form when the user accepts it. Writing the authored form verbatim would leave an accepted kit disabled, which is indistinguishable to the user from the prompt having no effect.

#### Scenario: New default kit added to config
- **WHEN** kit `rust` (tier `TierDefault`) is activated and `config.yaml` has a `kits` mapping without `rust`
- **THEN** a `rust` key-value pair is appended to the `kits` mapping node using the kit's `ConfigNodes` output

#### Scenario: Accepted opt-in kit is enabled, not commented
- **WHEN** a kit whose tier is `TierOptIn` is offered at the kit-sync prompt and the user accepts it
- **THEN** the entry written to the config is active, and the kit is enabled on the next session

#### Scenario: Existing config preserved
- **WHEN** a new kit is inserted into `config.yaml` that has user comments and custom ordering
- **THEN** all existing comments, key order, and values are preserved in the output

#### Scenario: Kit already in config
- **WHEN** kit `docker` is detected as new (not in `known_kits`) but `config.yaml` already has a `docker` key in `kits`
- **THEN** no modification is made to the config for that kit

### Requirement: Insert commented kit into existing config
When a new kit with tier `TierOptIn` is detected and **not** accepted by the user, the system SHALL insert it as a YAML comment in the `kits` section so the user can see and enable it later.

#### Scenario: Opt-in kit added as comment
- **WHEN** kit `apt` (tier `TierOptIn`) is new and `config.yaml` has a `kits` mapping
- **THEN** a commented-out block for `apt` is added after the existing kit entries

#### Scenario: Declined opt-in kit stays commented
- **WHEN** an opt-in kit is offered at the kit-sync prompt and the user does not accept it
- **THEN** the entry written to the config is commented out and the kit stays inactive
