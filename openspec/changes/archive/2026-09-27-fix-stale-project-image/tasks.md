# Tasks

## 1. Implementation

- [x] 1.1 Add a pure `projectImageCurrent` decision helper with a table-driven test covering matching hashes, a changed packages hash, a changed base hash, a missing base label, and `noCache`; verify `go test ./internal/image/...`
- [x] 1.2 In `EnsureProject`, read the base `asylum.hash`, use the helper, and write `asylum.base.hash` on build; remove the `baseRebuilt` parameter and update the call site in `cmd/asylum/main.go` and `image_test.go`; verify `go build ./...` and `go test ./...`
- [x] 1.3 Update the "Two-tier images" note in `CLAUDE.md` and add a **Fixed** entry to `CHANGELOG.md`

## 2. Verification

- [x] 2.1 On the Mac, start this project after the base image was rebuilt by another project, and verify Asylum rebuilds the project image and the new container has `asylum-dropshare`
