## Why

`asylum-open` (and its `open`/`xdg-open`/`sensible-browser` aliases) is the only way for an agent inside the container to hand something to the host desktop, but the broker's `/open` handler hard-rejects everything that is not `http` or `https`. Host tools that integrate through a custom URL scheme — Dropshare's `dropshare5:///`, editor handoffs like `vscode://`, note tools like `obsidian://` — are therefore unreachable from the sandbox, even when the user explicitly wants them. Users need a way to opt specific schemes in without weakening the default.

## What Changes

- Add a `schemes` option to the `browser-open` kit config (`kits.browser-open.schemes`) listing extra URL schemes the `/open` route accepts, in addition to the always-allowed `http` and `https`.
- Scheme entries are normalized (lowercased, `:` and trailing `//` stripped) and validated against RFC 3986 scheme syntax; invalid entries are reported and ignored rather than silently widening or breaking the open path.
- The allowlist accumulates across config layers (global `~/.asylum/config.yaml` + project `.asylum` + `.asylum.local`), matching how `packages` and `build` already merge.
- The broker process learns the effective allowlist at spawn time, so the host-side handler can validate against it. Non-`http(s)` allowlisted URLs are accepted without requiring a host component (`dropshare5:///path` is valid).
- The sandbox rules snippet tells the agent which extra schemes are available, so it knows `open dropshare5://…` will work.
- Docs (`docs/kits/browser-open.md`, `docs/concepts/security-model.md`) describe the option and the trust implications of opting a scheme in.
- Default behaviour is unchanged: with no `schemes` configured, only `http`/`https` are opened.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `browser-open-kit`: the URL scheme validation requirement gains a configurable allowlist; the kit gains a `schemes` config option and advertises configured schemes in its rules snippet.
- `host-broker`: kit route handlers need their kit's effective configuration in the detached broker process; the broker spawn contract gains a configuration channel.

## Impact

- `internal/kit/browser_open.go` — `validBrowserURL` becomes allowlist-aware; `RulesSnippet` becomes `RulesSnippetFunc`; handler reads the effective allowlist.
- `internal/kit/kit.go` — `SnippetConfig` gains a `Schemes` field.
- `internal/config/config.go` — `KitConfig` gains `Schemes []string` (`merge:"concat"`); `KitSnippetConfig` passes it through.
- `internal/broker/broker.go` — `EnsureBroker` gains a way to pass kit configuration to the spawned broker process.
- `cmd/asylum/main.go` — `ensureBroker`/`runBroker` plumb the allowlist between the session process and the detached broker.
- Docs: `docs/kits/browser-open.md`, `docs/concepts/security-model.md`, `CHANGELOG.md`.
- No image rebuild is triggered: the container shim is unchanged and validation is host-side.
