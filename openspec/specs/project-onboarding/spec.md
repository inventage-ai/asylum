## Requirements

### Requirement: Task detection
The onboarding system SHALL scan the project directory using tasks registered by active profiles and collect pending workloads. A workload is pending if it has no state record or its input hash has changed.

#### Scenario: First run with detectable tasks
- **WHEN** the project has lockfiles and no onboarding state exists
- **THEN** all matching workloads from active profiles' onboarding tasks are collected as pending

#### Scenario: Subsequent run with unchanged inputs
- **WHEN** onboarding state exists and lockfile hashes match
- **THEN** no workloads are collected (all skipped)

#### Scenario: Lockfile changed since last run
- **WHEN** a lockfile hash differs from the stored state
- **THEN** that workload is collected as pending

#### Scenario: No detectable tasks
- **WHEN** no active profile's onboarding tasks detect any workloads
- **THEN** onboarding completes immediately with no prompt

#### Scenario: Profile not active
- **WHEN** the node/npm profile is not active
- **THEN** npm onboarding tasks are not registered and lockfiles are not scanned

### Requirement: Consolidated user prompt
When pending workloads exist, the system SHALL display all of them in a single prompt and ask for confirmation before executing.

#### Scenario: User accepts
- **WHEN** the user presses Enter or types anything other than "n"
- **THEN** all pending workloads are executed

#### Scenario: User declines
- **WHEN** the user types "n" or "N"
- **THEN** no workloads are executed and the session starts normally

### Requirement: Workload execution via docker exec
Post-container workloads SHALL be executed inside the running container via `docker exec` from the Go binary, with stdout/stderr streamed to the user's terminal.

#### Scenario: Successful execution
- **WHEN** a workload command succeeds
- **THEN** its hash is saved to onboarding state

#### Scenario: Failed execution
- **WHEN** a workload command fails
- **THEN** an error is shown to the user, state is not updated for that workload, and the session starts normally (non-fatal)

### Requirement: State persistence
Onboarding state SHALL be stored at `~/.asylum/projects/<container-name>/onboarding.json` with input hashes keyed by task name and workload label.

#### Scenario: State saved after successful workload
- **WHEN** a workload completes successfully
- **THEN** its hash is written to onboarding.json

#### Scenario: State cleared on cleanup
- **WHEN** `asylum --cleanup` is run
- **THEN** onboarding state is removed along with other project data

### Requirement: Skip mechanisms
Onboarding SHALL be skippable at three levels: CLI flag (`--skip-onboarding`), global config (`features: { onboarding: false }`), and per-task config (`onboarding: { <task-name>: false }`).

#### Scenario: CLI flag skips all
- **WHEN** `--skip-onboarding` is passed
- **THEN** no task detection or prompting occurs

#### Scenario: Global config disables all
- **WHEN** config has `features: { onboarding: false }`
- **THEN** no task detection or prompting occurs

#### Scenario: Per-task config disables one task
- **WHEN** config has `onboarding: { npm: false }`
- **THEN** that task is skipped but other tasks still run

### Requirement: Agent mode only
Onboarding SHALL only run in agent mode, not in shell, admin shell, or run modes.

#### Scenario: Shell mode
- **WHEN** `asylum shell` is run
- **THEN** no onboarding occurs

#### Scenario: Agent mode
- **WHEN** `asylum` is run in agent mode
- **THEN** onboarding runs between container start and agent exec

### Requirement: Only on first container start
Onboarding SHALL only run when a new container is created, not when exec'ing into an already-running container.

#### Scenario: New container
- **WHEN** no container is running and one is created
- **THEN** onboarding runs after the container is ready

#### Scenario: Existing container
- **WHEN** the container is already running (second session)
- **THEN** onboarding is skipped
