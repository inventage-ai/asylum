## 1. Core broker package (`internal/broker`)

- [x] 1.1 `Route{Path string, Handler http.HandlerFunc}` type and a token-auth middleware (constant-time compare of `Authorization: Bearer`).
- [x] 1.2 `Server` that mounts routes + a `GET /healthz`, binds `0.0.0.0:<port>`, and serves. Bind failure (`EADDRINUSE`) is a clean "already running" exit, not an error.
- [x] 1.3 Token generation (crypto/rand) and a free-port picker (`bind :0` → read → close).
- [x] 1.4 `ensureBroker(cname, host, port, token)`: probe `GET /healthz` with token; if no live broker answers, spawn the detached `asylum __broker` process.
- [x] 1.5 Broker teardown: block on `docker wait <cname>` and exit when it returns.

## 2. `__broker` subcommand & lifecycle wiring

- [x] 2.1 Hidden `asylum __broker --container <cname> --port <p> --token <t>` dispatch in `cmd/asylum/main.go`: collect enabled kits' routes, start the server, run the `docker wait` teardown.
- [x] 2.2 Create path: pick port + token, bake `ASYLUM_BROKER_HOST/PORT/TOKEN` into run args, then `ensureBroker` after `RunDetached`.
- [x] 2.3 Attach path: read `ASYLUM_BROKER_*` from the running container (`docker inspect`), then `ensureBroker`.
- [x] 2.4 Start the broker only when ≥1 enabled kit contributes a route.

## 3. Kit route contribution

- [x] 3.1 Add a route-contribution point to `Kit` (`Routes []broker.Route` or a `RoutesFunc`); the `__broker` dispatch gathers routes from enabled kits.
- [x] 3.2 `internal/docker/docker.go`: helper to read a container's env via `docker inspect`; `docker wait` wrapper.

## 4. `browser-open` kit

- [x] 4.1 `internal/kit/browser_open.go`: `TierAlwaysOn`, opt-out config node, independent of `agent-browser`.
- [x] 4.2 Dockerfile snippet: install `/usr/local/bin/asylum-open` (curls broker `/open` with token from env), symlink `open`/`xdg-open`/`sensible-browser`.
- [x] 4.3 Entrypoint/env: `BROWSER=/usr/local/bin/asylum-open`.
- [x] 4.4 `/open` route handler: parse `url`, allow only `http`/`https`, `exec` `open` (darwin) / `xdg-open` (linux) via `runtime.GOOS`, URL as a single arg, no shell.
- [x] 4.5 Rules snippet: tell the agent it can open a URL in the user's browser with `open <url>`.

## 5. Verification

- [x] 5.1 `go build ./...`, `go vet ./...`, `go test ./...` pass; unit tests for token auth (accept/reject), URL scheme validation, free-port pick, and `ensureBroker` idempotency (no double-serve).
- [x] 5.2 Manual (Docker): full-screen agent runs `open http://localhost:7036` → host browser opens. *(User-verified.)*
- [x] 5.3 Manual (Docker): open two windows on one project; exit the window that started the broker → `open` still works from the other; kill the broker → next `open` triggers a respawn. *(User-verified.)*
- [x] 5.4 Manual (Docker): stop the container → broker process exits (no orphan). Sibling container without the token gets `401`. *(User-verified.)*
- [x] 5.5 Docs: `docs/kits/browser-open.md` and a note in the security model's opt-in section.
