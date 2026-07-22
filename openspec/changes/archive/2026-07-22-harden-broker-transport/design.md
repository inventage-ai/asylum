## Context

The broker (shipped in `add-host-broker`) binds `0.0.0.0:<port>` and authenticates with a 256-bit per-container token. The `0.0.0.0` bind was forced by a reachability asymmetry: on native Linux, `--add-host host.docker.internal:host-gateway` points the container at the **bridge gateway IP**, so a host process bound to `127.0.0.1` is unreachable from the container. On Docker Desktop, `host.docker.internal` proxies to host loopback, so `127.0.0.1` *is* reachable. `0.0.0.0` was the one bind that worked on both — at the cost of exposing the port to sibling containers and the LAN.

The insight for this change: the two platform families have **complementary** secure transports, so we never need `0.0.0.0`.

## Goals / Non-Goals

**Goals:**
- No broker interface reachable beyond the host, on any supported platform.
- Preserve the current behaviour the user verified: browser-open works on Docker Desktop for Mac and on native Linux.
- Keep the token as defense-in-depth.
- Misdetection degrades to "feature unavailable", never to "exposed bind".

**Non-Goals:**
- Remote `DOCKER_HOST` / rootless-over-TCP daemons (browser-open assumes the daemon and the browser share a host).
- TLS (unnecessary once nothing is exposed beyond the host).
- Changing token generation, lifetime, self-healing, or the kit-route mechanism.

## Decisions

- **Transport is chosen by the host, at broker-start time, from `runtime.GOOS` + engine kind:**
  - `darwin` → **loopback TCP** (`127.0.0.1`). macOS always runs the engine in a VM; a bind-mounted host socket cannot cross the VM boundary.
  - `linux` + native engine → **Unix domain socket**. Host and container share a kernel, so a socket in a bind-mounted directory is directly connectable.
  - `linux` + VM-backed engine (Docker Desktop for Linux, Colima, etc.) → **loopback TCP**.
- **Native-vs-VM detection: `docker info` `OperatingSystem`.** A VM-backed engine reports `Docker Desktop`; a native engine reports the host distro (e.g. `Debian GNU/Linux 13`). `docker.IsDesktop()` returns true when the field contains `Docker Desktop`. Detection runs once per invocation; the result is cheap and stable.
- **Unix transport mounts the *directory*, not the socket file.** The broker starts *after* the container (it needs the container name for `docker wait` teardown), so the socket does not exist at `docker run` time. Bind-mounting the socket path directly would make Docker create an empty directory at the source. Instead the per-container directory `~/.asylum/projects/<cname>/` is created up front and bind-mounted at a fixed container path; the broker later binds `broker.sock` inside it, and the socket appears live in the container (shared kernel). `ASYLUM_BROKER_SOCK` holds the container-side socket path.
- **`Endpoint{Network, Addr}` abstraction in the broker.** `Serve`/`alive`/`EnsureBroker` take an `Endpoint` (`{"unix", "/path/broker.sock"}` or `{"tcp", "127.0.0.1:<port>"}`). The `__broker` subcommand receives `--net`/`--addr` and reconstructs it; the token still travels via the environment. For the Unix listener, `alive` dials the socket with an `http.Client` whose transport `DialContext`s the socket path.
- **Socket hygiene.** Before listening, the broker unlinks any stale `broker.sock`; on exit (container stop) it removes the socket. A leftover socket from a crash is harmless — the next `EnsureBroker` unlinks it.
- **No `0.0.0.0`, no exposed fallback.** If the transport is misdetected, the shim's connection simply fails and the agent falls back to printing the URL. There is no automatic exposed-bind fallback — losing the feature is acceptable; re-exposing the port is not.

## Risks / Trade-offs

- **Engine-kind misdetection on Linux.** An unusual engine that neither reports `Docker Desktop` nor behaves like a native shared-kernel daemon could get the wrong transport and the feature would silently not work. Mitigation: the failure mode is graceful (URL printed, not an exposed port), and `docker info` covers the common native and Desktop cases. If real setups slip through, a config override can be added later (consult first — no speculative config).
- **Socket permissions.** The container connects as the aligned host UID (`host-user-alignment`), so it has access to a socket the host user created. If UID alignment is off, connect could be denied — same class of constraint the project already relies on elsewhere.
- **Two code paths to maintain.** Unavoidable given the platform asymmetry; the `Endpoint` abstraction keeps the divergence to endpoint construction plus one `--unix-socket` branch in the shim.
- **Verification needs both platforms.** Must be checked on Docker Desktop for Mac *and* a native Linux engine before archiving; unit tests cannot exercise the container round-trip.
