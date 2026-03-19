## 1. Config: FeatureOff method

- [x] 1.1 Add `FeatureOff(name string) bool` method to `Config` — returns true only when feature is explicitly false

## 2. Node modules detection

- [x] 2.1 Implement `findNodeModulesDirs(projectDir)` — find `package.json` files and return `node_modules` paths next to them
- [x] 2.2 Return paths even when `node_modules` does not exist yet (proactive shadow for fresh clones)
- [x] 2.3 Skip `.git`, `.venv`, `vendor`, `target`, `dist` during walk
- [x] 2.4 Skip `node_modules` directories during walk (don't shadow packages inside node_modules)
- [x] 2.5 Sort results for deterministic output
- [x] 2.6 Add tests: root, subdirectory, monorepo, no package.json, no node_modules, heavy dir skip, node_modules recursion

## 3. Volume shadow integration

- [x] 3.1 Add shadow mount logic to `appendVolumes` gated on `!FeatureOff("shadow-node-modules")`
- [x] 3.2 Name volumes as `<container-name>-npm-<sha256(rel_path)[:11]>`
- [x] 3.3 Use `--mount type=volume,src=<name>,dst=<path>` syntax
- [x] 3.4 Add test: shadow mount present with correct volume name
- [x] 3.5 Add test: shadow mount absent when feature disabled

## 4. Volume cleanup

- [x] 4.1 Add `ListVolumes(prefix)` and `RemoveVolumes(...)` to `internal/docker/docker.go`
- [x] 4.2 Add volume removal to `runCleanup` in `main.go` — remove all `asylum-` prefixed volumes
- [x] 4.3 Update cleanup log message to mention volumes

## 6. Changelog and documentation

- [x] 6.1 Add changelog entry under Unreleased

## 7. Verification

- [x] 7.1 Run `go test ./...` and `go vet ./...`
- [x] 7.2 Manual test: run asylum on a Node.js project, verify `docker volume ls` shows the named volume
