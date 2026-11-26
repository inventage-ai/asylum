# Spec Delta

## MODIFIED Requirements

### Requirement: Seed honored only in user-maintained configs
`seed` SHALL have scope `personal` (see `config-layer-scopes`): it SHALL be honored in `~/.asylum/config.yaml` and the local layer (`.asylum.local` / `.asylum.local.yaml`). In the project layer (`.asylum` / `.asylum.yaml`) it SHALL be ignored and reported as a rejected option.

#### Scenario: Seed in the project layer
- **WHEN** the global config sets `seed: settings` for Claude and `.asylum` sets `seed: full`
- **THEN** the effective seed mode is `settings` and the rejected-option report names the `.asylum` file
