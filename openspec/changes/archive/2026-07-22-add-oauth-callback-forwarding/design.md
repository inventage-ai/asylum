## Context

The `host-broker` runs on the host, scoped to one container, and reachable from that container over its transport (Unix socket / loopback TCP). The `browser-open` kit's `/open` route opens URLs in the host browser. OAuth loopback flows (RFC 8252) bind a callback server on `127.0.0.1:<port>` **inside the container**, then open a provider URL with `redirect_uri=http://localhost:<port>/…`. After login the host browser is redirected to `localhost:<port>`, which hits the host's loopback — where nothing listens.

Two Docker constraints rule out the obvious fixes: you cannot add a `-p` to a running container, and the tool binds `127.0.0.1` (loopback), so a sibling container can't reach it over the bridge. The naive "publish the port" therefore needs a new container plus an in-container forwarder. But the broker is already a host process that can reach the container through the daemon, which collapses the whole thing.

## Goals / Non-Goals

**Goals:**
- Complete loopback OAuth callbacks with no manual step, for the common case.
- No permanent or arbitrary host ports; loopback-only; time-boxed.
- Reuse what exists: the per-container broker, `socat` (already in the image), `docker exec`.
- Best-effort: any failure degrades to the existing manual paste-the-code path.

**Non-Goals:**
- Non-loopback redirect URIs, or flows that don't carry the callback in `redirect_uri`.
- Guaranteeing the host port is free.
- A sidecar container or `-p` publishing (explicitly avoided).

## Decisions

- **Transport = `docker exec` tunnel, not a published port.** The broker binds the callback port on host loopback and, per accepted connection, runs `docker exec -i <cname> socat - TCP:<loopback>:<port>`, copying bytes both ways. This needs no sidecar, no `-p` (impossible on a running container anyway), and no persistent in-container listener. `docker exec` is daemon-mediated, so it works identically on Docker Desktop and native Linux.
- **Detection in `/open`.** The handler parses the opened URL's query for `redirect_uri`/`redirect_url`; if the value is a URL whose host ∈ {`localhost`, `127.0.0.1`, `::1`} with an explicit port, it requests a forward for that port. A pure helper `detectLoopbackCallback(rawURL) (port int, ipv6 bool, ok bool)` holds the logic and is unit-tested. Opening the browser and requesting the forward are independent; the open always proceeds.
- **IPv6, cheaply.** `::1` → bind host `[::1]:<port>` and tunnel with `socat … TCP6:[::1]:<port>`. `localhost`/`127.0.0.1` → bind `127.0.0.1:<port>`, tunnel `TCP4:127.0.0.1:<port>`. We do not dual-bind for `localhost`; the IPv4 loopback is the overwhelmingly common bind and keeps this a one-listener path. `::1` is handled only when the redirect explicitly uses it.
- **Route handlers get a broker `Ctx`.** `Route.Handler` becomes `func(Ctx, http.ResponseWriter, *http.Request)`, where `Ctx` exposes the container name and `ForwardLoopback(port int, ipv6 bool)`. The broker owns the forwarding manager and supplies the `Ctx`; kit handlers stay declarative and never touch Docker directly. This keeps forwarding logic and container identity inside the broker while detection stays with the kit that understands URLs.
- **Lifetime: 5-minute TTL, reset on repeat.** The forwarding manager keys forwards by port. `ForwardLoopback` creates the host listener on first request (or resets the timer if it already exists). On expiry it closes the listener; the container-side `socat` exec dies with its connection or when the container stops. All forwards die when the broker exits (container stop), because the listeners are held by the broker process and the exec children are tied to the container.
- **Best-effort, never fatal.** Bind failure (port taken), a not-yet-ready in-container listener, or a missing `socat` all degrade silently: the open returns success and the user falls back to pasting the code/URL. Detection/forward setup must not block or fail `/open`.

## Risks / Trade-offs

- **Per-connection `docker exec` overhead.** Fine for OAuth callbacks (1–2 short requests); not intended as a general high-throughput proxy.
- **Loopback family mismatch.** A tool that binds `::1` while the redirect says `localhost` (or vice-versa) could be missed by the single-family listener. Accepted: best-effort, and the explicit `::1` redirect case is covered.
- **Host-port collision across projects.** Two containers wanting the same host callback port — first-come; the second logs and skips. Best-effort by design.
- **Race: browser redirected before the in-container listener is up.** The tool starts its callback server before opening the browser, so by `/open` time it is normally listening; if not, `socat` connect fails and the user retries or pastes. Best-effort.
- **Security.** The forward is host-loopback:`P` → the agent's own container:`P`. The broker *binds* `P` (fails safely if taken) and only *connects* into its own container, so a crafted `redirect_uri` cannot hijack a host service. Loopback-only means no LAN exposure. Time-boxed to 5 minutes. This matches the trusted-host model and grants the agent nothing beyond briefly exposing a port it already controls.
- **`Route.Handler` signature change** touches the existing `/open` and `/healthz` handlers — all first-party, low-risk churn.
