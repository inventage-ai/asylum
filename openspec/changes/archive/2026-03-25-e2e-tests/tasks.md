## 1. Echo Agent

- [x] 1.1 Create `internal/agent/echo.go` with Echo agent: Name "echo", Binary "echo", no config dirs, no session detection, Command returns `["echo", args...]` (no zsh wrapper needed since echo is a basic command)
- [x] 1.2 Register Echo in the agents map
- [x] 1.3 Add unit test for Echo.Command with and without args

## 2. E2e Test Framework

- [x] 2.1 Create `e2e/` directory with `e2e` build tag
- [x] 2.2 Create `e2e/e2e_test.go` with TestMain: check Docker available, build asylum binary to temp dir, set up temp HOME and project dir with minimal config, cleanup images/containers after all tests
- [x] 2.3 Add helper function `runAsylum(t, args...)` that executes the built binary with the test HOME and project dir, returns stdout/stderr/exit code
- [x] 2.4 Add helper function `runAsylumSuccess(t, args...)` that calls runAsylum and fails test if exit code != 0

## 3. Basic Tests

- [x] 3.1 Test: `asylum --help` exits 0 and stdout contains "Usage:"
- [x] 3.2 Test: `asylum --version` exits 0 and stdout contains version string
- [x] 3.3 Test: `asylum run echo ok` builds image, starts container, outputs "ok", cleans up container

## 4. Lifecycle Tests

- [x] 4.1 Test: `asylum -a echo -- hello world` starts container, runs echo agent, outputs "hello world", cleans up
- [x] 4.2 Test: second `asylum run echo again` reuses cached image (faster than first run)
- [x] 4.3 Test: after asylum exits, no asylum container is left running for that project

## 5. Wiring

- [x] 5.1 Add `test-e2e` target to Makefile: `go test -tags e2e -v -timeout 30m ./e2e/`
- [x] 5.2 Add CHANGELOG entry if appropriate (probably not — test infrastructure isn't user-facing)
