## 1. Add runtime session detection

- [x] 1.1 Add `HasOtherSessions(containerName string) bool` to `internal/docker/docker.go` — runs `docker exec <container> ps -o pid,ppid --no-headers`, counts PPID=0 processes excluding PID 1, returns true if count > 1 (more than just the check itself)
- [x] 1.2 Add tests for `HasOtherSessions` parsing logic (table-driven, testing the output parsing with various `ps` outputs)

## 2. Replace counter with runtime check in CLI

- [x] 2.1 Remove `IncrementSessions`/`DecrementSessions` calls and `incremented` variable from `cmd/asylum/main.go`, replace cleanup condition with `if !docker.HasOtherSessions(cname)`
- [x] 2.2 Simplify rebuild prompt — remove `SessionCount` call, use plain "Container is running. Kill it and rebuild?" message
- [x] 2.3 Add `syscall.SIGHUP` to `signal.Notify` in `runDocker`

## 3. Remove counter code

- [x] 3.1 Remove `IncrementSessions`, `DecrementSessions`, `adjustCounter`, `SessionCount`, `sessionCounterPath` from `internal/container/container.go`
- [x] 3.2 Remove counter-related tests from `internal/container/` if any exist

## 4. Verify

- [x] 4.1 Run `go test ./...` and `go vet ./...`
- [x] 4.2 Build binary and manually test: clean session starts and stops container, second concurrent session keeps container alive
