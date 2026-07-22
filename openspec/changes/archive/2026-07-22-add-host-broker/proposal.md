## Why

Many agents render a full-screen TUI in the terminal, which disables text selection in most terminal emulators. When an agent prints a URL — a dev-server preview, an OAuth login link, a generated HTML report — the user often **cannot even copy it**, let alone open it. The container has no browser and no display, and there is currently no `open`/`xdg-open`/`$BROWSER` mechanism, so the agent flails (tries `xdg-open`, errors, guesses) and the user is stuck.

The clean fix is to let the sandbox ask the **host** to open a URL in the real browser. Opening an `http(s)` URL is a capability every sandboxed app already has on every OS, so this is a friction fix, not a new trust boundary. We want the underlying transport to be a **core mechanism any kit can use**, with browser-opening as its first consumer.

## What Changes

- **New core mechanism: a host broker.** A small host-side HTTP server, scoped to a container's lifetime, that serves token-authenticated routes contributed by kits. It is started idempotently by any asylum session (create *or* attach), self-terminates when the container stops, and is respawned by the next session if it ever dies. Its port and per-container token are baked into the container's environment at creation.
- **Kits can register broker routes.** The `Kit` struct gains a route-contribution point; the broker collects routes from enabled kits, wraps each in token auth, and serves them. The server only starts when at least one route is registered.
- **New on-by-default kit: `browser-open`.** Registers `POST /open`, and ships container-side shims (`/usr/local/bin/asylum-open`, plus `open`, `xdg-open`, `sensible-browser`) and `BROWSER=/usr/local/bin/asylum-open`. Any tool that opens a browser now reaches the host. The handler validates the URL (`http`/`https` only) and execs the host opener (`open` on macOS, `xdg-open` on Linux) with no shell. Opt-out via `.asylum`. Independent of the `agent-browser` kit.

## Capabilities

### New Capabilities
- `host-broker`: A container-lifetime, token-authenticated host HTTP server that serves routes contributed by enabled kits, with an idempotent start/respawn and container-scoped teardown.
- `browser-open-kit`: An on-by-default kit that registers the broker's `/open` route and installs container shims so any URL-opening tool opens the URL in the host's real browser.

## Impact

- **New `internal/broker/`**: HTTP server, token auth, route registry, `ensureBroker(cname)` (idempotent spawn), and the hidden `asylum __broker` subcommand that binds the baked port, serves, and self-exits on `docker wait <cname>`.
- **New `internal/kit/browser_open.go`**: the `browser-open` kit (Dockerfile shims, `$BROWSER`, rules snippet, and the `/open` route handler).
- **`internal/kit/kit.go`**: add a route-contribution field to `Kit` (`[]broker.Route` or equivalent).
- **`cmd/asylum/main.go`**: call `ensureBroker` on both the create path (after `RunDetached`) and the attach path; add `__broker` dispatch. Bake `ASYLUM_BROKER_HOST/PORT/TOKEN` env at container creation (attach path reads them back via `docker inspect`).
- **`internal/container/container.go`**: emit the broker env `-e` args at `RunArgs` time.
- **`internal/docker/docker.go`**: helper to read container env (`docker inspect`) and to `docker wait`.
- **Out of scope (possible follow-ups):** binding the broker to the docker bridge IP instead of `0.0.0.0` on Linux (LAN-exposure hardening); additional routes (`/notify`, `/clipboard`, `/reveal`); a mounted `broker.json` rendezvous file to allow re-binding on a different port.
