## 1. Extend CredentialMount

- [x] 1.1 Add `HostPath string` field to `CredentialMount` in `internal/kit/kit.go`
- [x] 1.2 Update credential mount handling in `internal/container/container.go` to bind-mount `HostPath` directly when set (skip writing to staging dir)
- [x] 1.3 Add unit test for `HostPath` mount in `internal/container/container_test.go`

## 2. GitHub Kit Credential Provider

- [x] 2.1 Add `CredentialFunc` and `CredentialLabel` to the GitHub kit in `internal/kit/github.go`
- [x] 2.2 Add unit test for the GitHub credential func (hosts.yml exists, missing, etc.)

## 3. Config and Docs

- [x] 3.1 Update GitHub kit `ConfigSnippet` to include `credentials` option (N/A — credentials are a generic KitConfig field, not kit-specific config)
- [x] 3.2 Update GitHub kit docs page and CHANGELOG
