# browser-open-kit Specification

## Purpose
Provides a default-on `browser-open` kit that forwards URL-open requests from inside the container to the host's default browser via the host broker.

## Requirements

### Requirement: Browser-open kit enabled by default
The system SHALL provide a `browser-open` kit that is enabled by default and can be disabled (opt-out) via project or user configuration. It SHALL be independent of the `agent-browser` kit.

#### Scenario: On by default
- **WHEN** a project has no explicit `browser-open` configuration
- **THEN** the kit is active and its broker route and container shims are present

#### Scenario: Opt-out honored
- **WHEN** the kit is disabled in configuration
- **THEN** no `/open` broker route is registered and no open shims are installed

### Requirement: Container open shims
The kit SHALL install a container-side opener at `/usr/local/bin/asylum-open`, expose it under the conventional names `open`, `xdg-open`, and `sensible-browser`, and set `BROWSER=/usr/local/bin/asylum-open`. These SHALL take precedence over any distribution-provided opener. Invoking any of them with a URL SHALL forward the URL to the broker's `/open` route with the container's token, over whichever transport the environment describes: a Unix socket (`--unix-socket "$ASYLUM_BROKER_SOCK"`) when `ASYLUM_BROKER_SOCK` is set, otherwise the TCP endpoint `http://$ASYLUM_BROKER_HOST:$ASYLUM_BROKER_PORT`.

#### Scenario: Tool opens a browser over a Unix socket
- **WHEN** `ASYLUM_BROKER_SOCK` is set and a tool invokes `xdg-open <url>` (or `open`, `$BROWSER`, or `sensible-browser`)
- **THEN** the URL is sent to the broker's `/open` route over the Unix socket, authenticated with the container token

#### Scenario: Tool opens a browser over TCP
- **WHEN** `ASYLUM_BROKER_SOCK` is unset and `ASYLUM_BROKER_HOST`/`ASYLUM_BROKER_PORT` are set, and a tool invokes an opener with a URL
- **THEN** the URL is sent to the broker's `/open` route over `host.docker.internal`, authenticated with the container token

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

### Requirement: URL scheme validation
The `/open` handler SHALL accept only `http` and `https` URLs and SHALL reject any other input (including `file://`, application launches, and shell metacharacters) with a client error, without invoking the host opener.

#### Scenario: Non-http scheme rejected
- **WHEN** the broker receives an authenticated `/open` request for `file:///etc/passwd`
- **THEN** the request is rejected and the host opener is not invoked

#### Scenario: http scheme accepted
- **WHEN** the broker receives an authenticated `/open` request for an `https://` URL
- **THEN** the host opener is invoked with that URL
