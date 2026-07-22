## MODIFIED Requirements

### Requirement: Host open route
The kit SHALL register a `POST /open` route on the broker. The handler SHALL accept a `url`, SHALL open it in the host's default browser using the host opener (`open` on macOS, `xdg-open` on Linux) with the URL passed as a single argument and without a shell. If the URL carries a `redirect_uri` or `redirect_url` query parameter whose host is a loopback address (`localhost`, `127.0.0.1`, or `::1`) with an explicit port, the handler SHALL additionally request a loopback callback forward for that port via the broker context. Requesting the forward SHALL be best-effort: whether or not it succeeds, the open SHALL still proceed.

#### Scenario: URL opens on the host
- **WHEN** the broker receives an authenticated `/open` request for `http://localhost:7036`
- **THEN** the host's default browser is opened at that URL

#### Scenario: Loopback redirect_uri triggers a forward
- **WHEN** the opened URL contains `redirect_uri=http://127.0.0.1:54321/callback`
- **THEN** the handler requests a loopback callback forward for port `54321` in addition to opening the URL

#### Scenario: Non-loopback redirect_uri ignored
- **WHEN** the opened URL contains a `redirect_uri` whose host is not a loopback address, or has no explicit port
- **THEN** no forward is requested and the URL is opened normally

#### Scenario: Forward failure does not block the open
- **WHEN** a loopback `redirect_uri` is detected but the forward cannot be set up
- **THEN** the URL is still opened and the request succeeds
