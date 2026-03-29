# OpenSpec Kit

[OpenSpec](https://openspec.dev) CLI for structured change management.

**Activation: Default** — added to config on first detection; active when present. Depends on the [Node.js](node.md) kit.

## What's Included

- **openspec** — CLI for managing specs, changes, and proposals

## Configuration

```yaml
kits:
  openspec:
    disabled: true    # disable this default-on kit
```

## Dependencies

The OpenSpec kit depends on `node` because it's installed via npm. If the Node.js kit is not active, Asylum emits a warning.

## Usage

OpenSpec is used inside containers for structured change management:

```sh
openspec new change "add-feature"
openspec status --change "add-feature"
openspec list
```

See the [OpenSpec documentation](https://openspec.dev) for full usage details.

## Notes

- OpenSpec telemetry is disabled inside containers (`OPENSPEC_TELEMETRY=0`).
