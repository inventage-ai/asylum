## ADDED Requirements

### Requirement: Echo agent for testing
The agent package SHALL include an `echo` agent that runs the shell `echo` command with any provided args. It requires no CLI installation and no config directory.

#### Scenario: Echo agent with args
- **WHEN** `asylum -a echo -- hello world` is run
- **THEN** the container executes `echo hello world` and exits with code 0

#### Scenario: Echo agent without args
- **WHEN** `asylum -a echo` is run
- **THEN** the container executes `echo` (prints empty line) and exits with code 0

### Requirement: E2e test binary build
The e2e test suite SHALL compile the asylum binary once per test run and reuse it across all test cases.

#### Scenario: Binary builds successfully
- **WHEN** the e2e test suite starts
- **THEN** it builds the asylum binary to a temp directory and all subsequent tests use that binary

### Requirement: E2e test with minimal config
The e2e test suite SHALL use a minimal config (`kits: {}`, `agents: {}`, `agent: echo`) to minimize image build time.

#### Scenario: Minimal image build
- **WHEN** the e2e tests run with empty kits and agents
- **THEN** the base image contains only core OS tools (no language kits, no agent CLIs)

### Requirement: Help and version output
The binary SHALL print help text on `--help` and version on `--version`, both exiting with code 0.

#### Scenario: Help flag
- **WHEN** `asylum --help` is run
- **THEN** stdout contains usage text and the process exits 0

#### Scenario: Version flag
- **WHEN** `asylum --version` is run
- **THEN** stdout contains the version string and the process exits 0

### Requirement: Container lifecycle in e2e
The e2e tests SHALL verify that a container is started, commands are executed, and the container is cleaned up after the last session.

#### Scenario: Run mode executes and cleans up
- **WHEN** `asylum run echo ok` is run
- **THEN** "ok" appears in stdout and no asylum container is left running afterward

#### Scenario: Agent mode with echo agent
- **WHEN** `asylum -a echo -- hello` is run
- **THEN** "hello" appears in stdout and the container is cleaned up
