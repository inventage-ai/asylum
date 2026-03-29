# cleanup

Remove Asylum Docker images and cached data.

## Usage

```
asylum cleanup
asylum --cleanup    # flag alias
```

## Description

Removes all Asylum Docker images (base and project images) and named Docker volumes (shadow `node_modules` and package caches). Optionally removes host-side cached data.

Agent configuration (`~/.asylum/agents/`) is always preserved since it contains auth tokens and session data.

## What Gets Removed

| Resource | Removed |
|----------|---------|
| Base image (`asylum:latest`) | Yes |
| Project images (`asylum:proj-*`) | Yes |
| Named Docker volumes (`asylum-*`) | Yes |
| Host cache (`~/.asylum/cache/`) | Only if confirmed |
| Host project data (`~/.asylum/projects/`) | Only if confirmed (skips active sessions) |
| Agent config (`~/.asylum/agents/`) | Never |

## Interactive Prompt

After removing images and volumes, Asylum asks whether to also remove host-side cached data:

```
Remove cached data (~/.asylum/cache/ and ~/.asylum/projects/)? (y/N)
```

If running in a non-interactive terminal (e.g., a script), the prompt is skipped and host caches are preserved.
