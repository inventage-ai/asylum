## Context

Container names are deterministic: `asylum-<sha256(project_dir)[:12]>`. When a container is already running (agent session), a second `asylum shell` or `asylum run` fails because `docker run --name <same-name>` errors on the name conflict.

The fix is to detect the running container and use `docker exec` instead of `docker run` for shell/run modes.

## Goals / Non-Goals

**Goals:**
- `asylum shell` and `asylum run <cmd>` exec into a running container seamlessly
- `asylum shell --admin` works the same way (exec as root? or same user?)
- No behavior change when no container is running (still does `docker run`)

**Non-Goals:**
- Exec-ing a second agent into a running container (ambiguous UX, skip for now)
- Attaching to the running agent's session

## Decisions

### Use `docker exec -it` for shell/run when container is running

`docker exec -it <container-name> <command>` gives the user a shell or runs a command in the existing container. This inherits the container's environment, mounts, and network — exactly what you'd want.

### Detect running container via `docker inspect`

`docker inspect --format '{{.State.Running}}' <name>` returns `true`/`false` and is cheaper than `docker ps --filter`. Add `IsRunning(name string) bool` to the docker package.

### Skip image build when exec-ing

When we're going to exec into a running container, there's no need to run `EnsureBase` or `EnsureProject` — the container already has its image. This also avoids the host-side mutations from `RunArgs` (cache dir creation, agent config seeding).

### Agent mode still uses `docker run`

If the user runs `asylum` (agent mode) while a container is already running, we let it fail with the name conflict. Starting a second agent in the same container would be confusing. The error message from Docker is clear enough.

### `ExecArgs` in the container package

A new `ExecArgs(containerName string, mode Mode, extraArgs []string)` function builds the `docker exec` arg list. It's much simpler than `RunArgs` — no volumes, env, ports, or image tag needed.

For admin shell, exec as root with `-u root`.

## Risks / Trade-offs

- **Environment differences**: `docker exec` inherits the container's environment but not any new env vars from the host config. This is acceptable — the container was started with the right env, and shell/run just needs access to it.
- **No port forwarding**: If the user specified `-p` flags, those only apply to `docker run`. An exec can't add ports. This is fine — ports are set when the container starts.
