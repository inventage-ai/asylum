## Context

Asylum uses a file-based session counter (`~/.asylum/projects/<container>/sessions`) to track how many concurrent CLI sessions are attached to a container. When the counter reaches zero, the container is removed. This counter corrupts on any unclean exit (SIGHUP, SIGKILL, crash), permanently preventing container cleanup until `asylum cleanup` is run.

The container runs `sleep infinity` as PID 1 and asylum sessions are injected via `docker exec`. Inside the container's PID namespace, all `docker exec` processes have PPID=0, while container-internal processes (background tasks, daemons) have PPID=1. This distinction is a Linux kernel guarantee, not a Docker convention.

## Goals / Non-Goals

**Goals:**
- Eliminate ghost session bugs by removing persistent session state
- Self-heal from any unclean exit without manual intervention
- Forward SIGHUP for clean signal handling on terminal close

**Non-Goals:**
- Changing the detached container + exec architecture
- Adding Docker SDK dependency (continue shelling out to `docker` CLI)

## Decisions

### Runtime session detection via `ps`

At cleanup time, run `docker exec <container> ps -o pid,ppid --no-headers` and count processes with PPID=0 (excluding PID 1 and the `ps` check itself). If count > 0, other sessions are active.

**Why not Docker API (`/exec/<id>/json`)**: Requires Unix socket access and adds API coupling. The `ps` approach uses tools already in the container and works through the standard `docker exec` path.

**Why not `docker top`**: Shows host PIDs with host PPIDs. The PPID=0 signal is only clean inside the container's PID namespace.

### Place `HasOtherSessions` in `internal/docker/`

It shells out to `docker exec`, which is a Docker operation. Consistent with the existing `docker.IsRunning`, `docker.RemoveContainer` pattern.

### Add SIGHUP to signal forwarding

Even without the counter, SIGHUP should be forwarded to docker exec so the agent process inside the container gets a clean shutdown signal rather than having its exec session abruptly severed.

## Risks / Trade-offs

**[`ps` not available in container]** → Mitigated: `procps` is installed in the base Dockerfile. Would only be an issue if someone manually stripped it.

**[Brief race window at cleanup]** → A new session could start between our session exiting and the `ps` check running. This is safe: the check would see the new session and keep the container alive. The reverse (check runs before session starts) would require the new session to start in the milliseconds between check and `docker rm`, which is harmless — docker rm would fail and the new session would proceed normally.

**[`docker exec` overhead for check]** → Negligible. Single `ps` invocation adds ~50ms to session exit. Only runs once per session end.
