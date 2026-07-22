## Why

The host broker binds `0.0.0.0:<port>` and serves the per-container token over cleartext HTTP. That listener is reachable by sibling containers on the Docker network and by the host's LAN, mitigated only by a 256-bit token — a deliberate tradeoff taken because on native Linux `host.docker.internal` resolves to the bridge gateway, so a `127.0.0.1` bind is unreachable from the container. A background security review flagged the non-loopback bind (plaintext transport on a non-loopback interface).

We can remove the network exposure entirely by choosing the transport per host platform, because each platform's secure option is the other's broken one:

| Host | Transport | Reachable from container | Exposure beyond host |
|------|-----------|--------------------------|----------------------|
| Native Linux engine (shared kernel) | **Unix domain socket**, bind-mounted into that container only | ✅ shared kernel | none (no network) |
| Docker Desktop / VM-backed (macOS, Windows, Desktop-on-Linux) | **`127.0.0.1` TCP**, reached via `host.docker.internal` | ✅ VM proxies host loopback | none (loopback only) |

A bind-mounted Unix socket does **not** work on Docker Desktop (the socket's kernel endpoint lives on the host, and the VirtioFS/gRPC-FUSE file share does not proxy socket semantics across the VM boundary), which is exactly why the loopback-TCP path is used there. Neither transport is reachable beyond the host; the token remains as defense-in-depth.

## What Changes

- **Transport is selected by host platform**, replacing the fixed `0.0.0.0` bind:
  - Native Linux engine → a Unix domain socket under `~/.asylum/projects/<cname>/`, with the containing directory bind-mounted into the container. The shim connects via `curl --unix-socket`.
  - Docker Desktop / VM-backed engines (always on macOS; detected on Linux) → `127.0.0.1:<port>` TCP, reached via `host.docker.internal` as today but loopback-only.
- **Container env reflects the transport**: `ASYLUM_BROKER_SOCK` (socket path) for the Unix transport; `ASYLUM_BROKER_HOST`/`ASYLUM_BROKER_PORT` for the TCP transport. `ASYLUM_BROKER_TOKEN` stays for both.
- The `0.0.0.0` bind is removed. A misdetected transport **degrades gracefully** (the shim's connection fails and the agent falls back to printing the URL) — it never falls back to an exposed bind.

## Capabilities

### Modified Capabilities
- `host-broker`: The broker binds a platform-selected, host-only endpoint (Unix socket on native Linux, `127.0.0.1` TCP on VM-backed engines) instead of `0.0.0.0`, and bakes transport-specific connection parameters into the container environment.
- `browser-open-kit`: The container shim forwards to the broker over whichever endpoint the environment describes — a Unix socket (`--unix-socket`) or a TCP host:port.

## Impact

- `internal/broker/broker.go`: introduce an `Endpoint{Network, Addr}` describing how to listen and dial; `Serve`, `alive`, and `EnsureBroker` take an `Endpoint`. Unix listeners unlink a stale socket before `Listen` and remove it on exit.
- `internal/docker/docker.go`: `IsDesktop()` helper (reads `docker info` `OperatingSystem`) to distinguish a native Linux engine from a VM-backed one.
- `cmd/asylum/main.go`: compute the transport once (host `runtime.GOOS` + `IsDesktop()`), thread the endpoint into `RunArgs` (env + socket-dir mount) and into the `__broker` spawn; `runBroker` reconstructs the endpoint from its flags.
- `internal/container/container.go`: `RunOpts` carries the transport choice; `RunArgs` bakes `ASYLUM_BROKER_SOCK` **or** `ASYLUM_BROKER_HOST`/`PORT`, and adds the socket-directory bind mount on the Unix transport.
- `internal/kit/browser_open.go`: the `asylum-open` shim uses `--unix-socket "$ASYLUM_BROKER_SOCK"` when set, else the TCP URL.
- **Out of scope:** remote `DOCKER_HOST` setups (the broker must run on the same host as the browser; browser-open is not supported against a remote daemon).
