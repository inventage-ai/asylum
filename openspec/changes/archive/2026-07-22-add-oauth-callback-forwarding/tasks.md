## 1. Broker context for route handlers

- [x] 1.1 Define `Ctx` in `internal/broker` exposing `Container() string` and `ForwardLoopback(port int, ipv6 bool)`; change `Route.Handler` to `func(Ctx, http.ResponseWriter, *http.Request)`.
- [x] 1.2 `Serve` gains the container name and a forwarding manager; it wraps each route (and `/healthz`) with token auth and supplies the `Ctx`.
- [x] 1.3 `cmd/asylum/main.go` `runBroker`: pass `cname` into `broker.Serve`.

## 2. Forwarding manager (`internal/broker/forward.go`)

- [x] 2.1 Manager keyed by port, guarded by a mutex; each entry holds the host listener and a 5-minute timer.
- [x] 2.2 `ForwardLoopback(port, ipv6)`: if the port is already forwarded, reset its timer; else bind the host loopback address (`127.0.0.1:<port>` or `[::1]:<port>`). On bind failure, log and return (best-effort).
- [x] 2.3 Accept loop: per connection, `docker exec -i <cname> socat - TCP4:127.0.0.1:<port>` (or `TCP6:[::1]:<port>`), copying bytes both directions until either side closes.
- [x] 2.4 On timer expiry, close the host listener and drop the entry; on manager shutdown, close all listeners.

## 3. Loopback detection in `/open` (`internal/kit/browser_open.go`)

- [x] 3.1 `detectLoopbackCallback(rawURL) (port int, ipv6 bool, ok bool)`: parse the URL, read `redirect_uri`/`redirect_url`, require a loopback host (`localhost`/`127.0.0.1` → IPv4, `::1` → IPv6) with an explicit port.
- [x] 3.2 Update `openHandler` to the new `Ctx` signature; on a positive detection call `ctx.ForwardLoopback(port, ipv6)` before opening; keep the open unconditional and best-effort.

## 4. Verification

- [x] 4.1 `go build ./...`, `go vet ./...`, `go test ./...` pass; unit tests for `detectLoopbackCallback` (IPv4, IPv6, `localhost`, encoded `redirect_uri`, non-loopback, missing port, no param) and for the manager's timer reset + bind-failure skip.
- [x] 4.2 Manual (Docker Desktop for Mac): run a real loopback OAuth flow (e.g. `gh auth login --web` or `gcloud auth login`) inside the container; confirm the browser callback completes without pasting. *(User-verified.)*
- [x] 4.3 Manual (native Linux engine): same flow completes; confirm the host binds only loopback (`127.0.0.1`/`[::1]`), no `-p`, no sidecar container. *(User-verified.)*
- [x] 4.4 Manual: forward expires after 5 minutes idle; a repeat open of the same port resets it; forwards vanish when the container stops. *(User-verified.)*
- [x] 4.5 Docs: `docs/kits/browser-open.md` auth-callback note; `security-model.md` loopback-forward note.
