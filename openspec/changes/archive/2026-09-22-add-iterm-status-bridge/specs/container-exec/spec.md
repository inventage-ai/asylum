## ADDED Requirements

### Requirement: Per-session environment injection at exec

A container is created once and serves many sessions, so values baked in at container creation describe the first session rather than the current one. The system SHALL support setting environment variables at exec time, scoped to the single session being started, for values that are properties of the terminal session rather than of the container.

Per-session variables SHALL be set for shell, command, and agent exec modes alike. When the host value a variable derives from is absent, the system SHALL omit the variable rather than set it empty, so consumers can treat absence as "not applicable".

#### Scenario: Value reflects the starting session

- **WHEN** two sessions attach to the same running container from different terminal sessions
- **THEN** each session's process environment carries the value derived from its own terminal, not the other's

#### Scenario: Absent host value omits the variable

- **WHEN** a session is started on a host where the source value is not set
- **THEN** the variable is absent from that session's environment rather than present and empty

#### Scenario: Container-wide environment unchanged

- **WHEN** per-session variables are injected for a session
- **THEN** the container's own environment, as recorded at creation, is unchanged and other sessions are unaffected
