## ADDED Requirements

### Requirement: default-resume config key
The config system SHALL accept an optional top-level boolean key `default-resume`. When set to `true`, asylum's default agent invocation auto-resumes the prior session (the pre-default-new-session behaviour). When unset or `false`, the default is to start a new session. The key SHALL participate in the standard three-layer merge (global → project → local → CLI), last-wins semantics for scalars.

#### Scenario: Unset defaults to false
- **WHEN** no config layer sets `default-resume`
- **THEN** the resolved value is `false`

#### Scenario: Global config opts in
- **WHEN** `~/.asylum/config.yaml` contains `default-resume: true` and no other layer sets it
- **THEN** the resolved value is `true`

#### Scenario: Project layer overrides global
- **WHEN** global config sets `default-resume: true` and `$project/.asylum` sets `default-resume: false`
- **THEN** the resolved value is `false`

#### Scenario: Local layer overrides project
- **WHEN** project config sets `default-resume: false` and `$project/.asylum.local` sets `default-resume: true`
- **THEN** the resolved value is `true`

#### Scenario: Migration dialog writes to global layer
- **WHEN** the resume-migration dialog writes `default-resume: true` on the user's behalf
- **THEN** it is written to `~/.asylum/config.yaml` (the global layer), preserving any other keys already present in that file
