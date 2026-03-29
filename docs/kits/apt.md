# APT Packages Kit

Install extra system packages via apt in the project image.

## What's Included

Nothing by default — this kit exists solely to install additional Debian packages you specify.

## Configuration

```yaml
kits:
  apt:
    packages:
      - libpq-dev
      - redis-tools
      - ffmpeg
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `packages` | list | `[]` | Debian packages to install via `apt-get install` |

## How It Works

Packages are installed as root during the project image build using `apt-get install -y --no-install-recommends`. The project image is cached and only rebuilt when the package list changes.

## When to Use

Use this kit for system libraries and tools that aren't covered by other kits. Common examples:

- Database client libraries: `libpq-dev`, `libmysqlclient-dev`
- Media processing: `ffmpeg`, `imagemagick`
- Additional CLI tools: `redis-tools`, `postgresql-client`
