## 1. Config surface

- [x] 1.1 Add `Schemes []string \`yaml:"schemes,omitempty" merge:"concat"\`` to `config.KitConfig` in `internal/config/config.go`
- [x] 1.2 Add `Schemes` to `kit.SnippetConfig` in `internal/kit/kit.go` and pass it through `config.KitSnippetConfig`
- [x] 1.3 Add a merge test in `internal/config/config_test.go` covering `schemes` accumulating across layers

## 2. Scheme normalization

- [x] 2.1 In `internal/kit/browser_open.go`, add a normalizer that lowercases entries, strips a trailing `:`, `://`, or `:///`, drops duplicates and `http`/`https`, and separates valid schemes (`^[a-z][a-z0-9+.-]*$`) from rejected entries
- [x] 2.2 Table-driven test in `internal/kit/browser_open_test.go` for the normalizer: bare name, `scheme:`, `scheme://`, `scheme:///`, mixed case, duplicates, invalid entries (`1foo`, `a b`, empty)

## 3. Host-side validation

- [x] 3.1 Replace `validBrowserURL` with an allowlist-aware check: `http`/`https` still require a host; allowlisted schemes only need to parse with a matching scheme; no-scheme and unparseable input rejected
- [x] 3.2 Have `openHandler` read the allowlist from the broker process environment (`ASYLUM_OPEN_SCHEMES`), re-filtering it through the normalizer, and reject disallowed URLs with `400` before invoking the host opener
- [x] 3.3 Extend the `validBrowserURL` test to cover the allowlist cases: allowlisted `dropshare5:///upload?path=/tmp/x.png` accepted, `file:///etc/passwd` rejected when not listed and accepted when listed, hostless `http://` rejected, empty allowlist behaving exactly as today

## 4. Broker plumbing

- [x] 4.1 Add an `env []string` parameter to `broker.EnsureBroker` in `internal/broker/broker.go`, appended to the spawned process's environment; keep the package kit-agnostic
- [x] 4.2 In `cmd/asylum/main.go`, pass the merged config through `ensureBroker` and build the `ASYLUM_OPEN_SCHEMES` value from `kits.browser-open.schemes`, warning via the `log` package about entries the normalizer rejected
- [x] 4.3 Verify a broker started without the variable behaves exactly as before (defaults to `http`/`https`)

## 5. Agent-facing rules

- [x] 5.1 Convert the kit's `RulesSnippet` to `RulesSnippetFunc`, naming configured schemes in the snippet and handling a `nil` `SnippetConfig` (always-on kit, usually absent from config)
- [x] 5.2 Test both snippet forms: default `http(s)` wording with no config, and configured schemes named when present

## 6. Docs and changelog

- [x] 6.1 Document `kits.browser-open.schemes` in `docs/kits/browser-open.md`: syntax, accepted entry forms, layer accumulation, the restart-to-apply behavior, and a Dropshare example
- [x] 6.2 Update `docs/concepts/security-model.md` to state what allowlisting a scheme grants and that it is opt-in and per-scheme
- [x] 6.3 Add a CHANGELOG entry under **Unreleased** → **Added**

## 7. Verification

- [x] 7.1 `go test ./...` and `go vet ./...` pass
- [x] 7.2 Manual check: with `schemes: [dropshare5]` configured, `open "dropshare5:///…"` inside the container reaches the host app; without it, the same URL is rejected and `open https://example.com` still works
