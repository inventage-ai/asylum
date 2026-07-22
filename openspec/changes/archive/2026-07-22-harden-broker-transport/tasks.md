## 1. Broker endpoint abstraction (`internal/broker`)

- [x] 1.1 Add `Endpoint{Network, Addr string}`; change `Serve`, `alive`, and `EnsureBroker` to take an `Endpoint` instead of a bare port.
- [x] 1.2 `Serve`: `net.Listen(ep.Network, ep.Addr)`; for `unix`, unlink a stale socket before listening and `os.Remove` it on shutdown.
- [x] 1.3 `alive`: dial the endpoint — TCP to `127.0.0.1:<port>`, or an `http.Client` whose transport `DialContext`s the Unix socket — with the token.
- [x] 1.4 `EnsureBroker`: spawn `__broker` with `--net`/`--addr` (token still via env); drop the port-only signature.
- [x] 1.5 Keep `FreePort` for the TCP transport; it is unused on the Unix path.

## 2. Engine detection (`internal/docker`)

- [x] 2.1 `IsDesktop() bool` — reads `docker info --format '{{.OperatingSystem}}'` and reports whether it contains `Docker Desktop`.

## 3. Transport selection & wiring (`cmd/asylum/main.go`)

- [x] 3.1 Compute the transport once: `darwin` → loopback TCP; `linux` + `IsDesktop()` → loopback TCP; `linux` native → Unix socket.
- [x] 3.2 Create path: for the Unix transport, ensure `~/.asylum/projects/<cname>/` exists and thread its bind mount + `ASYLUM_BROKER_SOCK` through `RunOpts`; for TCP, thread port + `ASYLUM_BROKER_HOST`/`PORT` as today.
- [x] 3.3 `ensureBroker` (attach + create): build the `Endpoint` from the transport and the container env, then call `broker.EnsureBroker`.
- [x] 3.4 `runBroker`: parse `--net`/`--addr`, reconstruct the `Endpoint`, serve.

## 4. Container args (`internal/container/container.go`)

- [x] 4.1 `RunOpts` carries the transport (socket path + host dir, or port); `RunArgs` bakes `ASYLUM_BROKER_SOCK` **or** `ASYLUM_BROKER_HOST`/`PORT` accordingly, plus `ASYLUM_BROKER_TOKEN`.
- [x] 4.2 On the Unix transport, add the socket-directory bind mount (`-v <hostdir>:<containerdir>`); remove any `0.0.0.0` assumptions.

## 5. Shim (`internal/kit/browser_open.go`)

- [x] 5.1 `asylum-open` uses `curl --unix-socket "$ASYLUM_BROKER_SOCK" http://localhost/open` when `ASYLUM_BROKER_SOCK` is set, else `http://$ASYLUM_BROKER_HOST:$ASYLUM_BROKER_PORT/open`.

## 6. Verification

- [x] 6.1 `go build ./...`, `go vet ./...`, `go test ./...` pass; unit tests for `Serve`/`alive` over a Unix socket (auth accept/reject) and over loopback TCP; endpoint construction per platform.
- [x] 6.2 Manual (Docker Desktop for Mac): `open http://localhost:7036` opens the host browser; confirm the broker binds `127.0.0.1` (not `0.0.0.0`). *(User-verified.)*
- [x] 6.3 Manual (native Linux engine): `open http://localhost:7036` opens the host browser via the Unix socket; confirm no TCP port is bound and a sibling container cannot reach the socket. *(User-verified.)*
- [x] 6.4 Manual: multi-window survival and respawn still hold on both transports; broker stops (and the socket is removed) when the container stops. *(User-verified.)*
- [x] 6.5 Docs: update `docs/kits/browser-open.md` and the `security-model.md` note to describe the host-only transports (no `0.0.0.0`).
