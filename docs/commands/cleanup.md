# cleanup

Remove Asylum containers, volumes, and cached data.

## Usage

```
asylum cleanup          # current project only
asylum cleanup --all    # everything
asylum cleanup --orphans  # leftovers of deleted projects and node_modules
asylum --cleanup        # flag alias (current project)
```

## Description

By default, `cleanup` scopes to the **current project**: it removes the project's container, its Docker volumes, and its project data directory. Other projects and the base image are untouched.

Use `--all` for a global cleanup that removes all Asylum images, volumes, and optionally host-side cached data.

Agent configuration (`~/.asylum/agents/`) is always preserved since it contains auth tokens and session data.

## Project Cleanup (default)

Running `asylum cleanup` from a project directory removes:

| Resource | Removed |
|----------|---------|
| Project container | Yes |
| Project volumes (`<container>-*`) | Yes |
| Project data (`~/.asylum/projects/<container>/`: history, credentials, Claude temp files) | Yes |
| Port allocation for project | Yes |
| Base image | No |
| Other projects | No |

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

For a project that no longer exists, its stopped container, project data directory and port allocation are removed as well. Volumes used by a running container are skipped. A stopped container that still references an orphaned volume is removed, since Asylum recreates it on the next start anyway.

A project on an unmounted external drive looks deleted. Check the list before confirming.

Orphan cleanup requires an interactive terminal and cannot be combined with `--all`.

Stale `node_modules` volumes of the current project are also removed automatically whenever a fresh container starts.

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
