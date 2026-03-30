## MODIFIED Requirements

### Requirement: Container cleanup after last session
After any exec'd session exits, asylum SHALL check if other sessions remain in the container by querying active exec processes and remove the container if none do.

#### Scenario: Last session exits
- **WHEN** the last exec'd session in a container exits
- **THEN** asylum runs a process check inside the container, finds no other exec sessions, and removes the container

#### Scenario: Other sessions still running
- **WHEN** an exec'd session exits but other sessions are still running
- **THEN** asylum runs a process check inside the container, detects the other sessions, and leaves the container running

#### Scenario: Container already dead at cleanup
- **WHEN** an exec'd session exits and the container has already stopped
- **THEN** asylum treats it as "no sessions" and attempts removal (which is a no-op)

#### Scenario: Background tasks do not prevent cleanup
- **WHEN** the last exec'd session exits but user-started background processes (e.g., dev servers) are still running inside the container
- **THEN** asylum removes the container because background tasks are not exec sessions

## ADDED Requirements

### Requirement: Runtime exec session detection
The docker package SHALL provide a function to detect whether other exec sessions are active in a container by running `ps` inside the container and checking for processes with PPID=0 (excluding PID 1 and the check process itself).

#### Scenario: No other sessions
- **WHEN** `HasOtherSessions` is called and only the check process has PPID=0
- **THEN** it returns `false`

#### Scenario: Other sessions active
- **WHEN** `HasOtherSessions` is called and other exec sessions have PPID=0
- **THEN** it returns `true`

#### Scenario: Container not reachable
- **WHEN** `HasOtherSessions` is called but the container is stopped or missing
- **THEN** it returns `false`

### Requirement: SIGHUP signal forwarding
The CLI SHALL forward SIGHUP (in addition to SIGINT and SIGTERM) to the docker exec process so agents receive clean shutdown signals on terminal close.

#### Scenario: Terminal closes during session
- **WHEN** the terminal is closed while an agent session is running
- **THEN** SIGHUP is forwarded to the docker exec process

## REMOVED Requirements

### Requirement: Session counter file
**Reason**: Replaced by runtime exec session detection. The file-based counter was susceptible to corruption on unclean exits.
**Migration**: No migration needed. Stale counter files in `~/.asylum/projects/<container>/sessions` can be safely ignored or cleaned up.
