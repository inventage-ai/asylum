## MODIFIED Requirements

### Requirement: Dynamic cache directories
Cache directory volume mounts SHALL use paths resolved from the host user's home directory, not hardcoded `/home/claude` paths.

#### Scenario: Maven cache on macOS
- **WHEN** the host home is `/Users/simon` and the java/maven kit is active
- **THEN** the maven cache is mounted at `/Users/simon/.m2`

### Requirement: Container environment variables use dynamic paths
Container environment variables that reference the home directory SHALL use the host home directory path.

#### Scenario: HISTFILE path
- **WHEN** the container is started
- **THEN** HISTFILE is set to `<host-home>/.shell_history/zsh_history`

#### Scenario: Agent config dir
- **WHEN** the Claude agent config is mounted
- **THEN** CLAUDE_CONFIG_DIR is set to `<host-home>/.claude`
