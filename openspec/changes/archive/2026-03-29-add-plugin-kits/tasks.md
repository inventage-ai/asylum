## 1. ast-grep Kit

- [x] 1.1 Create `internal/kit/astgrep.go` with kit registration: name "ast-grep", TierDefault, Deps ["node"], DockerSnippet installing @ast-grep/cli via npm, Tools ["sg"], BannerLines, RulesSnippet, ConfigSnippet, and ConfigNodes
- [x] 1.2 Verify `go build` succeeds and kit appears in registry

## 2. Browser Kit

- [x] 2.1 Create `internal/kit/browser.go` with kit registration: name "browser", TierDefault, Deps ["node"], DockerSnippet installing playwright and chromium via `npx playwright install --with-deps chromium`, CacheDirs for playwright cache, Tools ["playwright"], BannerLines, RulesSnippet, ConfigSnippet, and ConfigNodes
- [x] 2.2 Verify `go build` succeeds and kit appears in registry

## 3. cx Kit

- [x] 3.1 Create `internal/kit/cx.go` with kit registration: name "cx", TierDefault, no Deps, DockerSnippet installing cx via install script, Tools ["cx"], BannerLines, RulesSnippet, ConfigSnippet (with commented-out packages/languages examples), and ConfigNodes
- [x] 3.2 Add `"cx": "cx-lang"` to `collectPackages` in `cmd/asylum/main.go`
- [x] 3.3 Add `"cx-lang"` to `knownPackageTypes` in `internal/image/image.go` and add a cx-lang block in `generateProjectDockerfile` that runs `cx lang add` for each configured language
- [x] 3.4 Verify `go build` succeeds and kit appears in registry

## 4. Verification

- [x] 4.1 Run `go test ./internal/kit/...` to ensure no regressions
- [x] 4.2 Run `go vet ./...` to check for issues
- [x] 4.3 Add CHANGELOG entry under Unreleased > Added for the three new kits
