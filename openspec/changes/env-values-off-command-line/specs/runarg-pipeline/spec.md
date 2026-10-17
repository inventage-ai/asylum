# Spec Delta

## MODIFIED Requirements

### Requirement: Debug output
When the `--debug` flag is passed, the system SHALL print every resolved RunArg to stderr in a table showing the flag+value and source, before launching the container. Values of `-e` args SHALL be masked as `NAME=***`, in the table and in the overrides section.

#### Scenario: Debug output format
- **WHEN** `--debug` is passed and the resolved args include `-p 10000:10000` from "ports kit" and `--privileged` from "docker kit"
- **THEN** stderr SHALL show a formatted table with each arg and its source

#### Scenario: Debug shows overrides
- **WHEN** `--debug` is passed and a config arg overrode a kit arg
- **THEN** the debug output SHALL include an "Overrides" section showing what was replaced and by what

#### Scenario: Env values masked
- **WHEN** `--debug` is passed and the resolved args include `-e GH_TOKEN=secret`
- **THEN** stderr SHALL show `-e GH_TOKEN=***` and SHALL NOT contain `secret`

### Requirement: Conflict detection
When two RunArgs have the same dedup key and the same priority but different values, the system SHALL abort with an error naming both sources and values. Values of `-e` args SHALL be masked as `NAME=***` in that error.

#### Scenario: Same-priority conflict aborts
- **WHEN** kit A produces `-v /a:/workspace:ro` (priority 1) and kit B produces `-v /b:/workspace:rw` (priority 1)
- **THEN** the system SHALL abort with an error message containing both kit names, the conflicting mount path `/workspace`, and both values

#### Scenario: Same-priority same-value is fine
- **WHEN** kit A produces `--privileged` (priority 1) and kit B also produces `--privileged` (priority 1)
- **THEN** only one `--privileged` SHALL appear in the final args (no conflict)

#### Scenario: Env conflict does not print values
- **WHEN** two args of the same priority set `-e TOKEN=a` and `-e TOKEN=b`
- **THEN** the error names both sources and shows `TOKEN=***` for both, without `a` or `b`
