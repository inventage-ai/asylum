# cleanup

Remove Asylum containers, volumes, and cached data.

## Usage

```
asylum cleanup          # current project only
asylum cleanup --all    # everything
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
| Project data (`~/.asylum/projects/<container>/`) | Yes |
| Port allocation for project | Yes |
| Base image | No |
| Other projects | No |

If run outside a project directory, Asylum suggests using `--all` instead.

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
