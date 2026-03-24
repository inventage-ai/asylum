## MODIFIED Requirements

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
