## 1. Kit Definition

- [x] 1.1 Create `internal/kit/rtk.go` with kit registration: name `rtk`, tier `TierOptIn`, no deps, tool `rtk`, `NeedsMount: true`, DockerSnippet (curl install + `rtk init -g` + save artifacts to `/tmp/asylum-kit-rtk/`), EntrypointSnippet (mount hooks dir and RTK.md into `~/.claude/`, register hook in settings.json), RulesSnippet, BannerLines, ConfigSnippet/ConfigNodes/ConfigComment
- [x] 1.2 Verify kit registration works via `go build ./...`

## 2. Documentation

- [x] 2.1 Create `docs/kits/rtk.md` with activation, what's included, and usage examples
- [x] 2.2 Add changelog entry under Unreleased > Added

## 3. Validation

- [x] 3.1 Run `go test ./internal/kit/...` to verify all tests pass
