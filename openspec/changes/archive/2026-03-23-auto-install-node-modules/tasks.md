## 1. Lockfile detection and prompt

- [x] 1.1 Implement `nodeInstallCmds(projectDir)` — detect lockfiles via `FindNodeModulesDirs`, map to install commands
- [x] 1.2 Consolidated prompt listing all projects with relative paths and install commands
- [x] 1.3 Prompt uses `log.Info` with blank line for visibility after build output
- [x] 1.4 Only prompt in agent mode, gated on `!FeatureOff("auto-install-node-modules")`

## 2. Container exec integration

- [x] 2.1 Add `PreRunCmds []string` to `ExecOpts`
- [x] 2.2 Wrap agent command in `bash -c` with PATH/fnm setup when PreRunCmds non-empty
- [x] 2.3 Add `shellJoin` helper for safe command quoting
- [x] 2.4 Install failures are non-fatal (`;` before `exec agent`)

## 3. Shadow volume ownership

- [x] 3.1 Add `docker.Exec(container, user, command...)` helper
- [x] 3.2 After container start, chown each shadow volume to `claude:claude` as root
- [x] 3.3 Only runs on new container creation (inside `!docker.IsRunning` block)

## 4. Changelog

- [x] 4.1 Add entry under Unreleased

## 5. Verification

- [x] 5.1 Run `go test ./...` and `go vet ./...`
- [x] 5.2 Manual test: run asylum on Node.js project, verify prompt appears and deps install
