## Why

Running `asylum -a pi` in a project that already has a Claude container fails: asylum execs into the existing container, which has no `pi` installed. Today there is one container per project (`asylum-<hash(project)>-<name>`), and switching the requested agent does not change that name.

We want to use a different agent ad-hoc (for a review, a quick test) without first wiring it into the project/user config — and without destroying the running container that another agent is using.

## What Changes

- New containers are labeled with the set of agents baked into their image (`asylum.agents=<sorted,comma-separated>`).
- On startup, after locating the project's container, asylum checks whether the requested agent is in that label.
- If it is, the container is reused as today.
- If it is not, asylum derives a **secondary** container name from `hash(project_dir + sorted_agents)`, repeats the running-container lookup against that name, and starts a new container there if none exists. The original container is left untouched — **two containers can run for one project at once**.
- The primary container (`hash(project_dir)`, byte-identical to today) keeps its port allocation. **Secondary containers forward no ports** — they exist for ad-hoc/review use until the agent is promoted into config.
- The shipped "agent missing → `docker rm` and rebuild the same name" path is removed; it destroyed the running agent's container and is the wrong behavior for this design.

Promoting an agent into project/user config bakes it into the **primary** image (its label then lists all configured agents), so a properly-configured multi-agent project never spills and every agent gets full ports.

## Capabilities

### New Capabilities
- `agent-container-resolution`: Resolve a per-project container by the requested agent set — reuse on a label match, otherwise spill to a portless `hash(project+agents)` secondary container without disturbing the primary.

### Modified Capabilities
- `container-assembly`: Container naming accepts the agent set; the primary name is unchanged, secondaries hash project + agents. Secondary containers omit port-forwarding args.

## Impact

- `internal/container/container.go`: `ContainerName(projectDir, agents)`; emit `asylum.agents` label (already prototyped); flag primary vs secondary so `RunArgs` can skip ports for secondaries.
- `internal/docker/docker.go`: `InspectLabels` / `ContainerHasAgent` helpers (already prototyped).
- `cmd/asylum/main.go`: two-pass name resolution; remove the `RemoveContainer`-and-rebuild branch.
- No changes to `internal/ports`: secondaries simply never call `Allocate`.
- **Out of scope (separate change):** container ephemerality and `asylum cleanup` enumerating every secondary container for a project. The existing session-counter teardown already removes a secondary when its last session exits.
