## 1. RunArg type and pipeline

- [x] 1.1 Define `RunArg` struct (Flag, Value, Source, Priority) and priority constants in a new `internal/container/runarg.go`
- [x] 1.2 Implement dedup key extraction: classify by flag type (`-p`, `-v`, `--mount`, `-e`, boolean, `--cap-add`, single-value) and extract the key per the spec
- [x] 1.3 Implement `Resolve([]RunArg) ([]RunArg, []Override, error)` — group by key, apply priority-wins, detect same-priority conflicts, return resolved args + list of overrides for debug output
- [x] 1.4 Implement `Flatten([]RunArg) []string` — convert resolved RunArgs to the `[]string` that `docker run` expects
- [x] 1.5 Add tests for dedup key extraction, priority resolution, conflict detection, and flatten

## 2. Kit ContainerFunc

- [x] 2.1 Add `ContainerFunc func(ContainerOpts) ([]RunArg, error)` and `ContainerOpts` struct to `internal/kit/kit.go`
- [x] 2.2 Add `ContainerFunc` to the ports kit that calls `ports.Allocate()` and returns `-p` RunArgs
- [x] 2.3 Add `ContainerFunc` to the docker kit that returns `--privileged` and `-e ASYLUM_DOCKER=1` RunArgs
- [x] 2.4 Add a helper in `internal/kit/` that iterates resolved kits, calls each `ContainerFunc`, and collects all RunArgs (logging warnings on errors)

## 3. Rewrite container.RunArgs

- [x] 3.1 Refactor `RunArgs()` to produce core RunArgs (structural flags, project mount, gitconfig, caches, history, agent config, env vars, etc.) with source `core` and priority 0
- [x] 3.2 Convert user-config sources (config ports, config volumes, config env vars) to RunArgs with source `user config (...)` and priority 2
- [x] 3.3 Convert CredentialFunc/MountFunc output to RunArgs with source `<kit> kit (credentials)`/`<kit> kit (mounts)` and priority 1
- [x] 3.4 Collect kit ContainerFunc RunArgs (priority 1) via the helper from 2.4
- [x] 3.5 Feed all RunArgs into `Resolve()`, then `Flatten()`, replacing the old procedural assembly
- [x] 3.6 Remove `appendPorts()` and `appendEnvVars()` helper functions (absorbed into pipeline)
- [x] 3.7 Remove the `KitActive("docker")` checks in `RunArgs()` and `appendEnvVars()`

## 4. Cleanup deletions

- [x] 4.1 Remove `ports.Release()`, `ports.ReleaseContainer()`, and `release()` from `internal/ports/ports.go`
- [x] 4.2 Remove `ReleaseContainer` callsites in `cmd/asylum/main.go` (cleanup command ~line 703, prune ~line 824)
- [x] 4.3 Delete `internal/kit/title.go` and remove the `KitActive("title")` check in `container.go` `agentCommand()`
- [x] 4.4 Remove the `if cfg.KitActive("ports")` block in `cmd/asylum/main.go` (~line 254) and the `AllocatedPorts` field from `RunOpts`

## 5. Debug output

- [x] 5.1 Add `--debug` flag parsing in `cmd/asylum/main.go`
- [x] 5.2 Implement debug table printer that formats resolved RunArgs with source alignment and override list, writing to stderr
- [x] 5.3 Wire debug output into the container start path: print after `Resolve()`, before `docker.RunDetached()`

## 6. Testing and verification

- [x] 6.1 Update existing `container` package tests to work with the new RunArg-based `RunArgs()` signature
- [x] 6.2 Verify `go test ./...` passes
- [x] 6.3 Verify `go vet ./...` passes
