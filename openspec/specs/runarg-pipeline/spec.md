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

### Requirement: Priority levels
The system SHALL define four priority levels: core=0, kit=1, config=2, cli=3. Higher priority SHALL silently override lower priority when two RunArgs share the same dedup key.

#### Scenario: Config overrides kit
- **WHEN** the ports kit produces `-p 8080:3000` (priority 1) and user config produces `-p 3000:3000` (priority 2) and both target container port 3000
- **THEN** only `-p 3000:3000` from user config SHALL appear in the final args

#### Scenario: Core overridden by kit
- **WHEN** core produces `-e FOO=bar` (priority 0) and a kit produces `-e FOO=baz` (priority 1)
- **THEN** only `-e FOO=baz` from the kit SHALL appear in the final args

### Requirement: Dedup key extraction
The system SHALL extract a dedup key from each RunArg based on its flag type. Args with the same flag type and dedup key are candidates for deduplication or conflict detection.

#### Scenario: Port dedup key
- **WHEN** the flag is `-p` and value is `8080:3000`
- **THEN** the dedup key SHALL be `3000` (container port, right side of colon)

#### Scenario: Port dedup key without mapping
- **WHEN** the flag is `-p` and value is `3000`
- **THEN** the dedup key SHALL be `3000`

#### Scenario: Volume dedup key
- **WHEN** the flag is `-v` and value is `/host/path:/container/path:ro`
- **THEN** the dedup key SHALL be `/container/path`

#### Scenario: Mount dedup key
- **WHEN** the flag is `--mount` and value is `type=volume,src=my-vol,dst=/data`
- **THEN** the dedup key SHALL be `/data`

#### Scenario: Env var dedup key
- **WHEN** the flag is `-e` and value is `COLORTERM=truecolor`
- **THEN** the dedup key SHALL be `COLORTERM`

#### Scenario: Boolean flag dedup key
- **WHEN** the flag is `--privileged` and value is empty
- **THEN** the dedup key SHALL be `--privileged`

#### Scenario: Cap-add dedup key
- **WHEN** the flag is `--cap-add` and value is `SYS_ADMIN`
- **THEN** the dedup key SHALL be `SYS_ADMIN`

#### Scenario: Single-value flag dedup key
- **WHEN** the flag is `--name` and value is `my-container`
- **THEN** the dedup key SHALL be `--name`

### Requirement: Conflict detection
When two RunArgs have the same dedup key and the same priority but different values, the system SHALL abort with an error naming both sources and values.

#### Scenario: Same-priority conflict aborts
- **WHEN** kit A produces `-v /a:/workspace:ro` (priority 1) and kit B produces `-v /b:/workspace:rw` (priority 1)
- **THEN** the system SHALL abort with an error message containing both kit names, the conflicting mount path `/workspace`, and both values

#### Scenario: Same-priority same-value is fine
- **WHEN** kit A produces `--privileged` (priority 1) and kit B also produces `--privileged` (priority 1)
- **THEN** only one `--privileged` SHALL appear in the final args (no conflict)

### Requirement: Deterministic output ordering
The resolved args SHALL be emitted in deterministic order: sorted by priority (ascending), then by source name (alphabetical), then by dedup category key. The docker subcommand `run -d` SHALL be prepended as a fixed prefix before the sorted args during flattening, ensuring they always appear first in the final `[]string`.

#### Scenario: Ordering across priorities
- **WHEN** args come from core (priority 0), docker kit (priority 1), and user config (priority 2)
- **THEN** core args SHALL appear first, then docker kit args, then user config args

#### Scenario: Subcommand always first
- **WHEN** the resolved args are flattened to `[]string`
- **THEN** the first two elements SHALL be `"run"` and `"-d"`, followed by the sorted options, then the image tag and command

### Requirement: Kit ContainerFunc
The Kit struct SHALL have an optional `ContainerFunc` field of type `func(ContainerOpts) ([]RunArg, error)`. When non-nil, it SHALL be called during container assembly for each active kit. `ContainerOpts` SHALL provide ProjectDir, ContainerName, HomeDir, and Config.

#### Scenario: Kit with ContainerFunc
- **WHEN** the ports kit has a non-nil ContainerFunc
- **THEN** it SHALL be called during container assembly and its returned RunArgs SHALL be included in the pipeline with priority 1

#### Scenario: Kit without ContainerFunc
- **WHEN** a kit has a nil ContainerFunc
- **THEN** no RunArgs are contributed by that kit and no error occurs

#### Scenario: ContainerFunc returns error
- **WHEN** a kit's ContainerFunc returns an error
- **THEN** the container assembly SHALL log a warning and continue without that kit's args

### Requirement: Debug output
When the `--debug` flag is passed, the system SHALL print every resolved RunArg to stderr in a table showing the flag+value and source, before launching the container.

#### Scenario: Debug output format
- **WHEN** `--debug` is passed and the resolved args include `-p 10000:10000` from "ports kit" and `--privileged` from "docker kit"
- **THEN** stderr SHALL show a formatted table with each arg and its source

#### Scenario: Debug shows overrides
- **WHEN** `--debug` is passed and a config arg overrode a kit arg
- **THEN** the debug output SHALL include an "Overrides" section showing what was replaced and by what
