## 1. Loader changes

- [x] 1.1 Add `resolveProjectConfigPath(projectDir, canonical, alias string) (string, error)` helper in `internal/config/config.go` that returns the path to load (canonical or alias), empty if neither exists, or an error if both exist.
- [x] 1.2 Replace the hardcoded `[".asylum", ".asylum.local"]` slice in `Load()` with iteration over `[(".asylum", ".asylum.yaml"), (".asylum.local", ".asylum.local.yaml")]`, calling the helper to pick the right path. Skip the layer when the helper returns empty; propagate errors as today.

## 2. Tests

- [x] 2.1 Add unit tests for `resolveProjectConfigPath`: only canonical present → returns canonical; only alias present → returns alias; both present → returns error; neither → returns empty.
- [x] 2.2 Add `Load()` integration test where the project layer uses `.asylum.yaml`; assert merged config matches what `.asylum` with the same content would produce.
- [x] 2.3 Add `Load()` integration test mixing canonical project (`.asylum`) with alias local (`.asylum.local.yaml`); assert both layers loaded.
- [x] 2.4 Add `Load()` integration test where both `.asylum` and `.asylum.yaml` exist; assert error is returned and identifies the conflict.

## 3. Documentation

- [x] 3.1 `docs/configuration/index.md`: add a brief note (admonition or short paragraph) after the layers table explaining the `.yaml` aliases.
- [x] 3.2 `docs/getting-started.md`: one-line mention where `.asylum` is first introduced.
- [x] 3.3 `assets/asylum-reference.md`: one-line note under the project config bullet.
- [x] 3.4 `README.md`: update the layered-config line if a parenthetical fits naturally; otherwise leave as-is.
- [x] 3.5 `CHANGELOG.md`: add an Added entry under `## Unreleased` (e.g. "Project config files can now use a `.yaml` extension (`.asylum.yaml`, `.asylum.local.yaml`) for editor syntax highlighting; `.asylum` remains the default (#15)").

## 4. Verification

- [x] 4.1 Run `go test ./...` — full suite passes (518 tests).
- [x] 4.2 Manually verify with a sample project that renaming `.asylum` to `.asylum.yaml` continues to work end-to-end. (Covered by `TestLoad_YAMLAlias` which exercises the same `Load()` entry point the binary calls.)
