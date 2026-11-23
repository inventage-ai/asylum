# Tasks

## 1. Volume labels

- [x] 1.1 Extract `NodeModulesVolumes(cname, projectDir)` from `coreVolumes` and use it for the npm mounts; verify existing `container` tests pass unchanged
- [x] 1.2 Add `volume-label=asylum.project=<dir>` and `volume-label=asylum.kind=cache|node_modules` to every named-volume `--mount`, CSV-quoting the label field when the path contains `,` or `"`; verify with table-driven tests in `container_test.go` (plain path, path with comma)
- [x] 1.3 Document the labels in `docs/concepts/mounts.md` and verify the page builds with `mkdocs build --strict` (or review rendered markdown if mkdocs is unavailable)

## 2. Prune stale node_modules volumes on start

- [x] 2.1 Add `docker.ListVolumes`-based helper that returns the stale `<cname>-npm-*` volumes given the expected set; verify with a unit test on the pure set-difference function
- [x] 2.2 Call pruning in `cmd/asylum/main.go` after `WaitReady` in the fresh-container branch, only when shadowing is enabled, logging removed volumes and warnings on failure; verify manually by deleting a `package.json` in a test project and restarting (volume disappears from `docker volume ls`)
- [x] 2.3 Mention pruning in `docs/concepts/mounts.md` (node_modules section)

## 3. `cleanup --orphans`

- [x] 3.1 Parse `--orphans` for `cleanup` (subcommand and `--cleanup` alias); reject combination with `--all`; update help text and the "only --all is supported" error; verify with parser tests in `main_test.go`
- [x] 3.2 Add `ports.Lookup(container)` (project by container name) and `ports.Release(project)` if not present; verify with unit tests against a temp registry file
- [x] 3.3 Add Docker helpers: list volumes with `asylum.project` label, list asylum containers with state and mounted volume names; verify parsing with unit tests on sample output
- [x] 3.4 Implement orphan classification as a pure function (volumes, labels, registry, project-exists check, expected npm volumes, container mounts → removable / in-use / unknown, grouped by project with reason); verify with table-driven tests covering every scenario in `specs/cleanup-command/spec.md`
- [x] 3.5 Implement `runCleanupOrphans`: terminal check, listing, confirmation, removal of stopped containers, volumes, project data dirs and port allocations for missing projects, per-item error reporting; verify manually against a Docker daemon with a deleted test project
- [x] 3.6 Document `--orphans` in `docs/commands/cleanup.md` (usage, example output, what is removed) and `docs/configuration/flags.md` if cleanup flags are listed there

## 4. Wrap-up

- [x] 4.1 Add **Added** entries to `CHANGELOG.md` (volume labels, `cleanup --orphans`, automatic pruning of stale node_modules volumes)
- [x] 4.2 Run `go vet ./...` and `go test ./...`; both pass
- [x] 4.3 `openspec validate volume-lifecycle --strict` passes

## Workflow follow-up

- Archive the change with `/opsx:archive` after review.
