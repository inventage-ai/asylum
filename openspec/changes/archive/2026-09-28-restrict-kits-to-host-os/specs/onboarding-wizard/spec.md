# Spec Delta

## MODIFIED Requirements

### Requirement: Kit selection step
On first-run invocations, the wizard SHALL include a multi-select step listing top-level kits (registry entries with no `/` in `Name`) excluding kits with `Tier == TierAlwaysOn` and kits that are not available on the host. `TierDefault` kits SHALL be pre-checked; `TierAvailable` kits SHALL be unchecked. Selection SHALL write uncommented `kits:` entries for chosen kits and commented entries for unchosen kits, matching the existing comment-vs-active pattern used by `WriteDefaults`. The step SHALL be skipped when not first-run.

#### Scenario: First run with defaults accepted
- **WHEN** the user presses enter without changing the kit selection
- **THEN** all `TierDefault` top-level kits SHALL be written as active entries in `~/.asylum/config.yaml` and all `TierAvailable` top-level kits SHALL be written as commented entries — identical to today's `WriteDefaults` output

#### Scenario: First run with TierAvailable kit enabled
- **WHEN** the user selects the `java` kit (TierAvailable) in the kit step
- **THEN** `java:` SHALL be written as an active entry under `kits:` in `~/.asylum/config.yaml`

#### Scenario: First run with TierDefault kit deselected
- **WHEN** the user deselects the `node` kit (TierDefault)
- **THEN** `node:` SHALL be written as a commented entry under `kits:` in `~/.asylum/config.yaml`

#### Scenario: Always-on kits excluded
- **WHEN** the kit multi-select is presented
- **THEN** kits with `Tier == TierAlwaysOn` (e.g. `ssh`, `ports`, `shell`, `node`'s always-on parts) SHALL NOT appear in the options

#### Scenario: Sub-kits excluded
- **WHEN** the kit multi-select is presented
- **THEN** kits whose `Name` contains `/` (e.g. `java/maven`, `python/pip`) SHALL NOT appear as separate options

#### Scenario: Subsequent run
- **WHEN** the user runs `asylum` on a non-first-run invocation
- **THEN** the kit step SHALL NOT be included in the wizard

#### Scenario: Kits for another host OS excluded
- **WHEN** the kit multi-select is presented on a Linux host
- **THEN** `iterm` and `dropshare` SHALL NOT appear in the options, and SHALL NOT be written to `~/.asylum/config.yaml` as active or commented entries
