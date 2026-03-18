## 1. Config loading: read .tool-versions

- [x] 1.1 Add `readToolVersionsJava(projectDir string) string` to `internal/config/config.go` that parses `.tool-versions` and returns the Java version (empty string if not found)
- [x] 1.2 Call it in `Load` and set `versions.java` only if not already set by any asylum config layer (before CLI flag overlay)

## 2. Tests

- [x] 2.1 Add unit tests: `.tool-versions` with java line, without java line, no file, asylum config overrides, CLI flag overrides
