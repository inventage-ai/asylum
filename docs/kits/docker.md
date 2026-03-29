# Docker Kit

Full Docker-in-Docker support with buildx and compose.

**Activation: Default** — added to config on first detection; active when present.

## What's Included

- **Docker CE** (full engine, not just CLI)
- **docker-buildx-plugin**
- **docker-compose-plugin**

## Configuration

```yaml
kits:
  docker: {}
```

The Docker kit has no additional configuration options. Enable it with an empty map.

## How It Works

When the Docker kit is active:

1. The container runs in **privileged mode** (required for Docker-in-Docker)
2. The Docker daemon (`dockerd`) starts automatically inside the container
3. The `claude` user is added to the `docker` group

The daemon uses the `vfs` storage driver and waits up to 30 seconds to be ready before the agent or shell starts.

!!! note
    Privileged mode gives the container full access to the host kernel. This is required for Docker-in-Docker but reduces isolation. Only enable this kit if you need to build or run Docker containers inside Asylum.
