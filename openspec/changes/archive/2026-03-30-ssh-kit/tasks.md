## 1. Extend Credential Infrastructure

- [x] 1.1 Add `ContainerName` field to `CredentialOpts` in `internal/kit/kit.go`
- [x] 1.2 Pass container name when calling `CredentialFunc` in `internal/container/container.go`
- [x] 1.3 Update credential mode check to treat empty mode as `auto` for `TierAlwaysOn` kits

## 2. Config Support

- [x] 2.1 Add `SSHIsolation()` accessor to `internal/config/config.go` (returns isolation value, defaults to `"isolated"`)
- [x] 2.2 Pass SSH isolation to the credential function (add `Isolation` field to `CredentialOpts`)

## 3. Create SSH Kit

- [x] 3.1 Create `internal/kit/ssh.go` with `TierAlwaysOn` registration, `CredentialFunc` implementing isolated/shared/project modes, and `ConfigNodes` for the isolation option
- [x] 3.2 Add tests for the credential function (all three modes, key generation, known_hosts detection, idempotent re-runs)

## 4. Remove Old SSH Code

- [x] 4.1 Remove the hardcoded SSH volume mount block from `internal/container/container.go`
- [x] 4.2 Remove the `ssh-init` command dispatch from `cmd/asylum/main.go`
- [x] 4.3 Delete `internal/ssh/` package

## 5. Documentation and Changelog

- [x] 5.1 Update `assets/asylum-reference.md` to document SSH as a kit with isolation options
- [x] 5.2 Add changelog entry under Unreleased
