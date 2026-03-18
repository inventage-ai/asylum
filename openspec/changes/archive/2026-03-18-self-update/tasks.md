## 1. Config: Add release-channel field

- [x] 1.1 Add `ReleaseChannel string` field to `Config` struct with `yaml:"release-channel"` tag
- [x] 1.2 Add scalar merge for `ReleaseChannel` in `merge()` (same as `Agent` — last wins)
- [x] 1.3 Add test cases for `release-channel` merging in `config_test.go`

## 2. Self-update package

- [x] 2.1 Create `internal/selfupdate/selfupdate.go` with `Run(currentVersion, channel, execPath string) error`
- [x] 2.2 Implement GitHub API client: resolve latest stable release (`releases/latest`) and dev release (`releases/tags/dev`)
- [x] 2.3 Implement platform detection (`runtime.GOOS`, `runtime.GOARCH`) and asset matching
- [x] 2.4 Resolve real binary path via `os.Executable()` + `filepath.EvalSymlinks()` (follows symlinks from `/usr/local/bin`)
- [x] 2.5 Implement binary download to temp file in same directory as resolved binary
- [x] 2.6 Implement atomic rename replacement with permission error handling (suggest sudo)
- [x] 2.7 Implement skip-when-current logic (compare version tags; always update for dev)
- [x] 2.8 Add unit tests for version comparison, asset name construction, and channel resolution

## 3. CLI integration

- [x] 3.1 Add `self-update` subcommand parsing in `parseArgs()` with `--dev` flag support
- [x] 3.2 Add `self-update` dispatch in `main()` — resolve channel from flag/config, call `selfupdate.Run()`
- [x] 3.3 Add `self-update` to `printUsage()` help text
- [x] 3.4 Add test cases for `self-update` and `self-update --dev` argument parsing

## 4. Install script: `~/.asylum/bin/` with symlink

- [x] 4.1 Change default `INSTALL_DIR` from `/usr/local/bin` to `~/.asylum/bin`, create directory if needed
- [x] 4.2 Download binary to `~/.asylum/bin/asylum`
- [x] 4.3 Create symlink `/usr/local/bin/asylum` → `~/.asylum/bin/asylum` (use `sudo` if needed)
- [x] 4.4 Handle case where `/usr/local/bin/asylum` already exists as a regular file (legacy install) — replace with symlink

## 5. Dev release CI workflow

- [x] 5.1 Create `.github/workflows/dev-release.yml` triggered on push to `main`
- [x] 5.2 Add step to delete existing `dev` tag and release via `gh release delete` + `git push --delete`
- [x] 5.3 Add build step with `VERSION=dev make build-all`
- [x] 5.4 Add step to create pre-release with `gh release create dev --prerelease` and attach binaries

## 6. Verification

- [x] 6.1 Run `go test ./...` and `go vet ./...` to confirm all tests pass
- [ ] 6.2 Manual smoke test: build binary, run `asylum self-update`, verify it queries GitHub and reports version
