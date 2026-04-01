## 1. Kit struct

- [x] 1.1 Add `Hidden bool` field to the `Kit` struct in `internal/kit/kit.go`
- [x] 1.2 Set `Hidden: true` on the `apt` kit in `internal/kit/apt.go`

## 2. Filter UI surfaces

- [x] 2.1 Skip hidden kits in the config TUI Kits tab (`cmd/asylum/config.go`)
- [x] 2.2 Skip hidden kits in the new-kit sync prompt (`internal/config/kitsync.go`) — silently add as comment instead
- [x] 2.3 Skip hidden kits in the disabled-kits section of sandbox rules (`internal/container/container.go`)

## 3. Verify

- [x] 3.1 Run `go build ./...` and `go test ./...`
