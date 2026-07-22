# Browser-Open Kit

Open URLs from inside the container in your real browser on the host.

**Activation: Always on** — active in every container, opt-out via config.

## What It Does

Agents often print a URL — a dev-server preview, a generated report, an OAuth login link. Many agents render a full-screen TUI that disables text selection in the terminal, so you cannot even copy the URL. This kit lets the agent open it for you: running `open <url>` (or `xdg-open <url>`, or anything that honours `$BROWSER`) inside the container opens the URL in your host's default browser.

Only `http` and `https` URLs are opened.

This is unrelated to the [agent-browser](agent-browser.md) kit: agent-browser drives a headless browser for the agent to read pages; browser-open shows a page to *you*.

## Configuration

Opt out by disabling the kit:

```yaml
kits:
  browser-open:
    disabled: true
```

## How It Works

The kit installs a shim at `/usr/local/bin/asylum-open` (exposed as `open`, `xdg-open`, and `sensible-browser`, with `BROWSER` pointing at it). The shim forwards the URL to the **host broker** — a small HTTP server Asylum runs on the host for the container's lifetime. The broker validates the URL and opens it with `open` (macOS) or `xdg-open` (Linux).

The broker never binds a publicly reachable address. On a native Linux engine it listens on a **Unix domain socket** bind-mounted only into that one container; on Docker Desktop and macOS it listens on **`127.0.0.1`**, which the container reaches via `host.docker.internal`. Either way it is unreachable from other hosts and from sibling containers. A per-container token authenticates every request as defense-in-depth. The broker starts automatically when the container starts, is respawned by any session if it dies, and exits when the container stops.

## OAuth login flows

Many CLI logins (`gh auth login`, `gcloud auth login`, `vercel login`, …) start a callback server on `localhost:<port>` **inside the container**, then open a provider URL that redirects back to that port after you sign in. Because the browser runs on your host, that redirect would normally hit the host's `localhost` where nothing is listening.

When the opened URL carries a loopback `redirect_uri` (`localhost`, `127.0.0.1`, or `::1`) with an explicit port, the broker briefly bridges that host port into the container so the callback lands. The bridge:

- binds **host loopback only** (never the LAN) and relays into the container over `docker exec` — it never publishes a host port or starts another container;
- lives **5 minutes**, with the timer reset if the same port is opened again;
- is **best-effort** — if the host port is already in use it is skipped, and you can still finish the flow by pasting the code or callback URL as the tool offers.
