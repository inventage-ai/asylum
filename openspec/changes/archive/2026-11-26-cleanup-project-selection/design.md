# Design

## Context

- The session that started a container removes it when the last session ends (`cmd/asylum/main.go`, after `runDocker`). A stopped container only remains after a crash or a killed session. Container state therefore cannot tell when a project was last used.
- Project images are tagged `asylum:proj-<sha256(dockerfile)[:12]>` (`image.EnsureProject`). Projects with identical project configuration share a tag, and a project without project-level snippets runs directly on `asylum:latest`. Nothing maps an image back to a project.
- Per-project data lives in `~/.asylum/projects/<container>/`. It already holds `onboarding.json`, history, credentials, agent configs and the sandbox rules. The directory name is the container name, not the project path.
- The port registry maps container names to project paths for every project started since the `ports` kit exists. The `volume-lifecycle` change adds the `asylum.project` label to volumes and the orphan classification in `internal/container/orphans.go`.
- `runCleanupProject` removes the container only if it is running, never releases the port allocation and does not touch images.

## Goals / Non-Goals

**Goals:**
- One overview of all asylum projects with status, last use and size, and an interactive way to clean up several at once.
- One per-project cleanup routine used by `cleanup`, `cleanup --select` and the deleted-project part of `cleanup --orphans`.
- Project images get removed with the last project that uses them, without guessing.

**Non-Goals:**
- A separate listing command (`asylum projects`) or non-interactive selection by flags (`--older-than`). Can be added later on top of the inventory.
- Configurable staleness threshold. 30 days is a constant.
- Removing project images of older asylum versions that are recorded by no project. `cleanup --all` covers them.
- Pruning the base image or build cache.

## Decisions

### Usage record: `~/.asylum/projects/<container>/project.json`
```json
{"project": "/Volumes/Code/app", "lastUsed": "2026-11-25T10:00:00Z", "image": "asylum:proj-1a2b3c4d5e6f"}
```
Written on every session start (fresh container and attach) after `ensureImages`, with the tag `ensureImages` returned. On attach this is the tag a fresh start would use; it differs from the running container's image only while a stale container is kept, which is short-lived. Write errors are warnings.

A file inside the data directory, rather than a new field in the port registry, because it is removed together with the data directory by every existing cleanup path, needs no lock, and the project path in it also resolves data directories that have no port allocation.

Alternative: label the container with the project path and read `FinishedAt`. Rejected: the container is removed after each session, so the information is gone exactly when it is needed.

### Inventory assembly as pure function
A new `container.Inventory(InventoryInput) []Project` gets everything as plain data, like `ClassifyOrphans`:
- usage records (container name → record), port registry, volume labels (volume → project), container list with running state and image, data directory names with mtime, sizes, and a `ProjectExists` function.

It groups by container name (volume owner via the existing `volumeOwner`, old-format names via `OldContainerName` mapping as in `ClassifyOrphans`), resolves the path (record → registry → volume label), and computes status, last use and size per the spec. `main.go` only gathers the inputs and renders. This keeps the logic table-testable without Docker.

### Sizes from one `docker system df -v --format '{{json .}}'`
Gives sizes of all volumes and images in one call. Sizes come as human strings (`1.2GB`); parse them into bytes with a small parser (`kB`, `MB`, `GB`, `TB`, decimal, and `B`). Data directory size via `filepath.WalkDir`. A failing call (e.g. Podman without `-v` support) yields unknown sizes; the listing continues.

`docker system df -v` can take several seconds with many volumes. A log line (`collecting project sizes...`) is printed before the call.

### Output and selection
```
  PROJECT                          STATUS    LAST USED   SIZE
  /Volumes/Code/inventage/asylum   running   now         4.1 GB
  /Users/me/code/shop              idle      3 days      2.3 GB
  /Users/me/code/old-poc           idle      5 months    1.8 GB
  /Users/me/code/gone              missing   8 months    900 MB
  asylum-1a2b3c4d5e6f              idle      unknown     120 MB
```
Sorted by last use, oldest first, unknown last. Then `tui.MultiSelect` over non-running projects, label = path, description = `status · last used · size`. Then a summary and `Proceed? (y/N)` (same pattern as `--all` and `--orphans`).

### Shared per-project cleanup
`cleanupProject(cname, projectDir string) error`-style routine in `main.go`:
1. `docker rm -f` the container (running or stopped). For `--select` running projects are never passed in; for `asylum cleanup` in the project the current behavior of removing a running container is kept.
2. Remove volumes with prefix `<cname>-` and `<oldName>-` (old name only if the path is known).
3. Read the usage record's image, remove the data directory, release the port allocation (if path known).
4. Remove the image if it starts with `asylum:proj-`, no remaining usage record names it and no container uses it. `docker rmi` without `-f`, so a race with a starting container fails safely.

`runCleanupOrphans` calls steps 3–4 for missing projects instead of its own inline removal. When several projects are cleaned in one run, the image check runs after all of them, so two selected projects sharing an image free it.

### Flag parsing
`--select` joins `--all` and `--orphans` in the existing cleanup flag switch; any combination of two of them is an error.

## Risks / Trade-offs

- [Projects last used before this change have no usage record] → Last use falls back to the data directory mtime, which is often older than the real last use, so such projects may be preselected too eagerly. The overview shows the date before anything is removed, and the first start after upgrading writes a correct record.
- [Shared images are never counted in project sizes] → The total freed space can be larger than shown. Acceptable; under-reporting is safer than double counting.
- [External drive not mounted shows projects as `missing`] → Same as `--orphans`: the list is shown before confirmation. `missing` projects are preselected, so the user must deselect them.
- [`docker system df -v` is slow] → Only run by the interactive command, with a log line.
