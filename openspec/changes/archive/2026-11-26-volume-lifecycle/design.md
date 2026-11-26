# Design

## Context

Volumes are created implicitly by `docker run --mount type=volume,src=<name>,...` in `coreVolumes` (`internal/container/container.go`). Names are `<container>-cache-<tool>` and `<container>-npm-<sha256(relpath)[:11]>`, where `<container>` is `asylum-<sha256(projectDir)[:6]>-<basename>`. The name carries the project basename but not the full path, so it cannot tell whether the project still exists.

Containers are not started with `--rm`: a stopped container stays until the next start of that project (which removes it before `docker run`). `docker volume rm` refuses volumes referenced by any container, stopped or running.

The port registry (`~/.asylum/ports.json`, `ports` kit is always on) stores `{project, container}` for every project that has ever been started, which gives a project path for volumes that predate labels.

## Goals / Non-Goals

**Goals:**
- Volumes created from now on are self-describing.
- Volumes that already exist can be classified for most users without labels.
- No data of an existing, still-referenced `node_modules` or cache is ever removed.

**Non-Goals:**
- Sharing caches across projects or avoiding creation of cache volumes for kits a project does not use. Many observed cache volumes are empty because global kits apply to every project; that is a separate change.
- Relabeling or migrating existing volumes.
- Reporting volume sizes.

## Decisions

### Labels via `volume-label` mount option
`--mount type=volume,src=X,dst=Y,volume-label=asylum.project=<dir>,volume-label=asylum.kind=cache` sets the labels when Docker creates the volume and is ignored for an existing volume. No separate `docker volume create` call is needed.

The `--mount` value is parsed by Docker as CSV, so a project path containing a comma or quote must be CSV-quoted as a field (`"volume-label=asylum.project=/a,b"`). The mount builder quotes the label field only when needed.

Alternative: `docker volume create --label ...` before `docker run`. Rejected: an extra Docker call per volume on every start for no gain.

### Expected node_modules volume names from one helper
Both the mount assembly and the pruning need "the npm volume names this project should have". Extract `NodeModulesVolumes(cname, projectDir) map[string]string` (volume name → mount path) from `coreVolumes` and reuse it. Stale = volumes with prefix `<cname>-npm-` not in that map. This works for unlabeled volumes too, because it compares names only.

### Pruning on start runs after the fresh container is up
The old stopped container is removed right before `docker run`, and the new container only references currently expected volumes. Pruning after `WaitReady` therefore never hits a self-inflicted "volume in use". It only runs in the fresh-container branch and when shadowing is enabled; errors are logged as warnings.

Secondary containers (`SecondaryContainerName`) have their own name prefix and prune their own volumes the same way.

### Orphan detection in `cleanup --orphans`
1. `docker volume ls --filter name=asylum- --format '{{.Name}}\t{{.Label "asylum.project"}}'`.
2. Owner container name: strip the `-cache-<tool>` or `-npm-<hash>` suffix (last occurrence).
3. Project path: label, else port registry lookup by container name, else "unknown" (listed, never removed).
4. Project missing → all its volumes orphaned. Project present → only stale npm volumes (step above) are orphaned.
5. `docker ps -a --filter name=asylum- --no-trunc --format '{{.Names}}\t{{.State}}\t{{.Mounts}}'` gives which containers reference which volumes. Volumes referenced by a running container are reported as in use and skipped; stopped containers referencing a volume to be removed are removed first.
6. For missing projects additionally remove `~/.asylum/projects/<cname>` (via the existing `removeTree`) and the port allocation.

Volumes whose project is resolved through the old container name format (`asylum-<hash>` without basename) are covered by the registry lookup because the registry stores whichever container name was used at allocation time.

Alternative: treat every unlabeled volume whose name prefix does not match a running project as orphan. Rejected: asylum cannot know all project paths, and a false positive deletes a warm Gradle cache.

### Confirmation and output
Same interaction pattern as `cleanup --all`: list grouped by project with reason (`project missing`, `node_modules gone`), separate sections for `in use` and `unknown project`, then `Proceed? (y/N)`. Requires a terminal.

## Risks / Trade-offs

- [A project on an unmounted external drive looks deleted] → It is listed with its full path and the reason `project missing` before confirmation; the user decides. Automatic pruning on start never touches other projects.
- [Unlabeled volumes from projects started before the port registry existed stay "unknown"] → Shown in a separate section; the user can still use `asylum cleanup` in that project or `cleanup --all`.
- [`FindNodeModulesDirs` walks every existing project during `--orphans`] → Same walk already happens on every start; acceptable for a manual command.
- [`{{.Mounts}}` format output differs between Docker and Podman] → Fall back to "volume in use" errors from `docker volume rm`, which are reported per volume and do not abort the rest.

## Migration Plan

No migration. New volumes get labels; existing ones are classified via the registry. Rollback is a binary downgrade; labels are ignored by older versions.
