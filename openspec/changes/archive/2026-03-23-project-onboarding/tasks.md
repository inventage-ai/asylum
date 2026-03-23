## 1. Onboarding framework

- [x] 1.1 Create `internal/onboarding/` package with `Task` interface and `Workload` struct
- [x] 1.2 Implement `Run(opts)` orchestrator: load state → detect → prompt → execute → save state
- [x] 1.3 Implement state persistence: load/save `~/.asylum/projects/<cname>/onboarding.json`
- [x] 1.4 Implement input hashing: SHA-256 of file contents for change detection
- [x] 1.5 Implement consolidated prompt: list pending workloads, single Y/n confirmation
- [x] 1.6 Implement workload execution via `docker exec` with streamed output
- [x] 1.7 Add tests for state load/save, hash comparison, and orchestrator logic

## 2. PATH resolution for docker exec

- [x] 2.1 Add `echo "$PATH" > /tmp/asylum-path` at end of entrypoint setup (before sleep)
- [x] 2.2 Implement `docker.ReadFile(container, path)` helper to read `/tmp/asylum-path`
- [x] 2.3 Pass `-e PATH=<resolved>` on onboarding `docker exec` calls

## 3. npm onboarding task

- [x] 3.1 Implement npm task struct: `Detect` using `FindNodeModulesDirs` + lockfile check
- [x] 3.2 Map lockfiles to install commands (npm ci, pnpm, yarn, bun)
- [x] 3.3 Use lockfile path as hash input
- [x] 3.4 Per-task config: `onboarding: { npm: false }` (checked in `Run()` via `opts.Onboarding`)
- [x] 3.5 Add tests for detection with various lockfile combinations

## 4. CLI integration

- [x] 4.1 Add `--skip-onboarding` flag to `parseArgs` and `cliFlags`
- [x] 4.2 Call `onboarding.Run()` in `main()` between container start and session exec, agent mode only
- [x] 4.3 Add `Onboarding map[string]bool` field to Config struct with map merge semantics
- [x] 4.4 Check global `features: { onboarding: false }` and per-task `onboarding: { <name>: false }` before running
- [x] 4.5 Add `--skip-onboarding` to `printUsage()` help text
- [x] 4.6 Add test cases for `--skip-onboarding` flag parsing and onboarding config merging

## 5. Remove old auto-install plumbing

- [x] 5.1 Remove `PreRunCmds` from `ExecOpts` and bash wrapper from `ExecArgs`
- [x] 5.2 Remove `shellJoin` helper
- [x] 5.3 Remove `nodeInstallCmds` / `promptNodeInstall` from `main.go`
- [x] 5.4 Remove Node.js install block from entrypoint.sh (if still present)
- [x] 5.5 Update changelog

## 6. Verification

- [x] 6.1 Run `go test ./...` and `go vet ./...`
- [x] 6.2 Manual test: first run prompts, installs, records state
- [x] 6.3 Manual test: second run skips (same lockfile hash)
- [x] 6.4 Manual test: `--skip-onboarding` suppresses prompt
- [x] 6.5 Manual test: `onboarding: { npm: false }` disables npm task
