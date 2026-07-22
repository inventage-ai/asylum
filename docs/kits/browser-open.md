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

The broker is authenticated with a per-container token that only that container's environment carries, so other containers on the same Docker network cannot use it. It starts automatically when the container starts, is respawned by any session if it dies, and exits when the container stops.
