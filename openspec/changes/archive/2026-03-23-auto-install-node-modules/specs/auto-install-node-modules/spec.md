## ADDED Requirements

### Requirement: Detect Node.js projects with lockfiles
The system SHALL scan for `package.json` files using `FindNodeModulesDirs` and check each project directory for a lockfile to determine the install command.

#### Scenario: npm project
- **WHEN** a directory contains `package-lock.json`
- **THEN** the install command is `npm ci`

#### Scenario: pnpm project
- **WHEN** a directory contains `pnpm-lock.yaml`
- **THEN** the install command is `pnpm install --frozen-lockfile`

#### Scenario: yarn project
- **WHEN** a directory contains `yarn.lock`
- **THEN** the install command is `yarn install --frozen-lockfile`

#### Scenario: bun project
- **WHEN** a directory contains `bun.lock` or `bun.lockb`
- **THEN** the install command is `bun install --frozen-lockfile`

#### Scenario: No lockfile
- **WHEN** a directory has `package.json` but no lockfile
- **THEN** no install command is generated for that directory

### Requirement: Prompt user before installing
In agent mode, the system SHALL display all detected projects with their install commands in a single consolidated prompt and ask for confirmation before proceeding.

#### Scenario: User accepts
- **WHEN** the user presses Enter or types anything other than "n"
- **THEN** install commands for all listed projects are queued

#### Scenario: User declines
- **WHEN** the user types "n" or "N"
- **THEN** no install commands are queued and the agent starts without installs

#### Scenario: No projects detected
- **WHEN** no lockfiles are found in any Node.js project directory
- **THEN** no prompt is shown

### Requirement: Install runs inside the container
The install commands SHALL run inside the container as part of the `docker exec` session, before the agent binary starts.

#### Scenario: Successful install
- **WHEN** install commands are queued
- **THEN** they run with PATH set up for fnm/node, then the agent starts via `exec`

#### Scenario: Install failure
- **WHEN** an install command fails
- **THEN** the agent still starts (install failures are non-fatal)

### Requirement: Feature can be disabled
The auto-install feature SHALL be disabled when `features: { auto-install-node-modules: false }` is set in config.

#### Scenario: Feature disabled
- **WHEN** config has `features: { auto-install-node-modules: false }`
- **THEN** no lockfile detection or prompting occurs

#### Scenario: Feature enabled by default
- **WHEN** config does not mention `auto-install-node-modules`
- **THEN** the auto-install behavior is active in agent mode

### Requirement: Only in agent mode
The auto-install prompt SHALL only appear in agent mode, not in shell, admin shell, or run modes.

#### Scenario: Shell mode
- **WHEN** `asylum shell` is run
- **THEN** no install prompt is shown

#### Scenario: Agent mode
- **WHEN** `asylum` is run (default agent mode)
- **THEN** the install prompt is shown if lockfiles are detected
