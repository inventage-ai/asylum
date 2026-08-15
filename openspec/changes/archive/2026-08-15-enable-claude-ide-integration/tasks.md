## 1. Implementation

- [x] 1.1 Add `CLAUDE_CODE_IDE_HOST_OVERRIDE: "host.docker.internal"` to the map returned by `Claude.EnvVars()` in `internal/agent/claude.go`, with a short comment explaining that the container's own loopback has no IDE listening on it
- [x] 1.2 Add a test in `internal/agent` asserting `Claude{}.EnvVars()` contains the override key and value (no existing `EnvVars` test to extend)

## 2. Documentation

- [x] 2.1 Add an `## IDE Integration` section to `assets/asylum-reference.md` stating that `/ide` connects to a host IDE, and naming both preconditions: a Docker Desktop engine, and `shared` agent config isolation (the default) so the container reads the same `~/.claude/ide` lock files the host IDE writes
- [x] 2.2 Note in that section that the connection is manual via `/ide`, and that Claude's own `autoConnectIde` setting makes it automatic and persists in the mounted config directory
- [x] 2.3 Add an **Added** entry to the Unreleased section of `CHANGELOG.md`

## 3. Verification

- [x] 3.1 Run `go test ./...` and `go vet ./...`
- [x] 3.2 Rebuild and start a container with a host IDE open on the project, run `/ide`, and confirm the IDE appears and connects
- [x] 3.3 Confirm the override reaches the container: `docker exec <container> env | grep CLAUDE_CODE_IDE_HOST_OVERRIDE`
