## 1. Config Struct and Loading

- [x] 1.1 Create `internal/config/config.go` with Config struct, CLIFlags struct, merge function, and Load function
- [x] 1.2 Implement volume shorthand parsing with tilde expansion

## 2. Tests

- [x] 2.1 Write table-driven tests for merge semantics (scalar, list, map-of-lists) and CLI flag overlay
- [x] 2.2 Write table-driven tests for volume shorthand parsing
- [x] 2.3 Verify `go build ./...` and `go test ./...` pass
