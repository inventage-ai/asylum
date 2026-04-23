## ADDED Requirements

### Requirement: Pi agent registration
The agent package SHALL register a `pi` agent in the global agents map, accessible via `agent.Get("pi")`.

#### Scenario: Pi agent lookup
- **WHEN** `Get("pi")` is called
- **THEN** it returns the Pi agent implementation with no error

### Requirement: Pi agent install definition
The agent package SHALL register an `AgentInstall` for pi with a Dockerfile snippet that installs pi via npm (through fnm-managed node) and a banner line for the welcome screen. The install SHALL declare `node` as a kit dependency.

#### Scenario: Pi install resolution
- **WHEN** `ResolveInstalls` is called with `{"pi": true}` and `["node"]` in active kits
- **THEN** it returns the pi AgentInstall with no error

#### Scenario: Pi install without node kit
- **WHEN** `ResolveInstalls` is called with `{"pi": true}` but `node` is not in active kits
- **THEN** it emits a warning that pi requires the node kit

### Requirement: Pi command generation
Pi SHALL generate commands with the correct binary name and argument forwarding. Extra args are passed through; resume is handled via pi's session mechanism.

#### Scenario: Default without resume
- **WHEN** Command is called with resume=false and no extra args
- **THEN** command runs `pi` with no resume flag

#### Scenario: With resume
- **WHEN** Command is called with resume=true and no extra args
- **THEN** command runs `pi` with resume behavior

#### Scenario: With extra args
- **WHEN** Command is called with resume=false and extra args `["fix the bug"]`
- **THEN** command runs `pi fix the bug`

### Requirement: Pi session detection
Pi SHALL detect existing sessions by checking for session files in its config directory.

#### Scenario: Pi session exists
- **WHEN** pi's config directory contains session data for the project
- **THEN** HasSession returns true

#### Scenario: No pi session
- **WHEN** pi's config directory is empty or missing
- **THEN** HasSession returns false
