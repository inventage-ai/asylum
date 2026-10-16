# Tasks

## 1. Config

- [x] 1.1 Add `AgentConfig.Seed` and `Config.AgentSeed` (default `full`); ignore `seed` from `.asylum`/`.asylum.yaml` with a warning in `Load`. Verify with `TestLoadSeedIgnoredInProjectLayer`.

## 2. Agents

- [x] 2.1 Add `SeedSettings()` to the `Agent` interface and define the allowlist for Claude, Codex, Gemini, Copilot and Pi (OpenCode and Echo: none). Verify with `go build ./...`.

## 3. Seeding

- [x] 3.1 Seed by mode in `EnsureAgentConfigAt`, make `EnsureAgentConfig` a wrapper, error on unknown modes, and pass the configured mode from `cmd/asylum/main.go`. Verify with `TestEnsureAgentConfigAtSeedModes` and `TestEnsureAgentConfigAtUnknownSeed`.

## 4. Documentation

- [x] 4.1 Document `seed` in `docs/concepts/isolation.md` and add a CHANGELOG entry. Verify by reading the rendered section.

## 5. Verification

- [x] 5.1 Run `go vet ./...`, `go test ./...` and `openspec validate agent-config-seed --strict`; verify all pass.
