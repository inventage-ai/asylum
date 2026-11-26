# Tasks

## 1. Usage record

- [x] 1.1 Add read/write of `~/.asylum/projects/<container>/project.json` (`project`, `lastUsed`, `image`); verify with a round-trip unit test against a temp home
- [x] 1.2 Write the record on every session start after `ensureImages` in `cmd/asylum/main.go` (fresh and attach), logging a warning on failure; verify manually: start a session, check the file content, attach a second session and see `lastUsed` change

## 2. Per-project cleanup fixes

- [x] 2.1 Extract a shared per-project cleanup routine from `runCleanupProject`: remove the container also when stopped, remove volumes (new and old prefix), remove the data directory, release the port allocation; verify `asylum cleanup` in a test project with a stopped container removes container, volumes, data and the `ports.json` entry
- [x] 2.2 Add `docker` helpers for the images used by containers and image removal without `-f`; add the "remove project image if no remaining record and no container uses it" step (only `asylum:proj-*` tags, checked after all projects of a run); verify the decision with a table-driven unit test (unused, shared with other record, used by container, unrecorded, base image)
- [x] 2.3 Use the shared routine's data/port/image steps for deleted projects in `runCleanupOrphans`; verify existing orphan tests still pass and a deleted project's recorded image is removed
- [x] 2.4 Update `docs/commands/cleanup.md` (project cleanup table: stopped container, port allocation, project image) and the `--orphans` section (project image); add CHANGELOG entries (Fixed: stopped container and port allocation; Added: project image removal)

## 3. Project inventory

- [x] 3.1 Add `docker` helper for `docker system df -v --format '{{json .}}'` returning volume and image sizes in bytes, with a parser for Docker's size strings; verify with unit tests on sample JSON and size strings (`0B`, `12.3kB`, `1.2GB`)
- [x] 3.2 Implement `container.Inventory` as a pure function (records, registry, volume labels, containers, data dirs with mtime, sizes, `ProjectExists` → projects with path or container name, status, last use, size); verify with table-driven tests covering every scenario in `specs/project-inventory/spec.md`

## 4. `cleanup --select`

- [x] 4.1 Parse `--select` for `cleanup` (subcommand and `--cleanup` alias); reject combination with `--all` or `--orphans`; update help text; verify with parser tests in `main_test.go`
- [x] 4.2 Implement the overview table (sorted oldest first, unknown last, human sizes and relative times) and the preselection rule (missing, or last use > 30 days); verify preselection with a unit test and the table with a test on rendered output
- [x] 4.3 Wire up `runCleanupSelect`: require a terminal, gather inputs, print table, `tui.MultiSelect` over non-running projects, summary with total size, `Proceed? (y/N)`, then run the shared cleanup per project, continuing after errors; verify manually with two idle test projects (one selected, one not) and a running one
- [x] 4.4 Document `--select` in `docs/commands/cleanup.md` and `docs/commands/index.md` with sample output; add a CHANGELOG entry under Added

## 5. Verification

- [x] 5.1 Run `go vet ./...` and `go test ./...`; all pass
- [x] 5.2 Run `openspec validate cleanup-project-selection --strict`; it passes

## Workflow follow-up

- Archive `volume-lifecycle` before this change, since this change's specs build on its orphan cleanup.
- Both changes ship in PR #54.
