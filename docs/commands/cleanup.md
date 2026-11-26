# cleanup

Remove Asylum containers, volumes, and cached data.

## Usage

```
asylum cleanup          # current project only
asylum cleanup --all    # everything
asylum cleanup --orphans  # leftovers of deleted projects and node_modules
asylum cleanup --select   # list all projects, pick which to remove
asylum --cleanup        # flag alias (current project)
```

## Description

By default, `cleanup` scopes to the **current project**: it removes the project's container, its Docker volumes, its project data directory and port allocation, and its project image if no other project uses it. Other projects and the base image are untouched.

Use `--all` for a global cleanup that removes all Asylum images, volumes, and optionally host-side cached data.

Agent configuration (`~/.asylum/agents/`) is always preserved since it contains auth tokens and session data.

## Project Cleanup (default)

Running `asylum cleanup` from a project directory removes:

| Resource | Removed |
|----------|---------|
| Project container (running or stopped) | Yes |
| Project volumes (`<container>-*`) | Yes |
| Project data (`~/.asylum/projects/<container>/`: history, credentials, Claude temp files) | Yes |
| Port allocation for project | Yes |
| Project image (`asylum:proj-*`) | If no other project and no container uses it |
| Base image | No |
| Other projects | No |

Projects with identical project configuration share one project image. Asylum records the image a project runs on at each start, and removes the image only when no other project's record names it. Images of projects last started with an older Asylum version are not recorded and are only removed by `--all`.

If run outside a project directory, Asylum suggests using `--all` instead.

## Orphan Cleanup (`--orphans`)

Volumes outlive the projects they were created for. `asylum cleanup --orphans` finds volumes that no longer belong to anything and removes them after confirmation. A volume is orphaned when:

- its project directory no longer exists (deleted, moved or renamed), or
- it backs a `node_modules` directory whose `package.json` is gone.

The project of a volume is read from its `asylum.project` label. Volumes created by older Asylum versions have no label. For those, Asylum looks the project up in the port registry (`~/.asylum/ports.json`). Volumes whose project cannot be determined are listed but never removed.

```
The following orphaned volumes will be removed:

  /Users/me/code/old-app (project missing)
    asylum-4e4041c8bdea-old-app-cache-gradle
    asylum-4e4041c8bdea-old-app-cache-npm
  /Users/me/code/shop (node_modules gone)
    asylum-d387fb6a1e3b-shop-npm-deadbeef000

Stopped containers holding them will be removed too:
  asylum-4e4041c8bdea-old-app

Project data and port allocations of missing projects will be removed too.

Skipped, project unknown (use 'asylum cleanup' in the project or 'cleanup --all'):
  asylum-999999999999-mystery-cache-pip

Proceed? (y/N)
```

For a project that no longer exists, its stopped container, project data directory and port allocation are removed as well, and its project image if nothing else uses it. Volumes used by a running container are skipped. A stopped container that still references an orphaned volume is removed, since Asylum recreates it on the next start anyway.

A project on an unmounted external drive looks deleted. Check the list before confirming.

Orphan cleanup requires an interactive terminal and cannot be combined with `--all`.

Stale `node_modules` volumes of the current project are also removed automatically whenever a fresh container starts.

## Project Selection (`--select`)

`asylum cleanup --select` lists every project Asylum has resources for, with its status, last use and disk size, and lets you pick which ones to remove:

```
  PROJECT                          STATUS    LAST USED      SIZE
  /Users/me/code/gone              missing   8 months ago   900.0 MB
  /Users/me/code/old-poc           idle      5 months ago   1.8 GB
  /Users/me/code/shop              idle      3 days ago     2.3 GB
  /Users/me/code/asylum            running   now            4.1 GB
  asylum-1a2b3c4d5e6f-mystery      idle      unknown        120.0 MB
```

| Status | Meaning |
|--------|---------|
| `running` | The project's container is running. Listed, but cannot be selected. |
| `missing` | The project directory no longer exists. |
| `stopped` | A stopped container is left over, e.g. after a crash. |
| `idle` | No container. The normal state between sessions. |

**Last used** is the start of the most recent session. Asylum records it at each start, so projects last started with an older version show the modification time of their project data directory instead, which is often older than the real last use.

**Size** adds up the project's volumes, its project data directory and its project image, unless another project uses the same image. Shared images are not counted for any project. Collecting sizes calls `docker system df -v`, which can take a few seconds.

Projects whose path is unknown are shown by their container name.

After the table, a multi-select lists all projects that are not running. Missing projects and projects unused for more than 30 days are preselected. After a summary with the total size and a confirmation, each selected project is removed the same way as `asylum cleanup` in that project.

A project on an unmounted external drive shows as `missing` and is preselected. Deselect it if you want to keep it.

Project selection requires an interactive terminal and cannot be combined with `--all` or `--orphans`.

## Global Cleanup (`--all`)

Running `asylum cleanup --all` shows all resources that will be removed and asks for confirmation before proceeding:

```
The following resources will be removed:

  Images:
    asylum:latest
    asylum:proj-abc123def456

  Volumes:
    asylum-7a3f2b-myapp-npm
    asylum-7a3f2b-myapp-pip

Proceed? (y/N)
```

After removing images and volumes, a second prompt offers to remove host-side cached data:

```
Remove cached data (~/.asylum/cache/ and ~/.asylum/projects/)? (y/N)
```

Global cleanup requires an interactive terminal — it won't run in scripts or CI.
