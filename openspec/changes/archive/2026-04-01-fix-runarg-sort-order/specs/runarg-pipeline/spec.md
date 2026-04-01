## MODIFIED Requirements

### Requirement: RunArg type
The system SHALL represent every docker run **option** as a `RunArg` struct containing: `Flag` (the docker flag, e.g. `-p`, `-v`, `-e`, `--privileged`), `Value` (the flag's value, empty for boolean flags), `Source` (human-readable origin label), and `Priority` (integer, higher wins). The docker subcommand (`run`) and mode flag (`-d`) SHALL NOT be represented as RunArgs.

#### Scenario: Port arg from kit
- **WHEN** the ports kit produces a port mapping
- **THEN** the RunArg SHALL have Flag=`-p`, Value=`10000:10000`, Source=`ports kit`, Priority=1

#### Scenario: Boolean flag from kit
- **WHEN** the docker kit produces a privileged flag
- **THEN** the RunArg SHALL have Flag=`--privileged`, Value=`""`, Source=`docker kit`, Priority=1

#### Scenario: Core structural arg
- **WHEN** the container builder produces the `--rm` flag
- **THEN** the RunArg SHALL have Flag=`--rm`, Value=`""`, Source=`core`, Priority=0

#### Scenario: Subcommand not in pipeline
- **WHEN** the RunArg pipeline is assembled
- **THEN** no RunArg with Flag=`run` or Flag=`-d` SHALL exist in the pipeline

### Requirement: Deterministic output ordering
The resolved args SHALL be emitted in deterministic order: sorted by priority (ascending), then by source name (alphabetical), then by dedup category key. The docker subcommand `run -d` SHALL be prepended as a fixed prefix before the sorted args during flattening, ensuring they always appear first in the final `[]string`.

#### Scenario: Ordering across priorities
- **WHEN** args come from core (priority 0), docker kit (priority 1), and user config (priority 2)
- **THEN** core args SHALL appear first, then docker kit args, then user config args

#### Scenario: Subcommand always first
- **WHEN** the resolved args are flattened to `[]string`
- **THEN** the first two elements SHALL be `"run"` and `"-d"`, followed by the sorted options, then the image tag and command
