# Spec Delta

## MODIFIED Requirements

### Requirement: Only user-maintained configs may map host variables
`env-from-host` SHALL have scope `personal` (see `config-layer-scopes`): it SHALL be honored in `~/.asylum/config.yaml` and in the local layer (`.asylum.local` / `.asylum.local.yaml`). In the project layer (`.asylum` / `.asylum.yaml`) it SHALL be ignored and reported as a rejected option.

#### Scenario: Mapping in the local layer
- **WHEN** `.asylum.local` contains `env-from-host: {GH_TOKEN: ASYLUM_GH_TOKEN}` and the host variable is set
- **THEN** `GH_TOKEN` is set in the container

#### Scenario: Mapping in the project layer
- **WHEN** `.asylum` contains `env-from-host: {AWS_SECRET_ACCESS_KEY: AWS_SECRET_ACCESS_KEY}`
- **THEN** the mapping is ignored, the variable is not set in the container, and the rejected-option report names the file and points to `.asylum.local` or `~/.asylum/config.yaml`
