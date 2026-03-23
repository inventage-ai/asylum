## ADDED Requirements

### Requirement: Detect Node.js projects with lockfiles
The npm onboarding task SHALL use `FindNodeModulesDirs` to locate package.json files and check each directory for a lockfile to determine the install command.

#### Scenario: npm project
- **WHEN** a directory contains `package-lock.json`
- **THEN** a workload with command `npm ci` is returned

#### Scenario: pnpm project
- **WHEN** a directory contains `pnpm-lock.yaml`
- **THEN** a workload with command `pnpm install --frozen-lockfile` is returned

#### Scenario: yarn project
- **WHEN** a directory contains `yarn.lock`
- **THEN** a workload with command `yarn install --frozen-lockfile` is returned

#### Scenario: bun project
- **WHEN** a directory contains `bun.lock` or `bun.lockb`
- **THEN** a workload with command `bun install --frozen-lockfile` is returned

#### Scenario: No lockfile
- **WHEN** a directory has `package.json` but no lockfile
- **THEN** no workload is returned for that directory

### Requirement: Lockfile hash for change detection
Each workload SHALL use the lockfile content hash as its input hash, so dependencies are only reinstalled when the lockfile changes.

#### Scenario: Lockfile unchanged
- **WHEN** the stored hash matches the current lockfile
- **THEN** the workload is skipped

#### Scenario: Lockfile changed
- **WHEN** the stored hash differs from the current lockfile
- **THEN** the workload is pending and included in the prompt

### Requirement: Per-task config
The npm task SHALL be disabled when `onboarding: { npm: false }` is set in config.

#### Scenario: Disabled
- **WHEN** config has `onboarding: { npm: false }`
- **THEN** no npm workloads are detected or prompted
