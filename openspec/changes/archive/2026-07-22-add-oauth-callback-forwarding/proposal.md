## Why

Many CLI auth flows (`gh`, `gcloud`, `vercel`, `wrangler`, …) start a **loopback callback server inside the container**, then open a provider URL whose `redirect_uri` points at `http://localhost:<port>`. We now open that URL in the host browser (browser-open kit), the user logs in, and the provider redirects the browser to `http://localhost:<port>/…` — which hits the **host's** loopback, where nothing is listening. The callback server is inside the container, so the flow stalls.

We want to bridge that specific case — a loopback `redirect_uri` — without letting the agent open permanent or arbitrary host ports. It is a **best-effort** convenience: most flows also accept pasting a code or the callback URL, so if the bridge can't be set up (e.g. the host port is taken) the flow still completes manually.

## What Changes

- **The `/open` handler detects a loopback `redirect_uri`.** When an opened URL carries a `redirect_uri`/`redirect_url` query parameter whose host is `localhost`, `127.0.0.1`, or `::1` with an explicit port, the broker sets up a temporary forward for that port (in addition to opening the URL). Non-loopback or port-less redirects are ignored.
- **The broker forwards host loopback → container loopback via `docker exec`.** The broker (a host process, scoped to the container) binds `127.0.0.1:<port>` (or `[::1]:<port>` for `::1`) on the host and, per connection, tunnels bytes into the container with `docker exec -i <cname> socat - TCP:<loopback>:<port>`. `socat` is already in the base image. No sidecar container, no `-p` publishing, no persistent in-container listener.
- **Best-effort and time-boxed.** If the host port can't be bound, the broker logs and skips — the open still succeeds. Each forward lives **5 minutes**; a repeat detection of the same port **resets** the timer. Forwards are broker state and die when the broker/container stops.
- **Loopback-only exposure.** The forward binds host loopback (the browser is always host-local), never `0.0.0.0`, so it is never reachable from the LAN. It exposes only the agent's *own* container port back to the host — the agent cannot reach host services (a taken port just fails the bind).
- **Route handlers gain a broker context.** To let the `/open` handler request a forward, the kit-route mechanism passes handlers a broker-supplied context exposing the container identity and a loopback-forward request.

## Capabilities

### Modified Capabilities
- `host-broker`: Adds a best-effort, time-boxed loopback-callback forwarding service (host loopback → container loopback over `docker exec` + `socat`), and passes route handlers a broker context (container identity + forward request).
- `browser-open-kit`: The `/open` route detects a loopback `redirect_uri` and requests a callback forward for its port, in addition to opening the URL; failure to forward never fails the open.

## Impact

- `internal/broker/`: new forwarding manager (`forward.go`) — per-port host loopback listener, per-connection `docker exec … socat` tunnel, 5-minute TTL with reset, teardown with the broker. `Serve` gains the container name; `Route.Handler` gains a broker `Ctx` (container name + `ForwardLoopback(port, ipv6)`), supplied by the broker.
- `cmd/asylum/main.go`: `runBroker` passes `cname` into `broker.Serve`.
- `internal/kit/browser_open.go`: `openHandler` parses `redirect_uri`/`redirect_url`, and on a loopback host with an explicit port calls `ctx.ForwardLoopback(port, ipv6)` before opening; a pure `detectLoopbackCallback(url)` helper carries the parsing (unit-tested).
- Docs: `docs/kits/browser-open.md` gains an auth-callback note; `security-model.md` notes the loopback-only, time-boxed forward.
- No shim/Dockerfile change (detection is server-side; `socat` already present).
- **Out of scope:** non-loopback redirect URIs; flows that don't put the callback in `redirect_uri`; guaranteeing the port is free (best-effort by design).
