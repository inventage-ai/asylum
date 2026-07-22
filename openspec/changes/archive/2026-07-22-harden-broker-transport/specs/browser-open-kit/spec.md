## MODIFIED Requirements

### Requirement: Container open shims
The kit SHALL install a container-side opener at `/usr/local/bin/asylum-open`, expose it under the conventional names `open`, `xdg-open`, and `sensible-browser`, and set `BROWSER=/usr/local/bin/asylum-open`. These SHALL take precedence over any distribution-provided opener. Invoking any of them with a URL SHALL forward the URL to the broker's `/open` route with the container's token, over whichever transport the environment describes: a Unix socket (`--unix-socket "$ASYLUM_BROKER_SOCK"`) when `ASYLUM_BROKER_SOCK` is set, otherwise the TCP endpoint `http://$ASYLUM_BROKER_HOST:$ASYLUM_BROKER_PORT`.

#### Scenario: Tool opens a browser over a Unix socket
- **WHEN** `ASYLUM_BROKER_SOCK` is set and a tool invokes `xdg-open <url>` (or `open`, `$BROWSER`, or `sensible-browser`)
- **THEN** the URL is sent to the broker's `/open` route over the Unix socket, authenticated with the container token

#### Scenario: Tool opens a browser over TCP
- **WHEN** `ASYLUM_BROKER_SOCK` is unset and `ASYLUM_BROKER_HOST`/`ASYLUM_BROKER_PORT` are set, and a tool invokes an opener with a URL
- **THEN** the URL is sent to the broker's `/open` route over `host.docker.internal`, authenticated with the container token
