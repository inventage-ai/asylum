# Tasks

## 1. Config

- [x] 1.1 Add `EnvFromHost map[string]string` (`env-from-host`) to `Config` and merge it per key in `Merge`; verify with a `Merge` test case.
- [x] 1.2 In `Load`, clear `EnvFromHost` from the project layer (`.asylum` / `.asylum.yaml`) with a warning; verify with `Load` tests for global, project, and local layers.

## 2. Container assembly

- [x] 2.1 Emit `-e NAME=value` RunArgs (source `user config (env-from-host)`) for set, non-empty host variables, skipping names present in `env`, the reserved agent variable, and failing on newlines; verify with table-driven tests in `internal/container`.

## 3. Documentation

- [x] 3.1 Document `env-from-host` in `docs/configuration/index.md` and `assets/asylum-reference.md`, including the layer restriction, precedence, and that values are fixed at container creation; add a CHANGELOG entry. Verify by reading the rendered sections.

## 4. Verification

- [x] 4.1 Run `go vet ./...` and `go test ./...` and verify both pass.
