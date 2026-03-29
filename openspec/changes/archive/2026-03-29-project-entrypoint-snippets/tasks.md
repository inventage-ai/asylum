## 1. Project Entrypoint Assembly

- [x] 1.1 Add `assembleProjectEntrypoint` to `internal/image/image.go` — concatenate `EntrypointSnippet`s from project kits, append `PROJECT_BANNER` export from `BannerLines`, return nil when empty
- [x] 1.2 Update `generateProjectDockerfile` to accept project entrypoint content and emit a `COPY` instruction for `project-entrypoint.sh` when non-nil
- [x] 1.3 Update `EnsureProject` to call `assembleProjectEntrypoint`, pass result to `generateProjectDockerfile`, and include the script in `extraFiles` for `buildImage`

## 2. Base Entrypoint Integration

- [x] 2.1 Add `source /usr/local/bin/project-entrypoint.sh` (guarded by existence check, with `|| true`) to `assets/entrypoint.tail` before the welcome banner block
- [x] 2.2 Add `eval "$PROJECT_BANNER"` (guarded by variable check) inside the banner block in `entrypoint.tail`

## 3. Project Image Trigger Fix

- [x] 3.1 Update the early-return check in `EnsureProject` to also consider whether project kits have `EntrypointSnippet`s or `BannerLines` (currently only checks `packages`, `javaVersion`, and `profileSnippets`)

## 4. Tests

- [x] 4.1 Add unit test for `assembleProjectEntrypoint` — kits with snippets, kits without, mixed, banner lines
- [x] 4.2 Add unit test for `generateProjectDockerfile` with project entrypoint content (verify COPY instruction present/absent)
- [x] 4.3 Verify `go build ./...` and `go test ./...` pass
