# Proposal

## Why

Asylum creates named Docker volumes per project (one per package cache, one per `node_modules` directory) but only removes them when the user runs `asylum cleanup` inside that project, or wipes everything with `cleanup --all`. Volumes of deleted or moved projects and of vanished `node_modules` directories pile up forever: a real installation with about 20 projects had over 120 `asylum-*` volumes and no way to tell which of them were still needed.

## What Changes

- Every named volume asylum creates carries labels identifying the project path it belongs to and what it holds (`cache` or `node_modules`). Labels are applied when Docker creates the volume. Existing volumes keep working unlabeled.
- New `asylum cleanup --orphans`: lists volumes that no longer belong to anything and removes them after confirmation. A volume is orphaned when its project directory no longer exists, or when it shadows a `node_modules` directory whose `package.json` is gone. For a project directory that no longer exists, the project's stopped container, project data directory and port allocation are removed too, the same as `asylum cleanup` would have done.
- When asylum starts a fresh container for a project, it removes that project's `node_modules` volumes whose `package.json` no longer exists. This needs no confirmation because those volumes are invisible to the project and are rebuilt by `npm install` anyway.

## Capabilities

### New Capabilities
- `volume-labels`: asylum-created named volumes carry labels that identify their owning project and kind.

### Modified Capabilities
- `cleanup-command`: adds `--orphans` to remove volumes (and leftovers) of projects or `node_modules` directories that no longer exist.
- `shadow-node-modules`: stale `node_modules` volumes of the current project are removed when a fresh container starts.

## Impact

- `internal/container/container.go`: volume mounts gain `volume-label` options; helper to compute a project's expected `node_modules` volume names.
- `internal/docker/docker.go`: list volumes with labels, find containers that reference volumes.
- `internal/ports`: reverse lookup of project path by container name (for unlabeled volumes) and removal of a project's allocation.
- `cmd/asylum/main.go`: `cleanup --orphans` parsing and command; pruning after fresh container start.
- Docs: `docs/commands/cleanup.md`, `docs/concepts/mounts.md`; `CHANGELOG.md`.
