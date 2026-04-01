## 1. Remove subcommand from RunArg pipeline

- [x] 1.1 Remove `core("run", "")` and `core("-d", "")` from `RunArgs()` in `internal/container/container.go`
- [x] 1.2 Prepend `"run", "-d"` to the `flat` slice before appending image and command in `RunArgs()`
- [x] 1.3 Remove `-d` from `booleanFlags` map in `internal/container/runarg.go`

## 2. Update tests

- [x] 2.1 Update `internal/container/runarg_test.go` — remove any tests expecting `run`/`-d` as RunArgs in resolved output
- [x] 2.2 Update `internal/container/container_test.go` — adjust tests that check flattened args to expect `run -d` as first two elements
- [x] 2.3 Run `go test ./internal/container/` and fix any remaining failures
