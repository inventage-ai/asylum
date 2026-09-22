## ADDED Requirements

### Requirement: Sealed per-session envelope

The broker's lifetime is the container's, not any session's, so its own environment reflects whichever session started it and is stale for every session that attaches later. A route handler that needs a fact about the *calling* session therefore cannot read it from the broker's environment, and MUST NOT simply trust a value the container asserts.

The system SHALL provide a sealed envelope carrying host-side, session-scoped facts from the asylum process that starts a session to the broker route handlers that serve it. The envelope SHALL be sealed with an authenticated cipher under a key readable only by the user's account on the host, stored outside any path the system bind-mounts into a container. Asylum SHALL seal the envelope when starting a session and place only the sealed form in that session's container environment. A handler SHALL open the envelope to recover its contents.

The container SHALL NOT be able to read the envelope's contents, produce an envelope the broker will open, or alter one it was given without detection. A request whose envelope is absent, unopenable, or tampered with SHALL be rejected and SHALL NOT cause any host process to run.

The envelope SHALL carry a structured payload so that new session-scoped facts can be added without changing any route's request shape.

#### Scenario: Handler recovers the calling session's facts

- **WHEN** a container posts to a route with the envelope its session was given
- **THEN** the handler recovers the session-scoped values that asylum sealed for that session

#### Scenario: Later sessions are not served stale facts

- **WHEN** a second session attaches to a container whose broker was started by an earlier session
- **THEN** the handler serving the second session's requests recovers the second session's facts, not the first session's

#### Scenario: Forged envelope rejected

- **WHEN** a request carries an envelope the container constructed or modified
- **THEN** the broker rejects the request and no host process runs

#### Scenario: Contents opaque to the container

- **WHEN** a session's container environment is inspected from inside the container
- **THEN** it contains the sealed form only, and the sealed values cannot be recovered from it

#### Scenario: Key not reachable from a container

- **WHEN** the sealing key exists on the host
- **THEN** it is stored outside every path the system bind-mounts into a container, with permissions restricting it to the user's account
