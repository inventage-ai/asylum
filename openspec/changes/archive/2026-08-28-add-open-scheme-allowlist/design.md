## Context

See proposal.md — Why. Three facts about the current implementation shape the approach:

- `openHandler` in `internal/kit/browser_open.go` runs **on the host**, inside the detached `asylum __broker` process. That process is spawned with only `--container`, `--net`, `--addr`, `--kits` and `ASYLUM_BROKER_TOKEN` in its environment; it never loads `~/.asylum/config.yaml` or the project `.asylum`, and it does not know the project directory.
- `broker.Route` is a static `{Path, Handler}` value registered at package init. There is no per-kit configuration hook on the route.
- The container-side shim is a `curl` one-liner baked into the image. It does no validation and must not: the container is the untrusted side, so the allowlist has to be enforced on the host.

## Goals / Non-Goals

**Goals:**
- One host-side source of truth for which schemes may be opened.
- Additive config that defaults to today's `http`/`https`-only behavior.
- No image rebuild when the allowlist changes.

**Non-Goals:**
- A generic "kit config reaches the broker" framework. Exactly one kit needs it today; the mechanism stays small and is generalized when a second one appears.
- Per-scheme argument filtering or payload inspection. Allowlisting a scheme grants the whole scheme.
- Changing the container shim, the `/open` request shape, or the OAuth loopback-forward behavior.

## Decisions

### Config lives on the kit: `kits.browser-open.schemes`

`KitConfig` gains `Schemes []string` with the `merge:"concat"` tag, so layers accumulate exactly like `packages` and `build` (see the `deep-merge-kit-config` spec — the reflection-driven merge needs no code change for a new tagged field). `KitSnippetConfig` passes it into `kit.SnippetConfig` so snippet functions can see it.

*Alternatives:* a top-level `open-schemes:` key — rejected, it is kit behavior and would be orphaned if the kit is disabled. Last-wins instead of concat — rejected: an allowlist is exactly the case where a project should extend, not silently drop, what the user allowed globally.

### The session hands the allowlist to the broker process via its environment

`broker.EnsureBroker` gains an `env []string` parameter appended to the spawned process's environment; the session (`cmd/asylum/main.go`) builds `ASYLUM_OPEN_SCHEMES=<comma list>` from the merged config and passes it. `openHandler` reads and parses that variable when validating a request.

The broker package stays kit-agnostic (it forwards opaque env entries); the variable name, its parsing, and its meaning live in `internal/kit/browser_open.go` next to the handler that consumes it.

*Alternatives:*
- Argv flag (`--open-schemes`) — works, but the env channel is already established for broker parameters and keeps `runBroker`'s hand-rolled arg loop unchanged.
- Teach the broker process to load the config itself — it would have to rediscover the project directory from the container name (a one-way hash), so it would need a new persisted mapping. Much more machinery for the same result.
- A `BrokerEnvFunc` field on `Kit`, collected generically — the cleaner shape if a second kit ever needs it, but speculative now.

### Normalization and validation happen in the session, not in the broker

`internal/kit/browser_open.go` exposes a normalizer that lowercases entries, strips a trailing `:`, `://`, or `:///`, and drops anything that is not a valid RFC 3986 scheme (`^[a-z][a-z0-9+.-]*$`). The session normalizes before building the env value and warns — with the project's `log` package — about each rejected entry, so the user sees the problem in the session where they can fix it. The handler applies the same filter to whatever it parses, so a hand-started broker cannot be widened by a malformed value.

Accepting the `dropshare5:///` form matters: that is how such schemes appear in tool documentation, and silently rejecting it would look like the feature is broken.

### `http`/`https` keep the host requirement; allowlisted schemes do not

`validBrowserURL` becomes `allowedURL(raw, schemes)`. For `http`/`https` the existing "must have a host" check stays (it is what rejects `http://` and bare text). Custom application schemes routinely use `scheme:///path`, which has no host, so for allowlisted schemes the check is "parses as a URL and the scheme matches". A URL with no scheme at all is still rejected, and the URL is still handed to the opener as a single argument with no shell, so metacharacters remain inert.

### The rules snippet becomes config-driven

`RulesSnippet` becomes `RulesSnippetFunc(*SnippetConfig)`, appending a sentence naming the configured schemes when there are any. It must tolerate a `nil` `SnippetConfig` — `browser-open` is always-on and is usually absent from the config entirely. The rules file is generated at container creation, so no image rebuild is involved.

## Risks / Trade-offs

- **A running broker keeps the allowlist it started with** → `ensureBroker` is a no-op while a live broker answers, so editing `schemes` mid-session has no effect until the container is restarted. This matches how other container-baked settings behave and is stated in both the spec and the kit docs.
- **Allowlisting a scheme is a real capability grant** → any process in the container can then launch the host app registered to that scheme with an arbitrary payload (`file:` would expose host files to host apps). Mitigated by being opt-in, per-scheme, and never defaulted; `docs/concepts/security-model.md` spells out what the grant means.
- **`.asylum` lives in the mounted project directory, so the agent can edit its own allowlist** → true, and already true for `build:` steps that run arbitrary Dockerfile commands. The trust model treats project config as user-owned; the change adds no new boundary. Worth one sentence in the security doc, not a new mechanism.
- **Cross-platform opener behavior varies** → macOS `open` resolves any registered scheme; Linux `xdg-open` resolves schemes via desktop-entry handlers and may fail if none is registered. The handler surfaces the opener's error as it does today.
