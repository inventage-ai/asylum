# Proposal

## Why

`asylum cleanup --orphans` only catches projects whose directory is gone. Projects that still exist on disk but have not been used for months keep their volumes, data and project image forever, and the only ways to remove them are `asylum cleanup` from inside each project or `cleanup --all`. There is no overview of which asylum projects exist, when they were last used and how much space they take, so users cannot make that decision.

Two gaps make this worse: project images (`asylum:proj-<hash>`) are never removed by `cleanup` or `--orphans`, and the per-project `asylum cleanup` leaves a stopped container (and with it the volumes it references) and the port allocation behind.

## What Changes

- Each project start records the project path, the time of the start and the image the container runs on in a small state file in the project data directory (`~/.asylum/projects/<container>/`).
- New `asylum cleanup --select`: prints all known asylum projects with status (`running`, `stopped`, `idle`, `missing`), last use and disk size, then lets the user pick projects in a multi-select. Missing projects and projects unused for more than 30 days are preselected; running projects are listed but cannot be selected. After a final confirmation each selected project is cleaned up the same way as `asylum cleanup` in that project.
- Cleaning up a project also removes its project image when no other known project and no container uses it. This applies to `asylum cleanup`, `cleanup --select` and to deleted projects in `cleanup --orphans`. Images not recorded by any project (from older versions) are left alone; `cleanup --all` still removes them.
- `asylum cleanup` removes the project's container also when it is stopped, and releases the project's port allocation, as the documentation already states.

## Capabilities

### New Capabilities
- `project-inventory`: asylum records per-project usage (path, last start, image) and can enumerate all known projects with status, last use and size.

### Modified Capabilities
- `cleanup-command`: adds `--select` for interactive multi-project cleanup; per-project cleanup also removes stopped containers, the port allocation and the project's no longer used project image.

## Impact

- `cmd/asylum/main.go`: `--select` parsing, inventory listing and selection, shared per-project cleanup function used by `cleanup`, `--select` and `--orphans`; writing the state file on start.
- `internal/container` (or a new small package): project inventory assembly from port registry, volume labels, project data directories, containers and state files; pure classification logic for tests.
- `internal/docker/docker.go`: volume and image sizes (`docker system df -v`), container image lookup.
- Docs: `docs/commands/cleanup.md`, `docs/commands/index.md`; `CHANGELOG.md`.
- Lands in PR #54 next to the `volume-lifecycle` change; that change is archived first because this one builds on its orphan cleanup.
