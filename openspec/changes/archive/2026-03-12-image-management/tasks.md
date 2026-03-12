## 1. Assets

- [x] 1.1 Create `assets/assets.go` with go:embed for Dockerfile and entrypoint.sh
- [x] 1.2 Create placeholder `assets/Dockerfile` and `assets/entrypoint.sh`

## 2. Image Package

- [x] 2.1 Create `internal/image/image.go` with EnsureBase, EnsureProject, hash computation, and project Dockerfile generation
- [x] 2.2 Verify `go build ./...` passes
