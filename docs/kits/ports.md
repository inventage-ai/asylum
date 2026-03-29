# Ports Kit

Automatic port forwarding for web services.

**Activation: Always on** — active in every container.

## What It Does

The ports kit automatically allocates and forwards a range of high ports for each project. This means web servers, dev tools, and other services started inside the container are accessible from your host without manually specifying `-p` flags.

## Configuration

```yaml
kits:
  ports:
    count: 5          # number of ports to allocate (default: 5)
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `count` | int | `5` | Number of ports to allocate per project |

## How It Works

Each project gets a unique range of consecutive ports starting from port 10000. The allocation is tracked globally in `~/.asylum/state.json` to prevent collisions between projects.

For example, if your project gets ports 10000-10004:

- Start a server on port 10000 inside the container
- Access it at `localhost:10000` on your host

The allocation is deterministic — the same project always gets the same port range (unless it's been freed and reallocated).

## Additional Ports

The ports kit handles automatic allocation. You can still forward specific ports manually with `-p`:

```sh
asylum -p 3000 -p 8080:80
```

Manual port forwards are in addition to the automatically allocated range.
