## Context

Asylum runs one **detached** container per project (`docker run -d`, `main.go:410` `RunDetached`) and attaches every session as a separate short-lived `asylum` process doing `docker exec` (`main.go:768` `runDocker`). The container outlives any single asylum process; a second window hits `IsRunning == true` (`main.go:361`), skips `RunDetached`, and execs straight in.

Two facts about reachability shape the transport:
- The container reaches the host at `host.docker.internal` (`--add-host host.docker.internal:host-gateway`, `container.go:83`).
- On **Docker Desktop** (macOS/Windows) `host.docker.internal` proxies to host loopback, so a `127.0.0.1`-bound host service is reachable. On **native Linux** `host-gateway` is the bridge gateway IP, so a `127.0.0.1` bind is *not* reachable — the server must bind `0.0.0.0` (or the bridge IP), which also makes it reachable by sibling containers on the bridge. A per-container secret is therefore mandatory, not optional.

The friction being fixed: full-screen TUI agents disable terminal text selection, so a printed URL cannot be copied. Auto-opening on the host is the only path, and there is currently no `open`/`xdg-open`/`$BROWSER` in the container.

## Goals / Non-Goals

**Goals:**
- A host-side transport the container can call to trigger host actions, generic enough that any kit can register routes.
- Correct lifetime: the broker is available whenever the container is up, regardless of how many sessions come and go — no dependency on which session started it.
- A `browser-open` kit, on by default (opt-out), that makes any URL-opening tool open the host's real browser.
- Secure against *other* containers and the LAN via a per-container shared secret.

**Non-Goals:**
- OAuth callback plumbing (redirects to `localhost:<random>` inside the container). Solved separately by preferring device-code flows or host-side login; not this change.
- The headless agent-browser use case (already covered by the `agent-browser` kit).
- Binding to the docker bridge IP on Linux to avoid LAN exposure (noted as hardening follow-up).
- A rendezvous file allowing the broker to re-bind on a different port (follow-up).

## Decisions

- **Broker lifetime = container, not session.** A per-session goroutine would die when its `docker exec` exits, even with the container up and other windows attached — that is the *normal* path here, not an edge case. Instead the broker is a **detached host process** (`asylum __broker --container <cname>`) that binds the baked port, serves, and blocks on `docker wait <cname>`, exiting exactly when the container stops (normal exit, `cleanup`, `--rebuild`→`RemoveContainer`).
- **Idempotent `ensureBroker(cname)` on every session — create *and* attach.** Create path: generate a free port + random token, bake them into the container env, then `ensureBroker`. Attach path: read the baked port/token back via `docker inspect`, then `ensureBroker`. `ensureBroker` probes the port (token-authed `GET /healthz`); if no live broker answers, it spawns one. Spawning when one already exists is harmless — the new process's bind fails (`EADDRINUSE`) and it exits. This self-heals a crashed broker and closes the "owner exits while other windows stay open" gap.
- **Port + token baked at container creation, immutable for the container's life.** The container env is the single source of truth (`ASYLUM_BROKER_HOST=host.docker.internal`, `ASYLUM_BROKER_PORT`, `ASYLUM_BROKER_TOKEN`); the attach path recovers them from `docker inspect`. A free port is chosen at create time (`bind :0` → read port → close → bake); the tiny TOCTOU window is acceptable, and a lost race just means the broker exits on bind conflict.
- **Bind `0.0.0.0`; the token is the gate.** Required for Linux bridge reachability. The per-container random token is checked in constant time on every route; requests present it as `Authorization: Bearer <token>`. Only *this* container has the token in its env, so sibling containers and the LAN cannot use the broker.
- **Kits contribute routes.** `Kit` gains a route field (`[]broker.Route{Path, Handler}`). The broker mounts every enabled kit's routes under token auth and starts only when ≥1 route exists — projects with no broker-using kit pay nothing.
- **`browser-open` kit, `TierAlwaysOn` (opt-out).** Registers `POST /open`. Ships `/usr/local/bin/asylum-open` (curls the broker with the token) symlinked as `open`, `xdg-open`, `sensible-browser`, and sets `BROWSER=/usr/local/bin/asylum-open`. `/usr/local/bin` precedes `/usr/bin`, so the shim shadows any distro `xdg-open`.
- **`/open` handler runs on the host, validates, no shell.** Accepts `url`, rejects anything but `http`/`https` (blocks `file://`, `open -a App`, etc. — cheap defense in depth even though opening `http(s)` is deemed safe), and `exec`s `open` (macOS) or `xdg-open` (Linux) chosen by `runtime.GOOS`, passing the URL as a single argument.
- **Independent of `agent-browser`.** That kit is headless "let the agent read a page"; this is "let the human see it." Different lifetimes, different tiers, no shared code.

## Risks / Trade-offs

- **`0.0.0.0` exposes the broker port to the LAN for the container's lifetime.** Mitigated by the per-container token; matches the accepted trust model (opening `http(s)` is a normal sandboxed capability). Hardening follow-up: bind the docker bridge IP on Linux.
- **A detached host process per running container.** Small and self-cleaning via `docker wait`. If asylum is `kill -9`'d mid-spawn, the broker (already detached) keeps its own `docker wait` and still exits with the container. A broker crash is repaired by the next session's `ensureBroker`.
- **Port TOCTOU at create.** Between choosing a free port and the broker binding it, another host process could grab it; the broker then exits on bind failure and the next `ensureBroker` retries. Rare and self-correcting.
- **Baked, fixed port.** Cannot change without recreating the container. Accepted for the first step; a mounted `broker.json` rendezvous file is the escape hatch if re-binding is ever needed.
- **Shim shadowing.** Overriding `xdg-open`/`open` is intended (no real browser exists in the container), but a tool hard-coding a browser binary path bypasses the shim. Acceptable; such tools are rare and already broken in the container today.
